package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	stopCmd := &cobra.Command{
		Use:     "stop",
		Aliases: []string{"s"},
		Short:   "Stop current recording",
		RunE: func(cmd *cobra.Command, args []string) error {
			resp, err := sendJSON("POST", "/api/v1/directives/astop", nil, true)
			if err != nil {
				return err
			}

			b, _ := json.MarshalIndent(resp, "", "  ")
			fmt.Println(string(b))
			return nil
		},
	}

	rootCmd.AddCommand(stopCmd)
}
