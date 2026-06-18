package mkpgo

import (
	"errors"
	"testing"
)

func TestRawDirectiveAdumjParseSuccess(t *testing.T) {
	parser := NewRawDirective_adumj()
	out := "adumj demo\n" +
		"I (922627) alog: logfile /eMMC/applog/demo.log\n" +
		"{\n" +
		"  \"format\": \"mkp-action-v1\",\n" +
		"  \"version\": \"MKv2\",\n" +
		"  \"meta\": { \"width\": 1920, \"height\": 1080, \"startX\": 0, \"startY\": 0 },\n" +
		"  \"events\": [\n" +
		"    { \"MouseMove\": { \"x\": 1, \"y\": 2, \"ts\": 1064 } }\n" +
		"  ]\n" +
		"}\n" +
		"<EOF>\ncli>"

	got, err := parser.Parse("adumj demo", out)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if got == "" || got[0] != '{' {
		t.Fatalf("Parse() = %q, want JSON object", got)
	}

	dump := &ActionDump{}
	if err := parser.UnmarshalTo(got, dump); err != nil {
		t.Fatalf("UnmarshalTo() error = %v", err)
	}
	if dump.Format != "mkp-action-v1" || dump.Version != "MKv2" {
		t.Fatalf("dump = %#v, want format/version", dump)
	}
	if dump.Meta.Width != 1920 || len(dump.Events) != 1 {
		t.Fatalf("dump = %#v, want meta/events", dump)
	}
}

func TestRawDirectiveAdumjParseExecutionFailed(t *testing.T) {
	parser := NewRawDirective_adumj()
	out := "adumj missing\n" +
		"E (1084411) alog: Failed to open file /eMMC/applog/missing.log\n" +
		"Command returned non-zero error code: 0xffffffff (ESP_FAIL)\n" +
		"cli>"

	_, err := parser.Parse("adumj missing", out)
	if !errors.Is(err, ErrRawDirecitveExecutionFailed) {
		t.Fatalf("Parse() error = %v, want %v", err, ErrRawDirecitveExecutionFailed)
	}
}

func TestRawDirectiveAdumjProperties(t *testing.T) {
	parser := NewRawDirective_adumj()
	if parser.String() != "adumj" {
		t.Fatalf("String() = %q, want adumj", parser.String())
	}
	if !parser.IsJSONOutput() {
		t.Fatal("IsJSONOutput() = false, want true")
	}
	if parser.EOFFlag() != EOFCLI {
		t.Fatalf("EOFFlag() = %q, want %q", parser.EOFFlag(), EOFCLI)
	}
}

func TestAdumjOptionCliArgs(t *testing.T) {
	if got := (&AdumjOption{LogPath: "demo"}).CliArgs(); len(got) != 1 || got[0] != "demo" {
		t.Fatalf("CliArgs() = %#v, want [demo]", got)
	}
	if got := (&AdumjOption{}).CliArgs(); got != nil {
		t.Fatalf("empty CliArgs() = %#v, want nil", got)
	}
	if got := (*AdumjOption)(nil).CliArgs(); got != nil {
		t.Fatalf("nil CliArgs() = %#v, want nil", got)
	}
}
