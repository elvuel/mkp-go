package helper

import (
	"errors"
	"os"
	"strings"
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

func TestParseDirectiveResultWrapsDebugCapture(t *testing.T) {
	t.Setenv("MKP_DEBUG_SERIAL_DIR", t.TempDir())
	sp := mkpgo.NewSFSerialPort()

	_, _, err := parseDirectiveResult(sp, "wifi_auto", "wifi_auto\nnot-json\ncli>")
	if !errors.Is(err, mkpgo.ErrRawDirectiveParseFailed) {
		t.Fatalf("parseDirectiveResult() error = %v, want %v", err, mkpgo.ErrRawDirectiveParseFailed)
	}

	var debugErr *mkpgo.SerialDebugError
	if !errors.As(err, &debugErr) {
		t.Fatalf("parseDirectiveResult() error = %T, want *mkpgo.SerialDebugError", err)
	}
	if debugErr.DebugPath == "" {
		t.Fatal("parseDirectiveResult() debug path empty")
	}
	if !strings.Contains(debugErr.Error(), "serial summary:") {
		t.Fatalf("parseDirectiveResult() error = %q, want serial summary note", debugErr.Error())
	}

	content, readErr := os.ReadFile(debugErr.DebugPath)
	if readErr != nil {
		t.Fatalf("ReadFile(%q) error = %v", debugErr.DebugPath, readErr)
	}
	text := string(content)
	if !strings.Contains(text, "[summary]") {
		t.Fatalf("debug capture missing [summary]: %s", text)
	}
	if !strings.Contains(text, "directive: wifi_auto") {
		t.Fatalf("debug capture missing directive metadata: %s", text)
	}
	if !strings.Contains(text, "[raw_output]") {
		t.Fatalf("debug capture missing [raw_output]: %s", text)
	}
	if !strings.Contains(text, "not-json") {
		t.Fatalf("debug capture missing raw payload: %s", text)
	}
	if !strings.Contains(text, "[error]") {
		t.Fatalf("debug capture missing [error]: %s", text)
	}
}
