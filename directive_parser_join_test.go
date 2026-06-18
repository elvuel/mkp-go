package mkpgo

import (
	"errors"
	"strings"
	"testing"
)

func TestRawDirectiveJoinParseConnected(t *testing.T) {
	parser := NewRawDirective_join()
	out := "join ssid password1234\n" +
		"I (29664) connect: Connecting to 'ssid'\n" +
		"I (31648) connect: Connected\n"

	got, err := parser.Parse("join ssid password1234", out)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if !strings.Contains(got, "connect: Connected") {
		t.Fatalf("Parse() = %q, want connected output", got)
	}
}

func TestRawDirectiveJoinParseConnectedWithoutArgs(t *testing.T) {
	parser := NewRawDirective_join()
	out := "join\n" +
		"I (736848) connect: Connecting to ''\n" +
		"ssid ChinaNet-9Wfg pass password1234\n" +
		"I (736872) connect: Connected\n" +
		"cli>"

	got, err := parser.Parse("join", out)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if !strings.Contains(got, "connect: Connected") {
		t.Fatalf("Parse() = %q, want connected output", got)
	}
}

func TestRawDirectiveJoinParseExecutionFailed(t *testing.T) {
	parser := NewRawDirective_join()
	out := "join ssid password1234\n" +
		"W (18376) connect: Connection timed out\n" +
		"Command returned non-zero error code: 0x1 (ERROR)\n"

	_, err := parser.Parse("join ssid password1234", out)
	if !errors.Is(err, ErrRawDirecitveExecutionFailed) {
		t.Fatalf("Parse() error = %v, want %v", err, ErrRawDirecitveExecutionFailed)
	}
}

func TestRawDirectiveJoinProperties(t *testing.T) {
	parser := NewRawDirective_join()
	if parser.String() != "join" {
		t.Fatalf("String() = %q, want join", parser.String())
	}
	if parser.IsJSONOutput() {
		t.Fatal("IsJSONOutput() = true, want false")
	}
	if parser.EOFFlag() != EOFCLI {
		t.Fatalf("EOFFlag() = %q, want %q", parser.EOFFlag(), EOFCLI)
	}
}

func TestJoinOptionCliArgs(t *testing.T) {
	opt := &JoinOption{SSID: "ssid", Password: "password1234"}
	got := strings.Join(opt.CliArgs(), " ")
	if got != "ssid password1234" {
		t.Fatalf("CliArgs() = %q, want %q", got, "ssid password1234")
	}

	if got := (&JoinOption{}).CliArgs(); got != nil {
		t.Fatalf("empty CliArgs() = %#v, want nil", got)
	}

	if got := (*JoinOption)(nil).CliArgs(); got != nil {
		t.Fatalf("nil CliArgs() = %#v, want nil", got)
	}
}
