// Conversion logic between the binary .log and the JSON envelope.
//
// Both conversions reproduce the firmware's state-machine behaviour
// exactly, so log -> JSON -> log is replay-equivalent (see
// docs/tool-log2json.md "Reversibility").
package converter

import (
	"fmt"
)

// KbdMaxKeys mirrors the HID Boot Keyboard report's 6-slot key array.
const KbdMaxKeys = 6

// mouseStateKbdState are the running state carried across records/events.
type mouseState struct {
	prevButton uint8
}

type kbdState struct {
	prevModifier uint8
	prevKeys     [KbdMaxKeys]uint8
}

// LogToJSON converts a binary log file to the JSON envelope form.
func LogToJSON(hdr LogHeader, recs []Record) *Envelope {
	env := &Envelope{
		Format:  "mkp-action-v1",
		Version: hdr.Version,
		Meta: MetaJSON{
			Width:  hdr.Meta.Width,
			Height: hdr.Meta.Height,
			StartX: int16(hdr.Meta.StartX),
			StartY: int16(hdr.Meta.StartY),
		},
	}

	mst := mouseState{}
	kst := kbdState{}
	var ts int32
	for _, rec := range recs {
		ts += rec.Delta
		switch {
		case rec.Port == 1 && rec.Length == 5:
			var p [5]byte
			copy(p[:], rec.Payload)
			r := MouseDecode(p)
			emitMouseEvents(env, r, &mst, ts)
		case rec.Port == 2 && rec.Length == 8:
			if len(rec.Payload) < 8 {
				continue
			}
			mod := rec.Payload[0]
			var keys [KbdMaxKeys]uint8
			copy(keys[:], rec.Payload[2:2+KbdMaxKeys])
			emitKbdEvents(env, mod, keys, &kst, ts)
		}
	}
	return env
}

// emitMouseEvents mirrors emit_mouse_events in
// player/main/cmd_logfile_adump.c.
func emitMouseEvents(env *Envelope, r MouseReport, st *mouseState, ts int32) {
	if r.X != 0 || r.Y != 0 {
		env.Events = append(env.Events, Event{
			TypeName:  "MouseMove",
			MouseMove: &MouseMoveEv{X: r.X, Y: r.Y, Ts: ts},
		})
	}
	down := r.Button &^ st.prevButton
	up := st.prevButton &^ r.Button
	for _, bit := range []uint8{MouseButtonLeft, MouseButtonRight, MouseButtonMiddle} {
		if name := ButtonName(bit); name != "" {
			if down&bit != 0 {
				env.Events = append(env.Events, Event{
					TypeName:  "MouseDown",
					MouseDown: &ButtonEv{Button: name, Ts: ts},
				})
			}
			if up&bit != 0 {
				env.Events = append(env.Events, Event{
					TypeName: "MouseUp",
					MouseUp:  &ButtonEv{Button: name, Ts: ts},
				})
			}
		}
	}
	if r.Wheel != 0 {
		env.Events = append(env.Events, Event{
			TypeName: "Wheel",
			Wheel:    &WheelEv{Delta: r.Wheel, Ts: ts},
		})
	}
	st.prevButton = r.Button
}

// emitKbdEvents mirrors emit_kbd_events in
// player/main/cmd_logfile_adump.c.
func emitKbdEvents(env *Envelope, modifier uint8, keys [KbdMaxKeys]uint8, st *kbdState, ts int32) {
	modDown := (modifier ^ st.prevModifier) & modifier
	modUp := (modifier ^ st.prevModifier) & st.prevModifier

	var downNames, upNames []string
	for b := uint8(0); b < 8; b++ {
		if modDown&(1<<b) != 0 {
			if n := ModName(b); n != "" {
				downNames = append(downNames, n)
			}
		}
		if modUp&(1<<b) != 0 {
			if n := ModName(b); n != "" {
				upNames = append(upNames, n)
			}
		}
	}
	for _, k := range keys {
		if k == 0 {
			continue
		}
		if !containsKey(st.prevKeys[:], k) {
			downNames = append(downNames, ScanName(k))
		}
	}
	for _, k := range st.prevKeys {
		if k == 0 {
			continue
		}
		if !containsKey(keys[:], k) {
			upNames = append(upNames, ScanName(k))
		}
	}
	if len(downNames) > 0 {
		env.Events = append(env.Events, Event{
			TypeName: "KeyDown",
			KeyDown:  &KeyEvent{Keys: downNames, Ts: ts},
		})
	}
	if len(upNames) > 0 {
		env.Events = append(env.Events, Event{
			TypeName: "KeyUp",
			KeyUp:    &KeyEvent{Keys: upNames, Ts: ts},
		})
	}
	st.prevModifier = modifier
	st.prevKeys = keys
}

func containsKey(arr []uint8, v uint8) bool {
	for _, x := range arr {
		if x == v {
			return true
		}
	}
	return false
}

// JSONToLog converts a JSON envelope to a sequence of records plus the
// needed stop trailer flags. Events are consumed in their existing JSON
// order. If timestamps go backwards, the binary log cannot represent a
// negative delta, so that step is clamped to zero.
func JSONToLog(env *Envelope) (LogHeader, []Record, bool, bool, error) {
	events := append([]Event(nil), env.Events...)

	hdr := LogHeader{
		Version: env.Version,
		Meta: Meta{
			Width:  env.Meta.Width,
			Height: env.Meta.Height,
			StartX: uint16(env.Meta.StartX),
			StartY: uint16(env.Meta.StartY),
		},
	}
	if hdr.Version != MagicV1 && hdr.Version != MagicV2 {
		hdr.Version = MagicV1
	}

	var recs []Record
	var curButton uint8
	var kst kbdState
	prevTS := int32(0)
	first := true
	hasMouse, hasKbd := false, false

	for i, e := range events {
		ts := eventTS(e)
		delta := int32(0)
		if first {
			delta = ts
			first = false
		} else {
			delta = ts - prevTS
			if delta < 0 {
				delta = 0
			}
		}
		prevTS = ts

		switch {
		case e.MouseMove != nil:
			p := MouseEncode(curButton, e.MouseMove.X, e.MouseMove.Y, 0)
			recs = append(recs, Record{Port: 1, Delta: delta, Length: 5, Payload: p[:]})
			hasMouse = true
		case e.MouseDown != nil:
			bit, ok := ButtonBit(e.MouseDown.Button)
			if !ok {
				return LogHeader{}, nil, false, false, fmt.Errorf("event %d MouseDown: unknown button %q", i, e.MouseDown.Button)
			}
			curButton |= bit
			p := MouseEncode(curButton, 0, 0, 0)
			recs = append(recs, Record{Port: 1, Delta: delta, Length: 5, Payload: p[:]})
			hasMouse = true
		case e.MouseUp != nil:
			bit, ok := ButtonBit(e.MouseUp.Button)
			if !ok {
				return LogHeader{}, nil, false, false, fmt.Errorf("event %d MouseUp: unknown button %q", i, e.MouseUp.Button)
			}
			curButton &^= bit
			p := MouseEncode(curButton, 0, 0, 0)
			recs = append(recs, Record{Port: 1, Delta: delta, Length: 5, Payload: p[:]})
			hasMouse = true
		case e.Wheel != nil:
			p := MouseEncode(curButton, 0, 0, e.Wheel.Delta)
			recs = append(recs, Record{Port: 1, Delta: delta, Length: 5, Payload: p[:]})
			hasMouse = true
		case e.KeyDown != nil:
			if err := applyKbdNames(&kst, e.KeyDown.Keys, true); err != nil {
				return LogHeader{}, nil, false, false, fmt.Errorf("event %d KeyDown: %w", i, err)
			}
			report := buildKbdReport(&kst)
			recs = append(recs, Record{Port: 2, Delta: delta, Length: 8, Payload: report[:]})
			hasKbd = true
		case e.KeyUp != nil:
			if err := applyKbdNames(&kst, e.KeyUp.Keys, false); err != nil {
				return LogHeader{}, nil, false, false, fmt.Errorf("event %d KeyUp: %w", i, err)
			}
			report := buildKbdReport(&kst)
			recs = append(recs, Record{Port: 2, Delta: delta, Length: 8, Payload: report[:]})
			hasKbd = true
		}
	}
	return hdr, recs, hasMouse, hasKbd, nil
}

// eventTS returns the timestamp of an event, or 0 if the event is
// missing its payload (which we treat as a no-op).
func eventTS(e Event) int32 {
	switch {
	case e.MouseMove != nil:
		return e.MouseMove.Ts
	case e.MouseDown != nil:
		return e.MouseDown.Ts
	case e.MouseUp != nil:
		return e.MouseUp.Ts
	case e.Wheel != nil:
		return e.Wheel.Ts
	case e.KeyDown != nil:
		return e.KeyDown.Ts
	case e.KeyUp != nil:
		return e.KeyUp.Ts
	}
	return 0
}

// applyKbdNames is the reverse-direction equivalent of the modifier
// and key handling in emitKbdEvents: it inserts or removes each name
// from the running kbdState.
func applyKbdNames(st *kbdState, names []string, isDown bool) error {
	for _, name := range names {
		if IsModName(name) {
			bit := ModBit(name)
			if isDown {
				st.prevModifier |= bit
			} else {
				st.prevModifier &^= bit
			}
			continue
		}
		sc, ok := ScanCode(name)
		if !ok || sc == 0 {
			return fmt.Errorf("unknown key name %q", name)
		}
		if isDown {
			if containsKey(st.prevKeys[:], sc) {
				continue
			}
			for i := 0; i < KbdMaxKeys; i++ {
				if st.prevKeys[i] == 0 {
					st.prevKeys[i] = sc
					sc = 0
					break
				}
			}
			if sc != 0 {
				return fmt.Errorf("too many simultaneous keys (max %d)", KbdMaxKeys)
			}
		} else {
			for i := 0; i < KbdMaxKeys; i++ {
				if st.prevKeys[i] == sc {
					st.prevKeys[i] = 0
					break
				}
			}
		}
	}
	return nil
}

// buildKbdReport assembles an 8-byte HID Boot Keyboard report from the
// running state.
func buildKbdReport(st *kbdState) [8]byte {
	var r [8]byte
	r[0] = st.prevModifier
	r[1] = 0 // reserved
	for i := 0; i < KbdMaxKeys; i++ {
		r[2+i] = st.prevKeys[i]
	}
	return r
}
