package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	mysqlDriver "github.com/go-sql-driver/mysql"
)

type AuthConfig struct {
	JWTSecret   string
	User        string
	Password    string
	Issuer      string
	TokenTTL    time.Duration
	MuteInDebug bool
}

type SQLiteDSN struct {
	WithOTEL bool   `json:"withOTEL,omitempty"`
	DSN      string `json:"dsn"`

	MaxIdleConns             int    `json:"maxIdleConns,omitempty"`
	MaxOpenConns             int    `json:"maxOpenConns,omitempty"`
	ConnMaxIdleTimeInSeconds *int64 `json:"connMaxIdleTimeInSeconds,omitempty"`
	ConnMaxLifetimeInSeconds *int64 `json:"connMaxLifetimeInSeconds,omitempty"`
}

type MySQLDSN struct {
	WithOTEL bool `json:"withOTEL,omitempty"`

	MaxIdleConns             int    `json:"maxIdleConns,omitempty"`
	MaxOpenConns             int    `json:"maxOpenConns,omitempty"`
	ConnMaxIdleTimeInSeconds *int64 `json:"connMaxIdleTimeInSeconds,omitempty"`
	ConnMaxLifetimeInSeconds *int64 `json:"connMaxLifetimeInSeconds,omitempty"`

	Username    string            `json:"username,omitempty"`
	Password    string            `json:"password,omitempty"`
	NetworkType string            `json:"networkType,omitempty"`
	Address     string            `json:"address,omitempty"`
	Database    string            `json:"database,omitempty"`
	Params      map[string]string `json:"params,omitempty"`

	Collation        string         `json:"collation,omitempty"`
	Loc              *time.Location `json:"loc,omitempty"`
	MaxAllowedPacket int            `json:"maxAllowedPacket,omitempty"`
	ServerPubKey     string         `json:"serverPubKey,omitempty"`
	TLSConfig        string         `json:"tlsConfig,omitempty"`
	Timeout          time.Duration  `json:"timeout,omitempty"`
	ReadTimeout      time.Duration  `json:"readTimeout,omitempty"`
	WriteTimeout     time.Duration  `json:"writeTimeout,omitempty"`

	AllowAllFiles           bool `json:"allowAllFiles,omitempty"`
	AllowCleartextPasswords bool `json:"allowCleartextPasswords,omitempty"`
	AllowNativePasswords    bool `json:"allowNativePasswords,omitempty"`
	AllowOldPasswords       bool `json:"allowOldPasswords,omitempty"`
	CheckConnLiveness       bool `json:"checkConnLiveness,omitempty"`
	ClientFoundRows         bool `json:"clientFoundRows,omitempty"`
	ColumnsWithAlias        bool `json:"columnsWithAlias,omitempty"`
	InterpolateParams       bool `json:"interpolateParams,omitempty"`
	MultiStatements         bool `json:"multiStatements,omitempty"`
	ParseTime               bool `json:"parseTime,omitempty"`
	RejectReadOnly          bool `json:"rejectReadOnly,omitempty"`
}

type PostgresDSN struct {
	WithOTEL bool   `json:"withOTEL,omitempty"`
	DSN      string `json:"dsn,omitempty"`

	MaxIdleConns             int    `json:"maxIdleConns,omitempty"`
	MaxOpenConns             int    `json:"maxOpenConns,omitempty"`
	ConnMaxIdleTimeInSeconds *int64 `json:"connMaxIdleTimeInSeconds,omitempty"`
	ConnMaxLifetimeInSeconds *int64 `json:"connMaxLifetimeInSeconds,omitempty"`

	Host     string            `json:"host,omitempty"`
	Port     int               `json:"port,omitempty"`
	Username string            `json:"username,omitempty"`
	Password string            `json:"password,omitempty"`
	Database string            `json:"database,omitempty"`
	SSLMode  string            `json:"sslMode,omitempty"`
	TimeZone string            `json:"timeZone,omitempty"`
	Params   map[string]string `json:"params,omitempty"`
}

type FileAuthConfig struct {
	Secret      string `json:"secret"`
	User        string `json:"user"`
	Password    string `json:"password"`
	Issuer      string `json:"issuer"`
	TTLSeconds  int    `json:"ttl_seconds"`
	MuteInDebug bool   `json:"muteInDebug,omitempty"`
}

type FileDatabaseConfig struct {
	Driver string `json:"driver"`
	DSN    string `json:"dsn,omitempty"`

	SQLite   SQLiteDSN   `json:"sqlite,omitempty"`
	MySQL    MySQLDSN    `json:"mysql,omitempty"`
	Postgres PostgresDSN `json:"postgres,omitempty"`
}

type FileConfig struct {
	Mode     string             `json:"mode,omitempty"`
	Port     string             `json:"port"`
	SFPort   string             `json:"sf_port"`
	JWT      FileAuthConfig     `json:"jwt"`
	Database FileDatabaseConfig `json:"database"`
}

type DatabaseConfig struct {
	Driver string
	DSN    string

	WithOTEL bool

	MaxIdleConns             int
	MaxOpenConns             int
	ConnMaxIdleTimeInSeconds *int64
	ConnMaxLifetimeInSeconds *int64
}

type AppConfig struct {
	Mode     string
	Port     string
	SFPort   string
	Auth     AuthConfig
	Database DatabaseConfig
}

const (
	// SQLite defaults favor safe single-writer behavior.
	defaultSQLiteMaxIdleConns   = 1
	defaultSQLiteMaxOpenConns   = 1
	defaultSQLiteConnMaxIdleS   = int64(300)  // 5m
	defaultSQLiteConnMaxLifeS   = int64(1800) // 30m
	defaultMySQLMaxIdleConns    = 10
	defaultMySQLMaxOpenConns    = 50
	defaultMySQLConnMaxIdleS    = int64(300)  // 5m
	defaultMySQLConnMaxLifeS    = int64(1800) // 30m
	defaultPostgresMaxIdleConns = 10
	defaultPostgresMaxOpenConns = 50
	defaultPostgresConnMaxIdleS = int64(300)  // 5m
	defaultPostgresConnMaxLifeS = int64(1800) // 30m
)

func Load(cliPath string) (AppConfig, string, error) {
	fileCfg, loadedPath, err := loadFileConfig(cliPath)
	if err != nil {
		return AppConfig{}, "", err
	}

	cfg := AppConfig{
		Mode:     resolveRunMode(fileCfg.Mode),
		Port:     firstNonEmpty(os.Getenv("PORT"), fileCfg.Port, "8080"),
		SFPort:   firstNonEmpty(os.Getenv("SF_PORT"), fileCfg.SFPort),
		Auth:     composeAuthConfig(fileCfg),
		Database: composeDatabaseConfig(fileCfg),
	}

	return cfg, loadedPath, nil
}

func resolveRunMode(fileMode string) string {
	// dev/development/debug -> debug, test/testing -> test, prod/production/release -> release
	mode := strings.ToLower(strings.TrimSpace(firstNonEmpty(
		os.Getenv("APP_MODE"),
		os.Getenv("RUN_MODE"),
		os.Getenv("GIN_MODE"),
		fileMode,
		"release",
	)))

	switch mode {
	case "dev", "development", "debug":
		return "debug"
	case "test", "testing":
		return "test"
	case "prod", "production", "release":
		return "release"
	default:
		return "release"
	}
}

func composeAuthConfig(fileCfg FileConfig) AuthConfig {
	ttlSeconds := 86400
	if fileCfg.JWT.TTLSeconds > 0 {
		ttlSeconds = fileCfg.JWT.TTLSeconds
	}
	if raw := strings.TrimSpace(os.Getenv("JWT_TTL_SECONDS")); raw != "" {
		ttlSeconds = parsePositiveInt(raw, ttlSeconds)
	}

	return AuthConfig{
		JWTSecret:   firstNonEmpty(os.Getenv("JWT_SECRET"), fileCfg.JWT.Secret, "mkp-go-dev-secret"),
		User:        firstNonEmpty(os.Getenv("JWT_USER"), fileCfg.JWT.User, "admin"),
		Password:    firstNonEmpty(os.Getenv("JWT_PASSWORD"), fileCfg.JWT.Password, "admin123"),
		Issuer:      firstNonEmpty(os.Getenv("JWT_ISSUER"), fileCfg.JWT.Issuer, "mkp-go-server"),
		TokenTTL:    time.Duration(ttlSeconds) * time.Second,
		MuteInDebug: parseBoolWithDefault(os.Getenv("MUTE_JWT_AUTH"), fileCfg.JWT.MuteInDebug),
	}
}

func composeDatabaseConfig(fileCfg FileConfig) DatabaseConfig {
	driver := normalizeDBDriver(firstNonEmpty(os.Getenv("DB_DRIVER"), fileCfg.Database.Driver, "sqlite3"))
	envDSN := strings.TrimSpace(os.Getenv("DB_DSN"))
	legacyDSN := strings.TrimSpace(fileCfg.Database.DSN)
	cfg := DatabaseConfig{Driver: driver}

	switch driver {
	case "sqlite3":
		s := fileCfg.Database.SQLite
		cfg.WithOTEL = s.WithOTEL
		cfg.MaxIdleConns = s.MaxIdleConns
		cfg.MaxOpenConns = s.MaxOpenConns
		cfg.ConnMaxIdleTimeInSeconds = s.ConnMaxIdleTimeInSeconds
		cfg.ConnMaxLifetimeInSeconds = s.ConnMaxLifetimeInSeconds
		cfg.DSN = firstNonEmpty(envDSN, s.DSN, legacyDSN, "mkp.db")

	case "mysql":
		m := fileCfg.Database.MySQL
		cfg.WithOTEL = m.WithOTEL
		cfg.MaxIdleConns = m.MaxIdleConns
		cfg.MaxOpenConns = m.MaxOpenConns
		cfg.ConnMaxIdleTimeInSeconds = m.ConnMaxIdleTimeInSeconds
		cfg.ConnMaxLifetimeInSeconds = m.ConnMaxLifetimeInSeconds

		built := ""
		if envDSN == "" && legacyDSN == "" {
			built = buildMySQLDSN(m)
		}
		cfg.DSN = firstNonEmpty(envDSN, legacyDSN, built)

	case "postgres":
		p := fileCfg.Database.Postgres
		cfg.WithOTEL = p.WithOTEL
		cfg.MaxIdleConns = p.MaxIdleConns
		cfg.MaxOpenConns = p.MaxOpenConns
		cfg.ConnMaxIdleTimeInSeconds = p.ConnMaxIdleTimeInSeconds
		cfg.ConnMaxLifetimeInSeconds = p.ConnMaxLifetimeInSeconds

		built := ""
		if envDSN == "" && legacyDSN == "" {
			built = buildPostgresDSN(p)
		}
		cfg.DSN = firstNonEmpty(envDSN, p.DSN, legacyDSN, built)

	default:
		cfg.DSN = firstNonEmpty(envDSN, legacyDSN)
	}

	applyRecommendedPoolDefaults(&cfg)
	return cfg
}

func applyRecommendedPoolDefaults(cfg *DatabaseConfig) {
	switch normalizeDBDriver(cfg.Driver) {
	case "sqlite3":
		if cfg.MaxIdleConns <= 0 {
			cfg.MaxIdleConns = defaultSQLiteMaxIdleConns
		}
		if cfg.MaxOpenConns <= 0 {
			cfg.MaxOpenConns = defaultSQLiteMaxOpenConns
		}
		if cfg.ConnMaxIdleTimeInSeconds == nil || *cfg.ConnMaxIdleTimeInSeconds <= 0 {
			cfg.ConnMaxIdleTimeInSeconds = int64Ptr(defaultSQLiteConnMaxIdleS)
		}
		if cfg.ConnMaxLifetimeInSeconds == nil || *cfg.ConnMaxLifetimeInSeconds <= 0 {
			cfg.ConnMaxLifetimeInSeconds = int64Ptr(defaultSQLiteConnMaxLifeS)
		}

	case "mysql":
		if cfg.MaxIdleConns <= 0 {
			cfg.MaxIdleConns = defaultMySQLMaxIdleConns
		}
		if cfg.MaxOpenConns <= 0 {
			cfg.MaxOpenConns = defaultMySQLMaxOpenConns
		}
		if cfg.ConnMaxIdleTimeInSeconds == nil || *cfg.ConnMaxIdleTimeInSeconds <= 0 {
			cfg.ConnMaxIdleTimeInSeconds = int64Ptr(defaultMySQLConnMaxIdleS)
		}
		if cfg.ConnMaxLifetimeInSeconds == nil || *cfg.ConnMaxLifetimeInSeconds <= 0 {
			cfg.ConnMaxLifetimeInSeconds = int64Ptr(defaultMySQLConnMaxLifeS)
		}

	case "postgres":
		if cfg.MaxIdleConns <= 0 {
			cfg.MaxIdleConns = defaultPostgresMaxIdleConns
		}
		if cfg.MaxOpenConns <= 0 {
			cfg.MaxOpenConns = defaultPostgresMaxOpenConns
		}
		if cfg.ConnMaxIdleTimeInSeconds == nil || *cfg.ConnMaxIdleTimeInSeconds <= 0 {
			cfg.ConnMaxIdleTimeInSeconds = int64Ptr(defaultPostgresConnMaxIdleS)
		}
		if cfg.ConnMaxLifetimeInSeconds == nil || *cfg.ConnMaxLifetimeInSeconds <= 0 {
			cfg.ConnMaxLifetimeInSeconds = int64Ptr(defaultPostgresConnMaxLifeS)
		}
	}
}

func normalizeDBDriver(driver string) string {
	switch strings.ToLower(strings.TrimSpace(driver)) {
	case "", "sqlited", "sqlite", "sqlite3":
		return "sqlite3"
	case "mysql":
		return "mysql"
	case "postgres", "postgresql":
		return "postgres"
	default:
		return strings.ToLower(strings.TrimSpace(driver))
	}
}

func buildMySQLDSN(cfg MySQLDSN) string {
	mysqlCfg := mysqlDriver.NewConfig()
	if cfg.Username != "" {
		mysqlCfg.User = cfg.Username
	}
	if cfg.Password != "" {
		mysqlCfg.Passwd = cfg.Password
	}
	if cfg.NetworkType != "" {
		mysqlCfg.Net = cfg.NetworkType
	}
	if cfg.Address != "" {
		mysqlCfg.Addr = cfg.Address
	}
	if cfg.Database != "" {
		mysqlCfg.DBName = cfg.Database
	}
	if len(cfg.Params) > 0 {
		mysqlCfg.Params = cfg.Params
	}
	if cfg.Collation != "" {
		mysqlCfg.Collation = cfg.Collation
	}
	if cfg.Loc != nil {
		mysqlCfg.Loc = cfg.Loc
	}
	if cfg.MaxAllowedPacket > 0 {
		mysqlCfg.MaxAllowedPacket = cfg.MaxAllowedPacket
	}
	if cfg.ServerPubKey != "" {
		mysqlCfg.ServerPubKey = cfg.ServerPubKey
	}
	if cfg.TLSConfig != "" {
		mysqlCfg.TLSConfig = cfg.TLSConfig
	}
	if cfg.Timeout > 0 {
		mysqlCfg.Timeout = cfg.Timeout
	}
	if cfg.ReadTimeout > 0 {
		mysqlCfg.ReadTimeout = cfg.ReadTimeout
	}
	if cfg.WriteTimeout > 0 {
		mysqlCfg.WriteTimeout = cfg.WriteTimeout
	}

	mysqlCfg.AllowAllFiles = cfg.AllowAllFiles
	mysqlCfg.AllowCleartextPasswords = cfg.AllowCleartextPasswords
	mysqlCfg.AllowNativePasswords = cfg.AllowNativePasswords
	mysqlCfg.AllowOldPasswords = cfg.AllowOldPasswords
	mysqlCfg.CheckConnLiveness = cfg.CheckConnLiveness
	mysqlCfg.ClientFoundRows = cfg.ClientFoundRows
	mysqlCfg.ColumnsWithAlias = cfg.ColumnsWithAlias
	mysqlCfg.InterpolateParams = cfg.InterpolateParams
	mysqlCfg.MultiStatements = cfg.MultiStatements
	mysqlCfg.ParseTime = cfg.ParseTime
	mysqlCfg.RejectReadOnly = cfg.RejectReadOnly

	return mysqlCfg.FormatDSN()
}

func buildPostgresDSN(cfg PostgresDSN) string {
	if strings.TrimSpace(cfg.DSN) != "" {
		return strings.TrimSpace(cfg.DSN)
	}
	parts := make([]string, 0, 8+len(cfg.Params))
	if cfg.Host != "" {
		parts = append(parts, "host="+cfg.Host)
	}
	if cfg.Port > 0 {
		parts = append(parts, "port="+strconv.Itoa(cfg.Port))
	}
	if cfg.Username != "" {
		parts = append(parts, "user="+cfg.Username)
	}
	if cfg.Password != "" {
		parts = append(parts, "password="+cfg.Password)
	}
	if cfg.Database != "" {
		parts = append(parts, "dbname="+cfg.Database)
	}
	if cfg.SSLMode != "" {
		parts = append(parts, "sslmode="+cfg.SSLMode)
	}
	if cfg.TimeZone != "" {
		parts = append(parts, "TimeZone="+cfg.TimeZone)
	}
	if len(cfg.Params) > 0 {
		keys := make([]string, 0, len(cfg.Params))
		for k := range cfg.Params {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			parts = append(parts, k+"="+cfg.Params[k])
		}
	}
	return strings.TrimSpace(strings.Join(parts, " "))
}

func loadFileConfig(cliPath string) (FileConfig, string, error) {
	cfgPath, err := resolveConfigPath(cliPath)
	if err != nil {
		return FileConfig{}, "", err
	}
	if cfgPath == "" {
		return FileConfig{}, "", nil
	}

	data, err := os.ReadFile(cfgPath)
	if err != nil {
		return FileConfig{}, "", err
	}

	cfg := FileConfig{}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return FileConfig{}, "", err
	}
	return cfg, cfgPath, nil
}

func resolveConfigPath(cliPath string) (string, error) {
	if path := strings.TrimSpace(cliPath); path != "" {
		abs, err := filepath.Abs(path)
		if err != nil {
			return "", err
		}
		if _, err := os.Stat(abs); err != nil {
			return "", err
		}
		return abs, nil
	}

	wd, err := os.Getwd()
	if err == nil {
		p := filepath.Join(wd, "server.json")
		if _, statErr := os.Stat(p); statErr == nil {
			return p, nil
		}
	}

	homeDir, err := os.UserHomeDir()
	if err == nil {
		p := filepath.Join(homeDir, ".mkp", "server.json")
		if _, statErr := os.Stat(p); statErr == nil {
			return p, nil
		}
	}
	return "", nil
}

func parsePositiveInt(raw string, fallback int) int {
	parsed, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func int64Ptr(v int64) *int64 {
	return &v
}

func parseBoolWithDefault(raw string, fallback bool) bool {
	value := strings.TrimSpace(raw)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}
	return parsed
}
