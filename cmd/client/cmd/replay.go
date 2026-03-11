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

	replayCmd := &cobra.Command{
		Use:   "replay",
		Short: "Replay a macro record",
		RunE: func(cmd *cobra.Command, args []string) error {
			id = strings.TrimSpace(id)
			if id == "" {
				return fmt.Errorf("id is required")
			}

			path := "/api/v1/directives/aplay/" + url.PathEscape(id)
			resp, err := sendJSON("POST", path, nil, true)
			if err != nil {
				return err
			}

			b, _ := json.MarshalIndent(resp, "", "  ")
			fmt.Println(string(b))
			return nil
		},
	}

	replayCmd.Flags().StringVarP(&id, "id", "i", "", "Macro record unique id")
	_ = replayCmd.MarkFlagRequired("id")
	rootCmd.AddCommand(replayCmd)
}
