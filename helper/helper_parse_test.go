package helper

import (
	"errors"
	"testing"

	mkpgo "github.com/elvuel/mkp-go"
)

func TestRequireSyncOutputEnabled(t *testing.T) {
	sp := mkpgo.NewSFSerialPort()
	if err := requireSyncOutputEnabled(sp); err != nil {
		t.Fatalf("requireSyncOutputEnabled() error = %v", err)
	}

	sp.DisableSyncOutput()
	err := requireSyncOutputEnabled(sp)
	if !errors.Is(err, errSyncOutputDisabled) {
		t.Fatalf("requireSyncOutputEnabled() error = %v, want %v", err, errSyncOutputDisabled)
	}
}

func TestParseDirectiveResultStringOutput(t *testing.T) {
	sp := mkpgo.NewSFSerialPort()

	parser, parsed, err := parseDirectiveResult(sp, "wifi_auto", "wifi_auto\nauto: on\ncli>")
	if err != nil {
		t.Fatalf("parseDirectiveResult() error = %v", err)
	}
	if parser == nil {
		t.Fatal("parseDirectiveResult() parser = nil")
	}
	if parser.IsJSONOutput() {
		t.Fatal("parseDirectiveResult() parser unexpectedly reports JSON output")
	}
	if parsed != "on" {
		t.Fatalf("parseDirectiveResult() parsed = %q, want %q", parsed, "on")
	}
}

func TestParseDirectiveResultJSONOutput(t *testing.T) {
	sp := mkpgo.NewSFSerialPort()

	parser, parsed, err := parseDirectiveResult(sp, "sn", "sn\n{\"sn\":\"demo-sn\"}\n")
	if err != nil {
		t.Fatalf("parseDirectiveResult() error = %v", err)
	}
	if parser == nil {
		t.Fatal("parseDirectiveResult() parser = nil")
	}
	if !parser.IsJSONOutput() {
		t.Fatal("parseDirectiveResult() parser should report JSON output")
	}
	if parsed == "" {
		t.Fatal("parseDirectiveResult() parsed empty JSON payload")
	}

	var sn mkpgo.SN
	if err := parser.UnmarshalTo(parsed, &sn); err != nil {
		t.Fatalf("UnmarshalTo() error = %v", err)
	}
	if sn.SN != "demo-sn" {
		t.Fatalf("UnmarshalTo() SN = %q, want %q", sn.SN, "demo-sn")
	}
}

func TestParseDirectiveResultMissingParser(t *testing.T) {
	sp := mkpgo.NewSFSerialPort()

	_, _, err := parseDirectiveResult(sp, "m10 --port 1 --x 1", "")
	if !errors.Is(err, mkpgo.ErrDirectiveParserMissing) {
		t.Fatalf("parseDirectiveResult() error = %v, want %v", err, mkpgo.ErrDirectiveParserMissing)
	}
}

func TestParseDirectiveResultPropagatesParserError(t *testing.T) {
	sp := mkpgo.NewSFSerialPort()

	_, _, err := parseDirectiveResult(sp, "wifi_auto", "wifi_auto\nCommand returned non-zero error code: 0x1 (ERROR)\n")
	if !errors.Is(err, mkpgo.ErrRawDirecitveExecutionFailed) {
		t.Fatalf("parseDirectiveResult() error = %v, want %v", err, mkpgo.ErrRawDirecitveExecutionFailed)
	}
}
