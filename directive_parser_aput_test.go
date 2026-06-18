package mkpgo

import (
	"errors"
	"strings"
	"testing"
)

func TestRawDirectiveAPutParseSuccess(t *testing.T) {
	parser := NewRawDirective_aput()
	out := "aput applog/demo.log\n" +
		"Uploading /eMMC/applog/demo.log -> http://192.168.71.6:8000/upload\n\n" +
		"I (297931) httpfile: Upload OK (75 bytes) -> http://192.168.71.6:8000/upload\n" +
		"Upload OK\n\n" +
		"cli>"

	got, err := parser.Parse("aput applog/demo.log", out)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if !strings.Contains(got, "Upload OK") || !strings.Contains(got, "Uploading /eMMC/applog/demo.log") {
		t.Fatalf("Parse() = %q, want success output", got)
	}
}

func TestRawDirectiveAPutParseFailure(t *testing.T) {
	parser := NewRawDirective_aput()
	out := "aput applog/missing.log\n" +
		"Uploading /eMMC/applog/missing.log -> http://192.168.71.6:8000/upload\n\n" +
		"E (345995) httpfile: Failed to open file: /eMMC/applog/missing.log\n" +
		"Upload FAILED\n\n" +
		"Command returned non-zero error code: 0xffffffff (ESP_FAIL)\n" +
		"cli>"

	_, err := parser.Parse("aput applog/missing.log", out)
	if !errors.Is(err, ErrRawDirecitveExecutionFailed) {
		t.Fatalf("Parse() error = %v, want %v", err, ErrRawDirecitveExecutionFailed)
	}
}

func TestRawDirectiveAPutProperties(t *testing.T) {
	parser := NewRawDirective_aput()
	if parser.String() != "aput" {
		t.Fatalf("String() = %q, want aput", parser.String())
	}
	if parser.IsJSONOutput() {
		t.Fatal("IsJSONOutput() = true, want false")
	}
	if parser.EOFFlag() != EOFCLI {
		t.Fatalf("EOFFlag() = %q, want %q", parser.EOFFlag(), EOFCLI)
	}
}

func TestAPutOptionCliArgs(t *testing.T) {
	if got := (&APutOption{FilePath: "applog/demo.log"}).CliArgs(); len(got) != 1 || got[0] != "applog/demo.log" {
		t.Fatalf("CliArgs() = %#v, want [applog/demo.log]", got)
	}
	if got := (&APutOption{}).CliArgs(); got != nil {
		t.Fatalf("empty CliArgs() = %#v, want nil", got)
	}
	if got := (*APutOption)(nil).CliArgs(); got != nil {
		t.Fatalf("nil CliArgs() = %#v, want nil", got)
	}
}
