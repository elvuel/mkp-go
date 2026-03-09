package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

type stopRequest struct {
	ID string `json:"id"`
}

func init() {
	var currentID string

	stopCmd := &cobra.Command{
		Use:   "stop",
		Short: "Stop astop recording",
		RunE: func(cmd *cobra.Command, args []string) error {
			resp, err := sendJSON("POST", "/api/v1/directives/astop", stopRequest{ID: currentID}, true)
			if err != nil {
				return err
			}

			b, _ := json.MarshalIndent(resp, "", "  ")
			fmt.Println(string(b))
			return nil
		},
	}

	stopCmd.Flags().StringVar(&currentID, "id", "", "Current alog ID")
	_ = stopCmd.MarkFlagRequired("id")

	rootCmd.AddCommand(stopCmd)
}
