package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

const (
	tokenRefreshSkew      = 30 * time.Second
	defaultServerEndpoint = "http://127.0.0.1:8080"
	defaultHTTPTimeout    = 15 * time.Second
	defaultClientMode     = "release"
)

var (
	serverAddr   string
	authToken    string
	authUser     string
	authPassword string
	clientMode   string
	configPath   string
	httpTimeout  time.Duration
)

type clientConfig struct {
	ServerAddr     string             `json:"server_addr,omitempty"`
	Mode           string             `json:"mode,omitempty"`
	Token          string             `json:"token,omitempty"`
	TokenExpiresAt string             `json:"token_expires_at,omitempty"`
	Username       string             `json:"username,omitempty"`
	Password       string             `json:"password,omitempty"`
	TimeoutSeconds int                `json:"timeout_seconds,omitempty"`
	VersionCache   *versionCacheEntry `json:"version_cache,omitempty"`
}

var rootCmd = &cobra.Command{
	Use:   "mkp",
	Short: "MKP API client",
}

func Execute() error {
	return rootCmd.Execute()
}

func SetVersion(version string) {
	rootCmd.Version = strings.TrimSpace(version)
}

func init() {
	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		return hydrateRuntimeConfig()
	}

	rootCmd.PersistentFlags().StringVarP(&serverAddr, "server", "s", "", "MKP server base URL (or MKP_SERVER)")
	rootCmd.PersistentFlags().StringVar(&clientMode, "mode", "", "Client mode: debug/release (or MKP_MODE)")
	rootCmd.PersistentFlags().StringVar(&authToken, "token", "", "Bearer token (or MKP_TOKEN)")
	rootCmd.PersistentFlags().StringVar(&authUser, "username", "", "Username used to fetch token (or MKP_USER)")
	rootCmd.PersistentFlags().StringVar(&authPassword, "password", "", "Password used to fetch token (or MKP_PASSWORD)")
	rootCmd.PersistentFlags().StringVar(&configPath, "config", os.Getenv("MKP_CONFIG_PATH"), "Client config file path (default: ~/.mkp/client.json)")
	rootCmd.PersistentFlags().DurationVar(&httpTimeout, "timeout", 0, "HTTP timeout (or MKP_TIMEOUT)")
}

func hydrateRuntimeConfig() error {
	cfg, err := loadClientConfig()
	if err != nil {
		return err
	}

	serverAddr = firstNonEmpty(
		flagValue("server", serverAddr),
		os.Getenv("MKP_SERVER"),
		cfg.ServerAddr,
		defaultServerEndpoint,
	)

	clientMode = firstNonEmpty(
		flagValue("mode", clientMode),
		os.Getenv("MKP_MODE"),
		cfg.Mode,
		defaultClientMode,
	)

	authToken = firstNonEmpty(
		flagValue("token", authToken),
		os.Getenv("MKP_TOKEN"),
		cfg.Token,
	)

	authUser = firstNonEmpty(
		flagValue("username", authUser),
		os.Getenv("MKP_USER"),
		cfg.Username,
	)

	authPassword = firstNonEmpty(
		flagValue("password", authPassword),
		os.Getenv("MKP_PASSWORD"),
		cfg.Password,
	)

	if !flagChanged("timeout") || httpTimeout <= 0 {
		if envTimeout := strings.TrimSpace(os.Getenv("MKP_TIMEOUT")); envTimeout != "" {
			if parsed, err := time.ParseDuration(envTimeout); err == nil && parsed > 0 {
				httpTimeout = parsed
			}
		}
	}

	if httpTimeout <= 0 && cfg.TimeoutSeconds > 0 {
		httpTimeout = time.Duration(cfg.TimeoutSeconds) * time.Second
	}
	if httpTimeout <= 0 {
		httpTimeout = defaultHTTPTimeout
	}

	return nil
}

func flagChanged(name string) bool {
	f := rootCmd.PersistentFlags().Lookup(name)
	return f != nil && f.Changed
}

func flagValue(name, value string) string {
	if flagChanged(name) {
		return strings.TrimSpace(value)
	}
	return ""
}

func sendJSON(method, path string, payload any, requireAuth bool) (map[string]any, error) {
	requireAuth = requireAuth && !isDebugMode()

	token := strings.TrimSpace(authToken)

	if requireAuth && token == "" {
		var err error
		token, err = ensureManagedToken(false)
		if err != nil {
			return nil, err
		}
	}

	resp, status, err := doJSONRequest(method, path, payload, token)
	if err != nil {
		return nil, err
	}

	if status == http.StatusUnauthorized && requireAuth && token != "" && strings.TrimSpace(flagValue("token", authToken)) == "" {
		token, err = ensureManagedToken(true)
		if err != nil {
			return nil, err
		}
		resp, status, err = doJSONRequest(method, path, payload, token)
		if err != nil {
			return nil, err
		}
	}

	if status >= http.StatusBadRequest {
		return nil, formatAPIError(status, resp)
	}
	return resp, nil
}

func doJSONRequest(method, path string, payload any, bearerToken string) (map[string]any, int, error) {
	base := strings.TrimRight(strings.TrimSpace(serverAddr), "/")
	if base == "" {
		return nil, 0, fmt.Errorf("server address is required")
	}

	var body io.Reader
	if payload != nil {
		data, err := json.Marshal(payload)
		if err != nil {
			return nil, 0, err
		}
		body = bytes.NewBuffer(data)
	}

	req, err := http.NewRequest(method, base+path, body)
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	if token := strings.TrimSpace(bearerToken); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	client := &http.Client{Timeout: httpTimeout}
	httpResp, err := client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer httpResp.Body.Close()

	respBytes, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, 0, err
	}

	result := map[string]any{}
	if len(respBytes) > 0 {
		if err := json.Unmarshal(respBytes, &result); err != nil {
			return nil, 0, fmt.Errorf("invalid response: %s", string(respBytes))
		}
	}

	return result, httpResp.StatusCode, nil
}

func formatAPIError(status int, payload map[string]any) error {
	if msg, ok := payload["error"].(string); ok && msg != "" {
		return fmt.Errorf("request failed (%d): %s", status, msg)
	}
	return fmt.Errorf("request failed with status %d", status)
}

func ensureManagedToken(forceRefresh bool) (string, error) {
	cfg, err := loadClientConfig()
	if err != nil {
		return "", err
	}

	if !flagChanged("server") && strings.TrimSpace(cfg.ServerAddr) != "" {
		serverAddr = cfg.ServerAddr
	}

	cfg.Username = firstNonEmpty(authUser, cfg.Username)
	cfg.Password = firstNonEmpty(authPassword, cfg.Password)

	if !forceRefresh && strings.TrimSpace(cfg.Token) != "" && tokenStillValid(cfg.TokenExpiresAt) {
		return cfg.Token, nil
	}

	if cfg.Username == "" || cfg.Password == "" {
		return "", fmt.Errorf("missing auth credentials: run `mkp auth login` or pass --username/--password")
	}

	token, expiresAt, err := fetchToken(cfg.Username, cfg.Password)
	if err != nil {
		return "", err
	}

	cfg.ServerAddr = strings.TrimSpace(serverAddr)
	cfg.Mode = strings.TrimSpace(clientMode)
	cfg.Token = token
	cfg.TokenExpiresAt = expiresAt.Format(time.RFC3339)
	if httpTimeout > 0 {
		cfg.TimeoutSeconds = int(httpTimeout.Seconds())
	}

	if err := saveClientConfig(cfg); err != nil {
		return "", err
	}

	return token, nil
}

func fetchToken(username, password string) (string, time.Time, error) {
	req := map[string]string{
		"username": username,
		"password": password,
	}

	resp, status, err := doJSONRequest(http.MethodPost, "/api/v1/auth/token", req, "")
	if err != nil {
		return "", time.Time{}, err
	}
	if status >= http.StatusBadRequest {
		return "", time.Time{}, formatAPIError(status, resp)
	}

	token, _ := resp["token"].(string)
	if strings.TrimSpace(token) == "" {
		return "", time.Time{}, fmt.Errorf("auth response missing token")
	}

	expiresIn := 0
	switch v := resp["expires_in"].(type) {
	case float64:
		expiresIn = int(v)
	case int:
		expiresIn = v
	}

	expiresAt := time.Now().Add(24 * time.Hour)
	if expiresIn > 0 {
		expiresAt = time.Now().Add(time.Duration(expiresIn) * time.Second)
	}

	return token, expiresAt, nil
}

func tokenStillValid(expiresAt string) bool {
	parsed, err := time.Parse(time.RFC3339, strings.TrimSpace(expiresAt))
	if err != nil || parsed.IsZero() {
		return false
	}
	return time.Now().Add(tokenRefreshSkew).Before(parsed)
}

func resolveConfigPath() (string, error) {
	if p := strings.TrimSpace(configPath); p != "" {
		return p, nil
	}

	if wd, err := os.Getwd(); err == nil {
		localPath := filepath.Join(wd, "client.json")
		if _, statErr := os.Stat(localPath); statErr == nil {
			return localPath, nil
		}
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".mkp", "client.json"), nil
}

func loadClientConfig() (*clientConfig, error) {
	path, err := resolveConfigPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &clientConfig{}, nil
		}
		return nil, err
	}

	cfg := &clientConfig{}
	if len(data) == 0 {
		return cfg, nil
	}
	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func saveClientConfig(cfg *clientConfig) error {
	path, err := resolveConfigPath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func isDebugMode() bool {
	return strings.EqualFold(strings.TrimSpace(clientMode), "debug")
}
