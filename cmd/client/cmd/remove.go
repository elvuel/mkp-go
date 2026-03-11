package cmd

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
)

func init() {
	var id string

	removeCmd := &cobra.Command{
		Use:   "remove",
		Short: "Delete a macro record",
		RunE: func(cmd *cobra.Command, args []string) error {
			id = strings.TrimSpace(id)
			if id == "" {
				return fmt.Errorf("id is required")
			}

			path := "/api/v1/directives/records/" + url.PathEscape(id)
			resp, err := sendJSON("DELETE", path, nil, true)
			if err != nil {
				return err
			}

			b, _ := json.MarshalIndent(resp, "", "  ")
			fmt.Println(string(b))
			return nil
		},
	}

	removeCmd.Flags().StringVarP(&id, "id", "i", "", "Macro record unique id")
	_ = removeCmd.MarkFlagRequired("id")
	rootCmd.AddCommand(removeCmd)
}
