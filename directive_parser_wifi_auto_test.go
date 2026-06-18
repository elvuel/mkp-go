package mkpgo

import (
	"errors"
	"testing"
)

func TestRawDirectiveWifiAutoParseStatusOn(t *testing.T) {
	parser := NewRawDirective_wifi_auto()
	out := "wifi_auto\n" +
		"auto: on\n" +
		"cli>"

	got, err := parser.Parse("wifi_auto", out)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if got != "on" {
		t.Fatalf("Parse() = %q, want on", got)
	}
}

func TestRawDirectiveWifiAutoParseStatusOff(t *testing.T) {
	parser := NewRawDirective_wifi_auto()
	out := "wifi_auto\n" +
		"auto: off\n"

	got, err := parser.Parse("wifi_auto", out)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if got != "off" {
		t.Fatalf("Parse() = %q, want off", got)
	}
}

func TestRawDirectiveWifiAutoParseSetState(t *testing.T) {
	parser := NewRawDirective_wifi_auto()
	out := "wifi_auto 1\ncli>"

	got, err := parser.Parse("wifi_auto 1", out)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if got != "" {
		t.Fatalf("Parse() = %q, want empty", got)
	}
}

func TestRawDirectiveWifiAutoParseExecutionFailed(t *testing.T) {
	parser := NewRawDirective_wifi_auto()
	out := "wifi_auto\nCommand returned non-zero error code: 0x1 (ERROR)\n"

	_, err := parser.Parse("wifi_auto", out)
	if !errors.Is(err, ErrRawDirecitveExecutionFailed) {
		t.Fatalf("Parse() error = %v, want %v", err, ErrRawDirecitveExecutionFailed)
	}
}

func TestRawDirectiveWifiAutoProperties(t *testing.T) {
	parser := NewRawDirective_wifi_auto()
	if parser.String() != "wifi_auto" {
		t.Fatalf("String() = %q, want wifi_auto", parser.String())
	}
	if parser.IsJSONOutput() {
		t.Fatal("IsJSONOutput() = true, want false")
	}
	if parser.EOFFlag() != EOFCLI {
		t.Fatalf("EOFFlag() = %q, want %q", parser.EOFFlag(), EOFCLI)
	}
}

func TestWifiAutoOptionCliArgs(t *testing.T) {
	if got := (&WifiAutoOption{State: "1"}).CliArgs(); len(got) != 1 || got[0] != "1" {
		t.Fatalf("CliArgs() = %#v, want [1]", got)
	}
	if got := (&WifiAutoOption{}).CliArgs(); got != nil {
		t.Fatalf("empty CliArgs() = %#v, want nil", got)
	}
	if got := (*WifiAutoOption)(nil).CliArgs(); got != nil {
		t.Fatalf("nil CliArgs() = %#v, want nil", got)
	}
}
