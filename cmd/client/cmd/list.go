package cmd

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/elvuel/mkp-go/cmd/client/helper"
	"github.com/spf13/cobra"
)

type listRecord struct {
	Name         string `json:"name"`
	UniqueID     string `json:"unique_id"`
	MKPPath      string `json:"mkp_path"`
	StartPointX  int    `json:"start_point_x"`
	StartPointY  int    `json:"start_point_y"`
	ScreenWidth  int    `json:"screen_width"`
	ScreenHeight int    `json:"screen_height"`
	OS           string `json:"os"`
	Seconds      int    `json:"seconds"`
	Milliseconds int    `json:"milliseconds"`
	CreatedAt    string `json:"created_at"`
}

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

			if resp["count"].(float64) == 0 {
				fmt.Println("No records found")
				return nil
			}

			records, err := parseListRecords(resp)
			if err != nil {
				return err
			}

			fmt.Println(renderListTable(records))
			return nil
		},
	}

	listCmd.Flags().IntVarP(&limits, "limits", "l", 10, "Number of latest macro records to list")
	listCmd.Flags().StringVarP(&name, "name", "n", "", "Filter records by name (substring match)")
	rootCmd.AddCommand(listCmd)
}

func parseListRecords(resp map[string]any) ([]listRecord, error) {
	raw, ok := resp["records"]
	if !ok {
		return nil, fmt.Errorf("response missing records")
	}

	data, err := json.Marshal(raw)
	if err != nil {
		return nil, err
	}

	var records []listRecord
	if err := json.Unmarshal(data, &records); err != nil {
		return nil, fmt.Errorf("invalid records: %w", err)
	}
	return records, nil
}

func renderListTable(records []listRecord) string {
	headers := []string{
		"Name",
		"RID",
		"Location",
		"Cursor Initial Position",
		"Screen Size",
		"OS",
		"Length",
		"created At",
	}

	var b strings.Builder
	b.WriteString("| ")
	b.WriteString(strings.Join(headers, " | "))
	b.WriteString(" |\n| ")
	b.WriteString(strings.Repeat("--- | ", len(headers)-1))
	b.WriteString("--- |\n")

	for _, record := range records {
		values := []string{
			helper.EscapeTableValue(record.Name),
			helper.EscapeTableValue(record.UniqueID),
			helper.EscapeTableValue(record.MKPPath),
			helper.EscapeTableValue(fmt.Sprintf("(%d,%d)", record.StartPointX, record.StartPointY)),
			helper.EscapeTableValue(fmt.Sprintf("%dx%d", record.ScreenWidth, record.ScreenHeight)),
			helper.EscapeTableValue(record.OS),
			helper.EscapeTableValue(formatLength(record.Seconds, record.Milliseconds)),
			helper.EscapeTableValue(record.CreatedAt),
		}
		b.WriteString("| ")
		b.WriteString(strings.Join(values, " | "))
		b.WriteString(" |\n")
	}

	return b.String()
}

func formatLength(seconds, milliseconds int) string {
	return fmt.Sprintf("%ds %dms", seconds, milliseconds)
}
