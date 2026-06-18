package mkpgo

import (
	"errors"
	"strings"
	"testing"
)

func TestRawDirectiveAJSON2LogParseSuccess(t *testing.T) {
	parser := NewRawDirective_ajson2log()
	out := "ajson2log jsons/aaaa.json -o jsons/aaaa.log\n" +
		"I (140523) ajson2log: Written /eMMC/applog/jsons/aaaa.log (version=MKv2, mouse=1, kbd=0)\n" +
		"JSON -> log: /eMMC/applog/jsons/aaaa.log\n\n" +
		"cli>"

	got, err := parser.Parse("ajson2log jsons/aaaa.json -o jsons/aaaa.log", out)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if !strings.Contains(got, "JSON -> log: /eMMC/applog/jsons/aaaa.log") {
		t.Fatalf("Parse() = %q, want success output", got)
	}
}

func TestRawDirectiveAJSON2LogParseExecutionFailed(t *testing.T) {
	parser := NewRawDirective_ajson2log()
	out := "ajson2log jsons/aaaattt.json -o jsons/aaaa.log\n" +
		"E (374939) ajson2log: Cannot open /eMMC/applog/jsons/aaaattt.json\n" +
		"Command returned non-zero error code: 0xffffffff (ESP_FAIL)\n" +
		"cli>"

	_, err := parser.Parse("ajson2log jsons/aaaattt.json -o jsons/aaaa.log", out)
	if !errors.Is(err, ErrRawDirecitveExecutionFailed) {
		t.Fatalf("Parse() error = %v, want %v", err, ErrRawDirecitveExecutionFailed)
	}
}

func TestRawDirectiveAJSON2LogProperties(t *testing.T) {
	parser := NewRawDirective_ajson2log()
	if parser.String() != "ajson2log" {
		t.Fatalf("String() = %q, want ajson2log", parser.String())
	}
	if parser.IsJSONOutput() {
		t.Fatal("IsJSONOutput() = true, want false")
	}
	if parser.EOFFlag() != EOFCLI {
		t.Fatalf("EOFFlag() = %q, want %q", parser.EOFFlag(), EOFCLI)
	}
}

func TestAJSON2LogOptionCliArgs(t *testing.T) {
	if got := (&AJSON2LogOption{JSONPath: "jsons/aaaa.json", OutputLogPath: "jsons/aaaa.log"}).CliArgs(); len(got) != 3 || got[0] != "jsons/aaaa.json" || got[1] != "-o" || got[2] != "jsons/aaaa.log" {
		t.Fatalf("CliArgs() = %#v, want [jsons/aaaa.json -o jsons/aaaa.log]", got)
	}
	if got := (&AJSON2LogOption{JSONPath: "jsons/aaaa.json"}).CliArgs(); len(got) != 1 || got[0] != "jsons/aaaa.json" {
		t.Fatalf("CliArgs() without output = %#v, want [jsons/aaaa.json]", got)
	}
	if got := (&AJSON2LogOption{}).CliArgs(); got != nil {
		t.Fatalf("empty CliArgs() = %#v, want nil", got)
	}
	if got := (*AJSON2LogOption)(nil).CliArgs(); got != nil {
		t.Fatalf("nil CliArgs() = %#v, want nil", got)
	}
}
