package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

type logRequest struct {
	LogName string `json:"log_name"`
	Width   int    `json:"width,omitempty"`
	Height  int    `json:"height,omitempty"`
	StPosX  *int   `json:"stposx,omitempty"`
	StPosY  *int   `json:"stposy,omitempty"`
}

func init() {
	var (
		logName string
		width   int
		height  int
		stposx  int
		stposy  int
	)

	logCmd := &cobra.Command{
		Use:   "log",
		Short: "Start alog recording",
		RunE: func(cmd *cobra.Command, args []string) error {
			req := logRequest{
				LogName: logName,
				Width:   width,
				Height:  height,
			}
			if cmd.Flags().Changed("stposx") {
				req.StPosX = &stposx
			}
			if cmd.Flags().Changed("stposy") {
				req.StPosY = &stposy
			}

			resp, err := sendJSON("POST", "/api/v1/directives/alog", req, true)
			if err != nil {
				return err
			}

			b, _ := json.MarshalIndent(resp, "", "  ")
			fmt.Println(string(b))
			return nil
		},
	}

	logCmd.Flags().StringVarP(&logName, "name", "n", "", "Log name")
	logCmd.Flags().IntVar(&width, "width", 0, "Screen width")
	logCmd.Flags().IntVar(&height, "height", 0, "Screen height")
	logCmd.Flags().IntVar(&stposx, "stposx", 0, "Start point x")
	logCmd.Flags().IntVar(&stposy, "stposy", 0, "Start point y")
	_ = logCmd.MarkFlagRequired("name")

	rootCmd.AddCommand(logCmd)
}
