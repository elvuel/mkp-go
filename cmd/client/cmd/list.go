package cmd

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/spf13/cobra"
)

func init() {
	var limits int
	var name string

	listCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"l"},
		Short:   "List returns latest macro records",
		RunE: func(cmd *cobra.Command, args []string) error {
			if limits <= 0 {
				return fmt.Errorf("limits must be a positive integer")
			}

			values := url.Values{}
			values.Set("limits", fmt.Sprintf("%d", limits))
			if name != "" {
				values.Set("name", name)
			}
			path := "/api/v1/directives/list?" + values.Encode()
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
	listCmd.Flags().StringVar(&name, "name", "", "Filter records by name (substring match)")
	rootCmd.AddCommand(listCmd)
}
