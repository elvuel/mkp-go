package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	var limits int

	listCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"l"},
		Short:   "List returns latest macro records",
		RunE: func(cmd *cobra.Command, args []string) error {
			if limits <= 0 {
				return fmt.Errorf("limits must be a positive integer")
			}

			path := fmt.Sprintf("/api/v1/directives/list?limits=%d", limits)
			resp, err := sendJSON("GET", path, nil, true)
			if err != nil {
				return err
			}

			b, _ := json.MarshalIndent(resp, "", "  ")
			fmt.Println(string(b))
			return nil
		},
	}

	listCmd.Flags().IntVar(&limits, "limits", 10, "Number of latest macro records to list")
	rootCmd.AddCommand(listCmd)
}
