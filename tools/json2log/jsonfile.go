// JSON envelope for applog (matches the firmware's adumj / ajson2log
// output, see docs/cli-adumj.md).
//
//	{
//	  "format":  "mkp-action-v1",
//	  "version": "MKv1" | "MKv2",
//	  "meta":    { "width": N, "height": N, "startX": N, "startY": N },
//	  "events":  [ { "MouseMove": {...} }, ... ]
//	}
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// Envelope is the top-level JSON object. Only Format is hard-coded —
// the rest is filled in from the log file (forward) or from the JSON
// itself (reverse).
type Envelope struct {
	Format  string   `json:"format"`
	Version string   `json:"version"`
	Meta    MetaJSON `json:"meta"`
	Events  []Event  `json:"events"`
}

// MetaJSON is the JSON form of the V2 file header metadata. startX
// and startY are signed here (int16_t) even though the binary file
// stores them as uint16, matching the firmware adumj output.
type MetaJSON struct {
	Width  uint16 `json:"width"`
	Height uint16 `json:"height"`
	StartX int16  `json:"startX"`
	StartY int16  `json:"startY"`
}

// Event is a single tagged event. Exactly one of the inner pointers
// is non-nil. We use a custom UnmarshalJSON so the JSON form stays
// the same as the firmware (a single-key object).
type Event struct {
	// exactly one of these is set after parsing
	MouseMove *MouseMoveEv `json:"-"`
	MouseDown *ButtonEv    `json:"-"`
	MouseUp   *ButtonEv    `json:"-"`
	Wheel     *WheelEv     `json:"-"`
	KeyDown   *KeyEvent    `json:"-"`
	KeyUp     *KeyEvent    `json:"-"`
	// TypeName is the JSON key (e.g. "MouseMove"). Kept for
	// lossless re-serialization when the input had an unknown
	// event we want to forward verbatim.
	TypeName string `json:"-"`
}

// MouseMoveEv is the payload of a "MouseMove" event.
type MouseMoveEv struct {
	X  int16 `json:"x"`
	Y  int16 `json:"y"`
	Ts int32 `json:"ts"`
}

// ButtonEv is the payload of "MouseDown" / "MouseUp".
type ButtonEv struct {
	Button string `json:"button"` // "left" | "right" | "middle"
	Ts     int32  `json:"ts"`
}

// WheelEv is the payload of "Wheel".
type WheelEv struct {
	Delta int8  `json:"delta"`
	Ts    int32 `json:"ts"`
}

// KeyEvent is the payload of "KeyDown" / "KeyUp".
type KeyEvent struct {
	Keys []string `json:"keys"`
	Ts   int32    `json:"ts"`
}

// rawEvent is used for the custom unmarshaler. We need to peek at
// the single key to know which typed struct to decode into.
type rawEvent map[string]json.RawMessage

// UnmarshalJSON decodes a single-key event object.
func (e *Event) UnmarshalJSON(data []byte) error {
	var r rawEvent
	if err := json.Unmarshal(data, &r); err != nil {
		return err
	}
	if len(r) != 1 {
		return fmt.Errorf("event must have exactly one key, got %d", len(r))
	}
	for k, v := range r {
		e.TypeName = k
		var err error
		switch k {
		case "MouseMove":
			e.MouseMove = &MouseMoveEv{}
			err = json.Unmarshal(v, e.MouseMove)
		case "MouseDown":
			e.MouseDown = &ButtonEv{}
			err = json.Unmarshal(v, e.MouseDown)
		case "MouseUp":
			e.MouseUp = &ButtonEv{}
			err = json.Unmarshal(v, e.MouseUp)
		case "Wheel":
			e.Wheel = &WheelEv{}
			err = json.Unmarshal(v, e.Wheel)
		case "KeyDown":
			e.KeyDown = &KeyEvent{}
			err = json.Unmarshal(v, e.KeyDown)
		case "KeyUp":
			e.KeyUp = &KeyEvent{}
			err = json.Unmarshal(v, e.KeyUp)
		default:
			return fmt.Errorf("unknown event type %q", k)
		}
		return err
	}
	return nil
}

// MarshalJSON encodes the event as a single-key object.
func (e Event) MarshalJSON() ([]byte, error) {
	switch {
	case e.MouseMove != nil:
		return json.Marshal(map[string]*MouseMoveEv{e.TypeNameOr("MouseMove"): e.MouseMove})
	case e.MouseDown != nil:
		return json.Marshal(map[string]*ButtonEv{e.TypeNameOr("MouseDown"): e.MouseDown})
	case e.MouseUp != nil:
		return json.Marshal(map[string]*ButtonEv{e.TypeNameOr("MouseUp"): e.MouseUp})
	case e.Wheel != nil:
		return json.Marshal(map[string]*WheelEv{e.TypeNameOr("Wheel"): e.Wheel})
	case e.KeyDown != nil:
		return json.Marshal(map[string]*KeyEvent{e.TypeNameOr("KeyDown"): e.KeyDown})
	case e.KeyUp != nil:
		return json.Marshal(map[string]*KeyEvent{e.TypeNameOr("KeyUp"): e.KeyUp})
	}
	return nil, fmt.Errorf("empty event")
}

func (e Event) TypeNameOr(def string) string {
	if e.TypeName != "" {
		return e.TypeName
	}
	return def
}

// ReadJSONFile loads the envelope from path.
func ReadJSONFile(path string) (*Envelope, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var env Envelope
	if err := json.Unmarshal(data, &env); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	if env.Format == "" {
		env.Format = "mkp-action-v1"
	}
	if env.Version == "" {
		env.Version = MagicV1
	}
	return &env, nil
}

// WriteJSONFile marshals the envelope to path using the firmware's
// compact format (one event per line, 4-space indent, ",\n    "
// separator). This matches the byte layout produced by the C
// reference adumj command (see docs/cli-adumj.md).
func WriteJSONFile(path string, env *Envelope) error {
	return os.WriteFile(path, []byte(formatEnvelope(env)), 0o644)
}

// formatEnvelope renders the envelope in the firmware's adumj style.
//
//	{
//	  "format": "mkp-action-v1",
//	  "version": "MKv1",
//	  "meta": { "width": 0, "height": 0, "startX": 0, "startY": 0 },
//	  "events": [
//	    { "KeyDown":    { "keys": ["w"], "ts": 1518 } }
//	    ,    { "KeyUp":   { "keys": ["w"], "ts": 2633 } }
//	    ,    { "MouseMove": { "x": -1, "y": 0, "ts": 42880 } }
//	    ...
//	  ]
//	}
//
// The alignment of the inner-struct keys after the event type is the
// same quirky layout the C code emits (extra spaces to make the
// payload column line up across event types of different key
// lengths). Cosmetic only; the JSON parser doesn't care.
func formatEnvelope(env *Envelope) string {
	var b bytes.Buffer
	b.WriteString("{\n")
	fmt.Fprintf(&b, "  \"format\": \"mkp-action-v1\",\n")
	fmt.Fprintf(&b, "  \"version\": \"%s\",\n", env.Version)
	fmt.Fprintf(&b, "  \"meta\": { \"width\": %d, \"height\": %d, \"startX\": %d, \"startY\": %d },\n",
		env.Meta.Width, env.Meta.Height, env.Meta.StartX, env.Meta.StartY)
	b.WriteString("  \"events\": [\n")
	for i, e := range env.Events {
		if i > 0 {
			b.WriteString("    ,\n")
		}
		b.WriteString("    ")
		b.WriteString(formatEvent(e))
		b.WriteByte('\n')
	}
	b.WriteString("  ]\n}\n")
	return b.String()
}

// formatEvent renders one event as a single-line single-key object.
func formatEvent(e Event) string {
	// The width after the colon is chosen so the payload column
	// lines up across the event-type names "KeyDown", "KeyUp",
	// "MouseMove" (8 + 2 = 10), "MouseDown" (9 + 1 = 10), "MouseUp"
	// (7 + 3 = 10), "Wheel" (5 + 5 = 10). The longest type is
	// "MouseDown" / "MouseMove" at 9 chars; we pad to 10 to leave
	// one space before the inner object.
	const innerAlign = 10

	pad := func(typeName string) string {
		if len(typeName) >= innerAlign {
			return " "
		}
		return strings.Repeat(" ", innerAlign-len(typeName))
	}

	switch {
	case e.MouseMove != nil:
		return fmt.Sprintf("{ \"MouseMove\":%s{ \"x\": %d, \"y\": %d, \"ts\": %d } }",
			pad("MouseMove"), e.MouseMove.X, e.MouseMove.Y, e.MouseMove.Ts)
	case e.MouseDown != nil:
		return fmt.Sprintf("{ \"MouseDown\":%s{ \"button\": \"%s\", \"ts\": %d } }",
			pad("MouseDown"), e.MouseDown.Button, e.MouseDown.Ts)
	case e.MouseUp != nil:
		return fmt.Sprintf("{ \"MouseUp\":%s{ \"button\": \"%s\", \"ts\": %d } }",
			pad("MouseUp"), e.MouseUp.Button, e.MouseUp.Ts)
	case e.Wheel != nil:
		return fmt.Sprintf("{ \"Wheel\":%s{ \"delta\": %d, \"ts\": %d } }",
			pad("Wheel"), e.Wheel.Delta, e.Wheel.Ts)
	case e.KeyDown != nil:
		return fmt.Sprintf("{ \"KeyDown\":%s{ \"keys\": [%s], \"ts\": %d } }",
			pad("KeyDown"), formatKeyList(e.KeyDown.Keys), e.KeyDown.Ts)
	case e.KeyUp != nil:
		return fmt.Sprintf("{ \"KeyUp\":%s{ \"keys\": [%s], \"ts\": %d } }",
			pad("KeyUp"), formatKeyList(e.KeyUp.Keys), e.KeyUp.Ts)
	}
	return "{}"
}

func formatKeyList(keys []string) string {
	quoted := make([]string, len(keys))
	for i, k := range keys {
		quoted[i] = "\"" + k + "\""
	}
	return strings.Join(quoted, ", ")
}
