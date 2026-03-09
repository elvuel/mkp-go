package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

func init() {
	authCmd := &cobra.Command{
		Use:   "auth",
		Short: "Manage JWT token and credentials",
	}

	authCmd.AddCommand(newAuthLoginCmd())
	authCmd.AddCommand(newAuthTokenCmd())
	authCmd.AddCommand(newAuthRevokeCmd())
	rootCmd.AddCommand(authCmd)
}

func newAuthLoginCmd() *cobra.Command {
	var (
		username string
		password string
	)

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Fetch and store a new JWT token",
		RunE: func(cmd *cobra.Command, args []string) error {
			if strings.TrimSpace(username) == "" {
				username = strings.TrimSpace(authUser)
			}
			if strings.TrimSpace(password) == "" {
				password = strings.TrimSpace(authPassword)
			}
			if username == "" || password == "" {
				return fmt.Errorf("username/password are required")
			}

			token, expiresAt, err := fetchToken(username, password)
			if err != nil {
				return err
			}

			state, err := loadClientConfig()
			if err != nil {
				return err
			}
			state.ServerAddr = strings.TrimSpace(serverAddr)
			state.Mode = strings.TrimSpace(clientMode)
			state.Username = username
			state.Password = password
			state.Token = token
			state.TokenExpiresAt = expiresAt.Format(time.RFC3339)
			if httpTimeout > 0 {
				state.TimeoutSeconds = int(httpTimeout.Seconds())
			}

			if err := saveClientConfig(state); err != nil {
				return err
			}

			fmt.Printf("token saved (expires at %s)\n", expiresAt.Format(time.RFC3339))
			return nil
		},
	}

	cmd.Flags().StringVar(&username, "username", "", "Login username")
	cmd.Flags().StringVar(&password, "password", "", "Login password")
	return cmd
}

func newAuthTokenCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "token",
		Short: "Print current token (auto-refresh if needed)",
		RunE: func(cmd *cobra.Command, args []string) error {
			token := strings.TrimSpace(authToken)
			var err error
			if token == "" {
				token, err = ensureManagedToken(false)
				if err != nil {
					return err
				}
			}
			fmt.Println(token)
			return nil
		},
	}
}

func newAuthRevokeCmd() *cobra.Command {
	var forgetCreds bool

	cmd := &cobra.Command{
		Use:   "revoke",
		Short: "Revoke local token cache (server token remains valid until expiry)",
		RunE: func(cmd *cobra.Command, args []string) error {
			state, err := loadClientConfig()
			if err != nil {
				return err
			}

			state.Token = ""
			state.TokenExpiresAt = ""
			if forgetCreds {
				state.Username = ""
				state.Password = ""
			}

			if err := saveClientConfig(state); err != nil {
				return err
			}

			if forgetCreds {
				fmt.Println("local token and credentials cleared")
			} else {
				fmt.Println("local token cleared")
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&forgetCreds, "forget", false, "Also clear stored username/password")
	return cmd
}
