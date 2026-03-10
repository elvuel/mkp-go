package cmd

import (
	"encoding/json"
	"fmt"
	"time"

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

	recordingCmd := &cobra.Command{
		Use:     "recording",
		Aliases: []string{"record", "r"},
		Short:   "Start a macro recording",
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

			if logName == "" {
				logName = fmt.Sprintf("mkp-%s", time.Now().Format("2006-01-02 15:04:05"))
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

	recordingCmd.Flags().StringVarP(&logName, "name", "n", "", "Record display name")
	recordingCmd.Flags().IntVar(&width, "width", 0, "Screen width")
	recordingCmd.Flags().IntVar(&height, "height", 0, "Screen height")
	recordingCmd.Flags().IntVar(&stposx, "stposx", 0, "Cursor start coordiante x")
	recordingCmd.Flags().IntVar(&stposy, "stposy", 0, "Cursor start coordiante y")
	// _ = recordingCmd.MarkFlagRequired("name")

	rootCmd.AddCommand(recordingCmd)
}
