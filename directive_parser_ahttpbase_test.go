package mkpgo

import (
	"errors"
	"testing"
)

func TestRawDirectiveAHTTPBaseParseQuerySuccess(t *testing.T) {
	parser := NewRawDirective_ahttpbase()
	out := "ahttpbase\n" +
		"W (954891) setupnvs: Error reading 'ahttpbase' from NVS: ESP_ERR_NVS_NOT_FOUND\n" +
		"{ \"ahttpbase\": \"\" }\n\n" +
		"<EOF>\ncli>"

	got, err := parser.Parse("ahttpbase", out)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	base := &AHTTPBase{}
	if err := parser.UnmarshalTo(got, base); err != nil {
		t.Fatalf("UnmarshalTo() error = %v", err)
	}
	if base.AHTTPBase != "" {
		t.Fatalf("AHTTPBase = %q, want empty", base.AHTTPBase)
	}
}

func TestRawDirectiveAHTTPBaseParseSetSuccess(t *testing.T) {
	parser := NewRawDirective_ahttpbase()
	out := "ahttpbase http://localhost:3000\n" +
		"OK\n\n" +
		"cli>"

	got, err := parser.Parse("ahttpbase http://localhost:3000", out)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	base := &AHTTPBase{}
	if err := parser.UnmarshalTo(got, base); err != nil {
		t.Fatalf("UnmarshalTo() error = %v", err)
	}
	if base.AHTTPBase != "http://localhost:3000" {
		t.Fatalf("AHTTPBase = %q, want http://localhost:3000", base.AHTTPBase)
	}
}

func TestRawDirectiveAHTTPBaseParseExecutionFailed(t *testing.T) {
	parser := NewRawDirective_ahttpbase()
	out := "ahttpbase http://localhost:3000\n" +
		"Command returned non-zero error code: 0x1 (ERROR)\n" +
		"cli>"

	_, err := parser.Parse("ahttpbase http://localhost:3000", out)
	if !errors.Is(err, ErrRawDirecitveExecutionFailed) {
		t.Fatalf("Parse() error = %v, want %v", err, ErrRawDirecitveExecutionFailed)
	}
}

func TestRawDirectiveAHTTPBaseProperties(t *testing.T) {
	parser := NewRawDirective_ahttpbase()
	if parser.String() != "ahttpbase" {
		t.Fatalf("String() = %q, want ahttpbase", parser.String())
	}
	if !parser.IsJSONOutput() {
		t.Fatal("IsJSONOutput() = false, want true")
	}
	if parser.EOFFlag() != EOFCLI {
		t.Fatalf("EOFFlag() = %q, want %q", parser.EOFFlag(), EOFCLI)
	}
}

func TestAHTTPBaseOptionCliArgs(t *testing.T) {
	if got := (&AHTTPBaseOption{URL: "http://localhost:3000"}).CliArgs(); len(got) != 1 || got[0] != "http://localhost:3000" {
		t.Fatalf("CliArgs() = %#v, want [http://localhost:3000]", got)
	}
	if got := (&AHTTPBaseOption{}).CliArgs(); got != nil {
		t.Fatalf("empty CliArgs() = %#v, want nil", got)
	}
	if got := (*AHTTPBaseOption)(nil).CliArgs(); got != nil {
		t.Fatalf("nil CliArgs() = %#v, want nil", got)
	}
}
