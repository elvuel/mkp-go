package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/elvuel/mkp-go/cmd/client/helper"
	"github.com/spf13/cobra"
)

type versionResponse struct {
	OK            bool        `json:"ok"`
	Directive     string      `json:"directive"`
	ServerVersion string      `json:"mkp_server"`
	MKPVersion    *mkpVersion `json:"mkp_device"`
	Error         string      `json:"error"`
}

type mkpVersion struct {
	UVersion string `json:"uver"`
	AVersion string `json:"aver"`
}

func init() {
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Show mkp agent & server and device version info",
		RunE: func(cmd *cobra.Command, args []string) error {
			resp, err := sendJSON("GET", "/api/v1/version", nil, false)
			if err != nil {
				return err
			}

			parsed, err := parseVersionResponse(resp)
			if err != nil {
				return err
			}

			fmt.Println(renderVersionTable(parsed))
			return nil
		},
	}

	rootCmd.AddCommand(versionCmd)
}

func parseVersionResponse(resp map[string]any) (*versionResponse, error) {
	if ok, okCast := resp["ok"].(bool); okCast && !ok {
		if msg, okMsg := resp["error"].(string); okMsg && strings.TrimSpace(msg) != "" {
			return nil, fmt.Errorf("request failed: %s", msg)
		}
		return nil, fmt.Errorf("request failed")
	}

	data, err := json.Marshal(resp)
	if err != nil {
		return nil, err
	}

	parsed := &versionResponse{}
	if err := json.Unmarshal(data, parsed); err != nil {
		return nil, fmt.Errorf("invalid version response: %w", err)
	}
	return parsed, nil
}

func renderVersionTable(info *versionResponse) string {
	headers := []string{"Component", "Version"}
	rows := make([][]string, 0, 3)

	serverVersion := defaultVersionValue(info.ServerVersion)
	rows = append(rows, []string{"mkp agent server", serverVersion})
	rows = append(rows, []string{"mkp agent client", rootCmd.Version})

	if info.MKPVersion != nil {
		rows = append(rows, []string{"mkp device uver", defaultVersionValue(info.MKPVersion.UVersion)})
		rows = append(rows, []string{"mkp device aver", defaultVersionValue(info.MKPVersion.AVersion)})
	} else {
		rows = append(rows, []string{"mkp device uver", "unknown"})
		rows = append(rows, []string{"mkp device aver", "unknown"})
	}

	var b strings.Builder
	b.WriteString("| ")
	b.WriteString(strings.Join(headers, " | "))
	b.WriteString(" |\n| ")
	b.WriteString(strings.Repeat("--- | ", len(headers)-1))
	b.WriteString("--- |\n")

	for _, row := range rows {
		values := []string{
			helper.EscapeTableValue(row[0]),
			helper.EscapeTableValue(row[1]),
		}
		b.WriteString("| ")
		b.WriteString(strings.Join(values, " | "))
		b.WriteString(" |\n")
	}

	return b.String()
}

func defaultVersionValue(value string) string {
	if strings.TrimSpace(value) == "" {
		return "unknown"
	}
	return value
}
