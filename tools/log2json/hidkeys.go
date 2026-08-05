// HID scan code <-> name tables.
//
// These tables are the Go mirror of the C tables in
// player/main/cmd_logfile_adump.c (hidkeycode_to_name) and
// player/main/cmd_logfile_ajson2log.c (keyname_to_scancode,
// modifier_name_to_bit, button_name_to_bit).
package main

// Modifier bit positions in the HID keyboard modifier byte.
const (
	ModLCtrl  = 0x01
	ModLShift = 0x02
	ModLAlt   = 0x04
	ModLGUI   = 0x08
	ModRCtrl  = 0x10
	ModRShift = 0x20
	ModRAlt   = 0x40
	ModRGUI   = 0x80
)

// nameToScan maps every HID key name that the firmware emits to its
// scan code. The firmware never emits anything outside this set, so an
// unknown name is treated as a conversion error.
var nameToScan = map[string]uint8{
	// letters a..z (0x04-0x1d)
	"a": 0x04, "b": 0x05, "c": 0x06, "d": 0x07, "e": 0x08,
	"f": 0x09, "g": 0x0a, "h": 0x0b, "i": 0x0c, "j": 0x0d,
	"k": 0x0e, "l": 0x0f, "m": 0x10, "n": 0x11, "o": 0x12,
	"p": 0x13, "q": 0x14, "r": 0x15, "s": 0x16, "t": 0x17,
	"u": 0x18, "v": 0x19, "w": 0x1a, "x": 0x1b, "y": 0x1c,
	"z": 0x1d,

	// digits 1..9, 0 (0x1e-0x27)
	"1": 0x1e, "2": 0x1f, "3": 0x20, "4": 0x21, "5": 0x22,
	"6": 0x23, "7": 0x24, "8": 0x25, "9": 0x26, "0": 0x27,

	"enter":      0x28,
	"esc":        0x29,
	"backspace":  0x2a,
	"tab":        0x2b,
	"space":      0x2c,
	"minus":      0x2d,
	"equal":      0x2e,
	"leftbrace":  0x2f,
	"rightbrace": 0x30,
	"backslash":  0x31,
	"semicolon":  0x33,
	"apostrophe": 0x34,
	"grave":      0x35,
	"comma":      0x36,
	"dot":        0x37,
	"slash":      0x38,

	"capslock": 0x39,

	"f1":  0x3a,
	"f2":  0x3b,
	"f3":  0x3c,
	"f4":  0x3d,
	"f5":  0x3e,
	"f6":  0x3f,
	"f7":  0x40,
	"f8":  0x41,
	"f9":  0x42,
	"f10": 0x43,
	"f11": 0x44,
	"f12": 0x45,

	"printscreen": 0x46,
	"scrolllock":  0x47,
	"pause":       0x48,
	"insert":      0x49,
	"home":        0x4a,
	"pageup":      0x4b,
	"delete":      0x4c,
	"end":         0x4d,
	"pagedown":    0x4e,
	"right":       0x4f,
	"left":        0x50,
	"down":        0x51,
	"up":          0x52,
	"numlock":     0x53,

	"lctrl":  0xe0,
	"lshift": 0xe1,
	"lalt":   0xe2,
	"lgui":   0xe3,
	"rctrl":  0xe4,
	"rshift": 0xe5,
	"ralt":   0xe6,
	"rgui":   0xe7,
}

// scanToName is the reverse map, precomputed once.
var scanToName = func() map[uint8]string {
	m := make(map[uint8]string, len(nameToScan))
	for n, c := range nameToScan {
		m[c] = n
	}
	return m
}()

// ScanCode returns the HID scan code for a name from nameToScan, or
// false if the name is unknown.
func ScanCode(name string) (uint8, bool) {
	c, ok := nameToScan[name]
	return c, ok
}

// ScanName returns the HID name for a scan code, or "" if unknown.
func ScanName(code uint8) string {
	return scanToName[code]
}

// nameToModBit maps modifier names to their bit in the modifier byte.
var nameToModBit = map[string]uint8{
	"lctrl":  ModLCtrl,
	"lshift": ModLShift,
	"lalt":   ModLAlt,
	"lgui":   ModLGUI,
	"rctrl":  ModRCtrl,
	"rshift": ModRShift,
	"ralt":   ModRAlt,
	"rgui":   ModRGUI,
}

// ModBit returns the modifier bit for a name, or 0 if the name is not
// a modifier. (0 is also a valid "no modifier" answer; callers should
// first check the boolean returned by IsModName.)
func ModBit(name string) uint8 {
	return nameToModBit[name]
}

// IsModName reports whether name is a modifier key name.
func IsModName(name string) bool {
	_, ok := nameToModBit[name]
	return ok
}

// modBitToName is the inverse of nameToModBit for the forward converter.
var modBitToName = [8]string{
	0: "lctrl", 1: "lshift", 2: "lalt", 3: "lgui",
	4: "rctrl", 5: "rshift", 6: "ralt", 7: "rgui",
}

// ModName returns the name for a modifier bit, or "" if out of range.
func ModName(bit uint8) string {
	if bit >= 8 {
		return ""
	}
	return modBitToName[bit]
}

// nameToButtonBit maps mouse button names to the button bits.
var nameToButtonBit = map[string]uint8{
	"left":   MouseButtonLeft,
	"right":  MouseButtonRight,
	"middle": MouseButtonMiddle,
}

// ButtonBit returns the bit for a mouse button name, or 0 if unknown.
func ButtonBit(name string) (uint8, bool) {
	b, ok := nameToButtonBit[name]
	return b, ok
}

// ButtonName returns the name for a single button bit, or "" if not
// a recognised bit.
func ButtonName(bit uint8) string {
	switch bit {
	case MouseButtonLeft:
		return "left"
	case MouseButtonRight:
		return "right"
	case MouseButtonMiddle:
		return "middle"
	}
	return ""
}
