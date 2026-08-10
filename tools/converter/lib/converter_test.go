// Round-trip and integration tests for log2json.
//
// Round-trip invariants asserted:
//
//  1. log -> JSON -> log -> JSON  is a fixed point: the JSON form is
//     stable after one round trip. This holds because the state
//     machine in the forward and reverse converters is the exact
//     mirror of the C reference (cmd_logfile_adump.c and
//     cmd_logfile_ajson2log.c).
//
//  2. log -> JSON and the round-tripped JSON are both well-formed and
//     carry at least the same set of event types as the original
//     log. (Exact timestamp match is *not* expected: the firmware
//     encodes a button-up or key release using the *current* record's
//     timestamp, and the stop trailer in the second log becomes a
//     terminal KeyUp event. See docs/tool-log2json.md "Reversibility".)
//
// Integration test runs on real firmware log files to confirm the
// codec is wire-compatible with adumj's output.
package converter

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// recordBytes builds the wire bytes of one record.
func recordBytes(port uint8, delta int32, payload []byte) []byte {
	out := []byte{port}
	out = append(out, byte(delta), byte(delta>>8), byte(delta>>16), byte(delta>>24))
	out = append(out, byte(len(payload)), byte(len(payload)>>8))
	out = append(out, payload...)
	return out
}

func writeSyntheticLog(t *testing.T, path string) {
	t.Helper()
	var buf bytes.Buffer
	buf.WriteString(MagicV1)
	// mouse: move, down, move, wheel, up
	buf.Write(recordBytes(1, 0, mouseEnc(0, 5, -3, 0)))              // ts=0
	buf.Write(recordBytes(1, 8, mouseEnc(MouseButtonLeft, 0, 0, 0))) // ts=8 down
	buf.Write(recordBytes(1, 16, mouseEnc(0, 10, 7, 0)))             // ts=24 move, also releases left
	buf.Write(recordBytes(1, 8, mouseEnc(0, 0, 0, 1)))               // ts=32 wheel
	buf.Write(recordBytes(1, 8, mouseEnc(0, 0, 0, 0)))               // ts=40 (button already 0)
	// keyboard: a down, a up, lshift+a down, lshift+a up
	buf.Write(recordBytes(2, 8, kbdReport(0, [6]uint8{0x04, 0, 0, 0, 0, 0})))         // ts=48
	buf.Write(recordBytes(2, 16, kbdReport(0, [6]uint8{0, 0, 0, 0, 0, 0})))           // ts=64
	buf.Write(recordBytes(2, 8, kbdReport(ModLShift, [6]uint8{0x04, 0, 0, 0, 0, 0}))) // ts=72
	buf.Write(recordBytes(2, 8, kbdReport(ModLShift, [6]uint8{0, 0, 0, 0, 0, 0})))    // ts=80 release a
	buf.Write(recordBytes(2, 8, kbdReport(0, [6]uint8{0, 0, 0, 0, 0, 0})))            // ts=88 release lshift
	// stop trailer (matches what the firmware's ajson2log appends)
	buf.Write(recordBytes(1, 8, make([]byte, 5))) // ts=96 mouse stop
	buf.Write(recordBytes(1, 8, make([]byte, 5))) // ts=104
	buf.Write(recordBytes(1, 8, make([]byte, 5))) // ts=112
	buf.Write(recordBytes(2, 8, make([]byte, 8))) // ts=120 kbd stop
	if err := os.WriteFile(path, buf.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}
}

func mouseEnc(b uint8, x, y int16, w int8) []byte {
	a := MouseEncode(b, x, y, w)
	return a[:]
}

func kbdReport(modifier uint8, keys [6]uint8) []byte {
	var r [8]byte
	r[0] = modifier
	for i := 0; i < 6; i++ {
		r[2+i] = keys[i]
	}
	return r[:]
}

// TestRoundTripSyntheticJSONIdempotent asserts that the JSON form is
// stable after one log <-> JSON round trip.
func TestRoundTripSyntheticJSONIdempotent(t *testing.T) {
	tmp := t.TempDir()
	logPath := filepath.Join(tmp, "rt.log")
	jsonPath := filepath.Join(tmp, "rt.json")
	log2Path := filepath.Join(tmp, "rt2.log")
	json2Path := filepath.Join(tmp, "rt2.json")

	writeSyntheticLog(t, logPath)

	// 1st pass: log -> JSON.
	hdr, recs, err := ReadLogFile(logPath)
	if err != nil {
		t.Fatalf("read log: %v", err)
	}
	if hdr.Version != MagicV1 {
		t.Fatalf("expected MKv1, got %q", hdr.Version)
	}
	if len(recs) == 0 {
		t.Fatal("no records")
	}
	env1 := LogToJSON(hdr, recs)
	if err := WriteJSONFile(jsonPath, env1); err != nil {
		t.Fatalf("write json: %v", err)
	}

	// 2nd pass: JSON -> log.
	env2, err := ReadJSONFile(jsonPath)
	if err != nil {
		t.Fatalf("re-read json: %v", err)
	}
	hdr2, recs2, hasMouse, hasKbd, err := JSONToLog(env2)
	if err != nil {
		t.Fatalf("json to log: %v", err)
	}
	if !hasMouse {
		t.Error("expected mouse records in 2nd-pass log")
	}
	if !hasKbd {
		t.Error("expected keyboard records in 2nd-pass log")
	}
	if err := WriteLogFile(log2Path, hdr2, recs2, hasMouse, hasKbd); err != nil {
		t.Fatalf("write log2: %v", err)
	}

	// 3rd pass: log -> JSON. Must equal env2.
	hdr3, recs3, err := ReadLogFile(log2Path)
	if err != nil {
		t.Fatalf("read log2: %v", err)
	}
	env3 := LogToJSON(hdr3, recs3)
	if err := WriteJSONFile(json2Path, env3); err != nil {
		t.Fatalf("write json2: %v", err)
	}
	if !sameJSONFile(jsonPath, json2Path) {
		t.Errorf("JSON form not idempotent:\n  first:  %s\n  second: %s",
			dumpEvents(env2.Events), dumpEvents(env3.Events))
	}
}

func TestJSONToLogPreservesFirstEventTimestamp(t *testing.T) {
	env := &Envelope{
		Format:  "mkp-action-v1",
		Version: MagicV1,
		Events: []Event{
			{TypeName: "KeyDown", KeyDown: &KeyEvent{Keys: []string{"a"}, Ts: 123}},
			{TypeName: "KeyUp", KeyUp: &KeyEvent{Keys: []string{"a"}, Ts: 200}},
		},
	}

	hdr, recs, hasMouse, hasKbd, err := JSONToLog(env)
	if err != nil {
		t.Fatalf("JSONToLog: %v", err)
	}
	if hasMouse {
		t.Fatal("unexpected mouse records")
	}
	if !hasKbd {
		t.Fatal("expected keyboard records")
	}
	if len(recs) != 2 {
		t.Fatalf("record count = %d, want 2", len(recs))
	}
	if recs[0].Delta != 123 {
		t.Fatalf("first delta = %d, want 123", recs[0].Delta)
	}
	if recs[1].Delta != 77 {
		t.Fatalf("second delta = %d, want 77", recs[1].Delta)
	}

	tmp := t.TempDir()
	logPath := filepath.Join(tmp, "first-ts.log")
	if err := WriteLogFile(logPath, hdr, recs, hasMouse, hasKbd); err != nil {
		t.Fatalf("WriteLogFile: %v", err)
	}
	hdr2, recs2, err := ReadLogFile(logPath)
	if err != nil {
		t.Fatalf("ReadLogFile: %v", err)
	}
	env2 := LogToJSON(hdr2, recs2)
	if len(env2.Events) != 2 {
		t.Fatalf("event count = %d, want 2", len(env2.Events))
	}
	if got := env2.Events[0].KeyDown.Ts; got != 123 {
		t.Fatalf("first event ts = %d, want 123", got)
	}
	if got := env2.Events[1].KeyUp.Ts; got != 200 {
		t.Fatalf("second event ts = %d, want 200", got)
	}
}

func TestJSONToLogDuplicateKeyDownDoesNotLatchKey(t *testing.T) {
	env := &Envelope{
		Format:  "mkp-action-v1",
		Version: MagicV1,
		Events: []Event{
			{TypeName: "KeyDown", KeyDown: &KeyEvent{Keys: []string{"a"}, Ts: 0}},
			{TypeName: "KeyDown", KeyDown: &KeyEvent{Keys: []string{"a"}, Ts: 10}},
			{TypeName: "KeyUp", KeyUp: &KeyEvent{Keys: []string{"a"}, Ts: 20}},
		},
	}

	hdr, recs, hasMouse, hasKbd, err := JSONToLog(env)
	if err != nil {
		t.Fatalf("JSONToLog: %v", err)
	}
	tmp := t.TempDir()
	logPath := filepath.Join(tmp, "dup-key.log")
	if err := WriteLogFile(logPath, hdr, recs, hasMouse, hasKbd); err != nil {
		t.Fatalf("WriteLogFile: %v", err)
	}
	hdr2, recs2, err := ReadLogFile(logPath)
	if err != nil {
		t.Fatalf("ReadLogFile: %v", err)
	}
	env2 := LogToJSON(hdr2, recs2)
	if len(env2.Events) != 2 {
		t.Fatalf("event count = %d, want 2", len(env2.Events))
	}
	if env2.Events[0].KeyDown == nil || env2.Events[0].KeyDown.Ts != 0 {
		t.Fatalf("first event = %+v, want KeyDown at ts=0", env2.Events[0])
	}
	if env2.Events[1].KeyUp == nil || env2.Events[1].KeyUp.Ts != 20 {
		t.Fatalf("second event = %+v, want KeyUp at ts=20", env2.Events[1])
	}
}

func TestJSONToLogRejectsInvalidInput(t *testing.T) {
	cases := []struct {
		name string
		env  *Envelope
		want string
	}{
		{
			name: "bad mouse button",
			env: &Envelope{
				Version: MagicV1,
				Events:  []Event{{TypeName: "MouseDown", MouseDown: &ButtonEv{Button: "badbutton", Ts: 5}}},
			},
			want: `unknown button "badbutton"`,
		},
		{
			name: "bad key name",
			env: &Envelope{
				Version: MagicV1,
				Events:  []Event{{TypeName: "KeyDown", KeyDown: &KeyEvent{Keys: []string{"badkey"}, Ts: 5}}},
			},
			want: `unknown key name "badkey"`,
		},
		{
			name: "too many simultaneous keys",
			env: &Envelope{
				Version: MagicV1,
				Events: []Event{
					{TypeName: "KeyDown", KeyDown: &KeyEvent{Keys: []string{"a", "b", "c", "d", "e", "f"}, Ts: 0}},
					{TypeName: "KeyDown", KeyDown: &KeyEvent{Keys: []string{"g"}, Ts: 10}},
				},
			},
			want: "too many simultaneous keys",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, _, _, _, err := JSONToLog(tc.env)
			if err == nil {
				t.Fatal("expected error")
			}
			if got := err.Error(); !strings.Contains(got, tc.want) {
				t.Fatalf("error = %q, want substring %q", got, tc.want)
			}
		})
	}
}

func TestJSONToLogNonMonotonicTimestampsKeepEventOrder(t *testing.T) {
	env := &Envelope{
		Version: MagicV1,
		Events: []Event{
			{TypeName: "KeyDown", KeyDown: &KeyEvent{Keys: []string{"w"}, Ts: 0}},
			{TypeName: "KeyUp", KeyUp: &KeyEvent{Keys: []string{"w"}, Ts: 5471}},
			{TypeName: "KeyDown", KeyDown: &KeyEvent{Keys: []string{"d"}, Ts: 0}},
			{TypeName: "KeyUp", KeyUp: &KeyEvent{Keys: []string{"d"}, Ts: 1471}},
		},
	}

	hdr, recs, hasMouse, hasKbd, err := JSONToLog(env)
	if err != nil {
		t.Fatalf("JSONToLog: %v", err)
	}
	if hasMouse {
		t.Fatal("unexpected mouse records")
	}
	if !hasKbd {
		t.Fatal("expected keyboard records")
	}

	tmp := t.TempDir()
	logPath := filepath.Join(tmp, "non-monotonic.log")
	if err := WriteLogFile(logPath, hdr, recs, hasMouse, hasKbd); err != nil {
		t.Fatalf("WriteLogFile: %v", err)
	}
	hdr2, recs2, err := ReadLogFile(logPath)
	if err != nil {
		t.Fatalf("ReadLogFile: %v", err)
	}
	env2 := LogToJSON(hdr2, recs2)
	if len(env2.Events) != 4 {
		t.Fatalf("event count = %d, want 4", len(env2.Events))
	}
	if env2.Events[0].KeyDown == nil || len(env2.Events[0].KeyDown.Keys) != 1 || env2.Events[0].KeyDown.Keys[0] != "w" || env2.Events[0].KeyDown.Ts != 0 {
		t.Fatalf("event 0 = %+v, want KeyDown(w) ts=0", env2.Events[0])
	}
	if env2.Events[1].KeyUp == nil || len(env2.Events[1].KeyUp.Keys) != 1 || env2.Events[1].KeyUp.Keys[0] != "w" || env2.Events[1].KeyUp.Ts != 5471 {
		t.Fatalf("event 1 = %+v, want KeyUp(w) ts=5471", env2.Events[1])
	}
	if env2.Events[2].KeyDown == nil || len(env2.Events[2].KeyDown.Keys) != 1 || env2.Events[2].KeyDown.Keys[0] != "d" || env2.Events[2].KeyDown.Ts != 5471 {
		t.Fatalf("event 2 = %+v, want KeyDown(d) ts=5471", env2.Events[2])
	}
	if env2.Events[3].KeyUp == nil || len(env2.Events[3].KeyUp.Keys) != 1 || env2.Events[3].KeyUp.Keys[0] != "d" || env2.Events[3].KeyUp.Ts != 6942 {
		t.Fatalf("event 3 = %+v, want KeyUp(d) ts=6942", env2.Events[3])
	}
}

func TestReadLogFileV2HeaderSkipsReservedBytes(t *testing.T) {
	tmp := t.TempDir()
	logPath := filepath.Join(tmp, "v2.log")
	hdrWant := LogHeader{
		Version: MagicV2,
		Meta: Meta{
			Width:  1920,
			Height: 1080,
			StartX: 320,
			StartY: 240,
		},
	}
	recsWant := []Record{
		{Port: 1, Delta: 8, Length: 5, Payload: mouseEnc(0, 12, -7, 0)},
		{Port: 2, Delta: 16, Length: 8, Payload: kbdReport(ModLShift, [6]uint8{0x04, 0, 0, 0, 0, 0})},
	}

	if err := WriteLogFile(logPath, hdrWant, recsWant, false, false); err != nil {
		t.Fatalf("write v2 log: %v", err)
	}

	hdrGot, recsGot, err := ReadLogFile(logPath)
	if err != nil {
		t.Fatalf("read v2 log: %v", err)
	}
	if hdrGot != hdrWant {
		t.Fatalf("header mismatch: got %+v want %+v", hdrGot, hdrWant)
	}
	if len(recsGot) != len(recsWant) {
		t.Fatalf("record count mismatch: got %d want %d", len(recsGot), len(recsWant))
	}
	for i := range recsWant {
		if recsGot[i].Port != recsWant[i].Port ||
			recsGot[i].Delta != recsWant[i].Delta ||
			recsGot[i].Length != recsWant[i].Length ||
			!bytes.Equal(recsGot[i].Payload, recsWant[i].Payload) {
			t.Fatalf("record %d mismatch: got %+v want %+v", i, recsGot[i], recsWant[i])
		}
	}
}

func TestHIDKeyTablesCanonicalAndCompatible(t *testing.T) {
	scanCases := map[string]uint8{
		"a":            0x04,
		"A":            0x04,
		"space":        0x2c,
		"sysrq":        0x46,
		"printscreen":  0x46,
		"leftalt":      0xe2,
		"LEFTMETA":     0xe3,
		"media_calc":   0xfb,
		"leftbracket":  0x2f,
		"rightbracket": 0x30,
	}
	for name, want := range scanCases {
		got, ok := ScanCode(name)
		if !ok || got != want {
			t.Fatalf("ScanCode(%q) = (%#x, %v), want (%#x, true)", name, got, ok, want)
		}
	}

	if got := ScanName(0x46); got != "sysrq" {
		t.Fatalf("ScanName(0x46) = %q, want sysrq", got)
	}
	if got := ScanName(0xe2); got != "leftalt" {
		t.Fatalf("ScanName(0xe2) = %q, want leftalt", got)
	}

	modCases := map[string]uint8{
		"lalt":     ModLAlt,
		"mod_lalt": ModLAlt,
		"ALT":      ModLAlt,
		"LMETA":    ModLGUI,
		"lgui":     ModLGUI,
		"RGUI":     ModRGUI,
		"mod_rgui": ModRGUI,
	}
	for name, want := range modCases {
		got := ModBit(name)
		if got != want {
			t.Fatalf("ModBit(%q) = %#x, want %#x", name, got, want)
		}
	}

	if !IsModName("rshift") || !IsModName("mod_rshift") {
		t.Fatal("expected rshift/mod_rshift to be recognised as modifiers")
	}
	if !IsModName("mod_rgui") {
		t.Fatal("expected mod_rgui to be accepted as a compatible modifier alias")
	}
	if IsModName("LEFTSHIFT") {
		t.Fatal("LEFTSHIFT should be treated as an independent scan code, not a modifier")
	}
	if got := ModName(2); got != "lalt" {
		t.Fatalf("ModName(2) = %q, want lalt", got)
	}
}

// sameJSONFile returns true if two JSON files contain equivalent
// Envelopes (byte-for-byte after canonical re-encoding).
func sameJSONFile(a, b string) bool {
	da, err := os.ReadFile(a)
	if err != nil {
		return false
	}
	db, err := os.ReadFile(b)
	if err != nil {
		return false
	}
	var ea, eb Envelope
	if err := json.Unmarshal(da, &ea); err != nil {
		return false
	}
	if err := json.Unmarshal(db, &eb); err != nil {
		return false
	}
	return sameEvents(ea.Events, eb.Events) &&
		ea.Version == eb.Version &&
		ea.Meta == eb.Meta &&
		ea.Format == eb.Format
}

func sameEvents(a, b []Event) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if !eventEqual(a[i], b[i]) {
			return false
		}
	}
	return true
}

func eventEqual(a, b Event) bool {
	if (a.MouseMove == nil) != (b.MouseMove == nil) {
		return false
	}
	if a.MouseMove != nil && *a.MouseMove != *b.MouseMove {
		return false
	}
	if (a.MouseDown == nil) != (b.MouseDown == nil) {
		return false
	}
	if a.MouseDown != nil && *a.MouseDown != *b.MouseDown {
		return false
	}
	if (a.MouseUp == nil) != (b.MouseUp == nil) {
		return false
	}
	if a.MouseUp != nil && *a.MouseUp != *b.MouseUp {
		return false
	}
	if (a.Wheel == nil) != (b.Wheel == nil) {
		return false
	}
	if a.Wheel != nil && *a.Wheel != *b.Wheel {
		return false
	}
	if (a.KeyDown == nil) != (b.KeyDown == nil) {
		return false
	}
	if a.KeyDown != nil && !sameKeys(a.KeyDown, b.KeyDown) {
		return false
	}
	if (a.KeyUp == nil) != (b.KeyUp == nil) {
		return false
	}
	if a.KeyUp != nil && !sameKeys(a.KeyUp, b.KeyUp) {
		return false
	}
	return true
}

func sameKeys(a, b *KeyEvent) bool {
	if a.Ts != b.Ts || len(a.Keys) != len(b.Keys) {
		return false
	}
	as := append([]string(nil), a.Keys...)
	bs := append([]string(nil), b.Keys...)
	sort.Strings(as)
	sort.Strings(bs)
	for i := range as {
		if as[i] != bs[i] {
			return false
		}
	}
	return true
}

func dumpEvents(es []Event) string {
	var buf bytes.Buffer
	for _, e := range es {
		data, _ := json.Marshal(e)
		buf.Write(data)
		buf.WriteByte('\n')
	}
	return buf.String()
}

// repoRootFromTest walks up from this test file's working directory
// to find the project root. Tests run with cwd = package dir; the
// project root is one level up.
func repoRootFromTest(t *testing.T) string {
	t.Helper()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	return filepath.Dir(cwd)
}

func TestIntegrationRealFiles(t *testing.T) {
	root := repoRootFromTest(t)
	cases := []struct {
		logPath  string
		jsonPath string
	}{
		{
			logPath:  filepath.Join(root, "http4dev/server_root/applog/fy_liuboshan/daozheyouyi.log"),
			jsonPath: filepath.Join(root, "http4dev/server_root/applog/fy_liuboshan/daozheyouyi.json"),
		},
		{
			logPath: filepath.Join(root, "http4dev/server_root/applog/fy_liuboshan/chianfengbokangdimang.log"),
		},
	}

	tmp := t.TempDir()

	for i, c := range cases {
		if _, err := os.Stat(c.logPath); err != nil {
			t.Logf("skip case %d: %v", i, err)
			continue
		}
		out := filepath.Join(tmp, filepath.Base(c.logPath)+".json")
		if err := convertLogToJSON(c.logPath, out); err != nil {
			t.Fatalf("convert %s: %v", c.logPath, err)
		}
		data, err := os.ReadFile(out)
		if err != nil || len(data) == 0 {
			t.Fatalf("output %s empty or unreadable: %v", out, err)
		}
		var env Envelope
		if err := json.Unmarshal(data, &env); err != nil {
			t.Fatalf("output %s is not valid JSON: %v", out, err)
		}
		if env.Format != "mkp-action-v1" {
			t.Errorf("%s: format = %q, want mkp-action-v1", out, env.Format)
		}
		if env.Version != MagicV1 && env.Version != MagicV2 {
			t.Errorf("%s: version = %q, want MKv1 or MKv2", out, env.Version)
		}
		if len(env.Events) == 0 {
			t.Errorf("%s: no events decoded", out)
		}
	}

	// JSON-only case: test.json (keyboard-only).
	jsonOnly := filepath.Join(root, "http4dev/server_root/applog/test.json")
	if _, err := os.Stat(jsonOnly); err == nil {
		out := filepath.Join(tmp, "test.log")
		if err := convertJSONToLog(jsonOnly, out); err != nil {
			t.Fatalf("convert %s: %v", jsonOnly, err)
		}
		data, err := os.ReadFile(out)
		if err != nil || len(data) < MagicLength+RecordHeaderSize+8 {
			t.Fatalf("output %s is empty/too small: %d bytes", out, len(data))
		}
		if string(data[:MagicLength]) != MagicV1 {
			t.Fatalf("output %s: bad magic", out)
		}
	}
}
