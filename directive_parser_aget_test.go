package mkpgo

import (
	"errors"
	"strings"
	"testing"
)

func TestRawDirectiveAGetParseSuccess(t *testing.T) {
	parser := NewRawDirective_aget()
	out := "aget applog/demo.log\n" +
		"Downloading http://192.168.71.6:8000/applog/demo.log\n\n" +
		"I (2925224) heh: HTTP_EVENT_ON_CONNECTED\n" +
		"I (2925424) heh: HTTP_EVENT_ON_DATA, len=75\n" +
		"I (2925424) heh: HTTP_EVENT_ON_FINISH\n" +
		"I (2925424) httpfile: GET status=200, length=75\n" +
		"Saved to /eMMC/applog/demo.log (75 bytes)\n\n" +
		"I (2925448) heh: HTTP_EVENT_DISCONNECTED\n" +
		"cli>"

	got, err := parser.Parse("aget applog/demo.log", out)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if !strings.Contains(got, "GET status=200") || !strings.Contains(got, "Saved to /eMMC/applog/demo.log") {
		t.Fatalf("Parse() = %q, want success output", got)
	}
}

func TestRawDirectiveAGetParseHTTPFailure(t *testing.T) {
	parser := NewRawDirective_aget()
	out := "aget applog/missing.log\n" +
		"Downloading http://192.168.71.6:8000/applog/missing.log\n\n" +
		"I (2970792) httpfile: GET status=404, length=76\n" +
		"E (2970792) httpfile: Server returned status 404\n" +
		"I (2970816) heh: HTTP_EVENT_DISCONNECTED\n" +
		"cli>"

	_, err := parser.Parse("aget applog/missing.log", out)
	if !errors.Is(err, ErrRawDirecitveExecutionFailed) {
		t.Fatalf("Parse() error = %v, want %v", err, ErrRawDirecitveExecutionFailed)
	}
}

func TestRawDirectiveAGetParseExecutionFailed(t *testing.T) {
	parser := NewRawDirective_aget()
	out := "aget applog/demo.log\n" +
		"Command returned non-zero error code: 0x1 (ERROR)\n" +
		"cli>"

	_, err := parser.Parse("aget applog/demo.log", out)
	if !errors.Is(err, ErrRawDirecitveExecutionFailed) {
		t.Fatalf("Parse() error = %v, want %v", err, ErrRawDirecitveExecutionFailed)
	}
}

func TestRawDirectiveAGetProperties(t *testing.T) {
	parser := NewRawDirective_aget()
	if parser.String() != "aget" {
		t.Fatalf("String() = %q, want aget", parser.String())
	}
	if parser.IsJSONOutput() {
		t.Fatal("IsJSONOutput() = true, want false")
	}
	if parser.EOFFlag() != EOFCLI {
		t.Fatalf("EOFFlag() = %q, want %q", parser.EOFFlag(), EOFCLI)
	}
}

func TestAGetOptionCliArgs(t *testing.T) {
	if got := (&AGetOption{FilePath: "applog/demo.log"}).CliArgs(); len(got) != 1 || got[0] != "applog/demo.log" {
		t.Fatalf("CliArgs() = %#v, want [applog/demo.log]", got)
	}
	if got := (&AGetOption{}).CliArgs(); got != nil {
		t.Fatalf("empty CliArgs() = %#v, want nil", got)
	}
	if got := (*AGetOption)(nil).CliArgs(); got != nil {
		t.Fatalf("nil CliArgs() = %#v, want nil", got)
	}
}
