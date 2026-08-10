// HID scan code <-> name tables.
//
// Canonical names mirror the root package table in
// ../../types_keycode_maps.go so log2json uses the same HID vocabulary
// as the main mkp-go package.
package converter

import "strings"

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

type hidKeySpec struct {
	name string
	code uint8
}

// hidKeySpecs is the canonical non-modifier HID key table mirrored
// from types_keycode_maps.go.
var hidKeySpecs = []hidKeySpec{
	{"none", 0x00},
	{"err_ovf", 0x01},

	{"a", 0x04}, {"b", 0x05}, {"c", 0x06}, {"d", 0x07}, {"e", 0x08},
	{"f", 0x09}, {"g", 0x0a}, {"h", 0x0b}, {"i", 0x0c}, {"j", 0x0d},
	{"k", 0x0e}, {"l", 0x0f}, {"m", 0x10}, {"n", 0x11}, {"o", 0x12},
	{"p", 0x13}, {"q", 0x14}, {"r", 0x15}, {"s", 0x16}, {"t", 0x17},
	{"u", 0x18}, {"v", 0x19}, {"w", 0x1a}, {"x", 0x1b}, {"y", 0x1c},
	{"z", 0x1d},

	{"1", 0x1e}, {"2", 0x1f}, {"3", 0x20}, {"4", 0x21}, {"5", 0x22},
	{"6", 0x23}, {"7", 0x24}, {"8", 0x25}, {"9", 0x26}, {"0", 0x27},

	{"enter", 0x28},
	{"esc", 0x29},
	{"backspace", 0x2a},
	{"tab", 0x2b},
	{"space", 0x2c},
	{"minus", 0x2d},
	{"equal", 0x2e},
	{"leftbrace", 0x2f},
	{"rightbrace", 0x30},
	{"backslash", 0x31},
	{"hashtilde", 0x32},
	{"semicolon", 0x33},
	{"apostrophe", 0x34},
	{"grave", 0x35},
	{"comma", 0x36},
	{"dot", 0x37},
	{"slash", 0x38},
	{"capslock", 0x39},

	{"f1", 0x3a}, {"f2", 0x3b}, {"f3", 0x3c}, {"f4", 0x3d},
	{"f5", 0x3e}, {"f6", 0x3f}, {"f7", 0x40}, {"f8", 0x41},
	{"f9", 0x42}, {"f10", 0x43}, {"f11", 0x44}, {"f12", 0x45},

	{"sysrq", 0x46},
	{"scrolllock", 0x47},
	{"pause", 0x48},
	{"insert", 0x49},
	{"home", 0x4a},
	{"pageup", 0x4b},
	{"delete", 0x4c},
	{"end", 0x4d},
	{"pagedown", 0x4e},
	{"right", 0x4f},
	{"left", 0x50},
	{"down", 0x51},
	{"up", 0x52},

	{"numlock", 0x53},
	{"kpslash", 0x54},
	{"kpasterisk", 0x55},
	{"kpminus", 0x56},
	{"kpplus", 0x57},
	{"kpenter", 0x58},
	{"kp1", 0x59},
	{"kp2", 0x5a},
	{"kp3", 0x5b},
	{"kp4", 0x5c},
	{"kp5", 0x5d},
	{"kp6", 0x5e},
	{"kp7", 0x5f},
	{"kp8", 0x60},
	{"kp9", 0x61},
	{"kp0", 0x62},
	{"kpdot", 0x63},

	{"102nd", 0x64},
	{"compose", 0x65},
	{"power", 0x66},
	{"kpequal", 0x67},
	{"f13", 0x68}, {"f14", 0x69}, {"f15", 0x6a}, {"f16", 0x6b},
	{"f17", 0x6c}, {"f18", 0x6d}, {"f19", 0x6e}, {"f20", 0x6f},
	{"f21", 0x70}, {"f22", 0x71}, {"f23", 0x72}, {"f24", 0x73},

	{"open", 0x74},
	{"help", 0x75},
	{"props", 0x76},
	{"front", 0x77},
	{"stop", 0x78},
	{"again", 0x79},
	{"undo", 0x7a},
	{"cut", 0x7b},
	{"copy", 0x7c},
	{"paste", 0x7d},
	{"find", 0x7e},
	{"mute", 0x7f},
	{"volumeup", 0x80},
	{"volumedown", 0x81},

	{"kpcomma", 0x85},
	{"ro", 0x87},
	{"katakanahiragana", 0x88},
	{"yen", 0x89},
	{"henkan", 0x8a},
	{"muhenkan", 0x8b},
	{"kpjpcomma", 0x8c},

	{"hangeul", 0x90},
	{"hanja", 0x91},
	{"katakana", 0x92},
	{"hiragana", 0x93},
	{"zenkakuhankaku", 0x94},

	{"kpleftparen", 0xb6},
	{"kprightparen", 0xb7},

	{"leftctrl", 0xe0},
	{"leftshift", 0xe1},
	{"leftalt", 0xe2},
	{"leftmeta", 0xe3},
	{"rightctrl", 0xe4},
	{"rightshift", 0xe5},
	{"rightalt", 0xe6},
	{"rightmeta", 0xe7},

	{"media_playpause", 0xe8},
	{"media_stopcd", 0xe9},
	{"media_previoussong", 0xea},
	{"media_nextsong", 0xeb},
	{"media_ejectcd", 0xec},
	{"media_volumeup", 0xed},
	{"media_volumedown", 0xee},
	{"media_mute", 0xef},
	{"media_www", 0xf0},
	{"media_back", 0xf1},
	{"media_forward", 0xf2},
	{"media_stop", 0xf3},
	{"media_find", 0xf4},
	{"media_scrollup", 0xf5},
	{"media_scrolldown", 0xf6},
	{"media_edit", 0xf7},
	{"media_sleep", 0xf8},
	{"media_coffee", 0xf9},
	{"media_refresh", 0xfa},
	{"media_calc", 0xfb},
}

// scanNameAliases keeps old/json historical names working while the
// canonical output now follows types_keycode_maps.go.
var scanNameAliases = map[string]string{
	"printscreen":  "sysrq",
	"leftbracket":  "leftbrace",
	"rightbracket": "rightbrace",
}

// nameToScan maps every supported HID key name to its scan code.
var nameToScan = func() map[string]uint8 {
	m := make(map[string]uint8, len(hidKeySpecs))
	for _, spec := range hidKeySpecs {
		m[spec.name] = spec.code
	}
	return m
}()

// scanToName is the reverse map, precomputed once.
var scanToName = func() map[uint8]string {
	m := make(map[uint8]string, len(nameToScan))
	for name, code := range nameToScan {
		m[code] = name
	}
	return m
}()

func normalizeScanName(name string) string {
	key := strings.ToLower(strings.TrimSpace(name))
	if alias, ok := scanNameAliases[key]; ok {
		return alias
	}
	return key
}

// ScanCode returns the HID scan code for a name from nameToScan, or
// false if the name is unknown.
func ScanCode(name string) (uint8, bool) {
	c, ok := nameToScan[normalizeScanName(name)]
	return c, ok
}

// ScanName returns the canonical lowercase HID name for a scan code, or "" if unknown.
func ScanName(code uint8) string {
	return scanToName[code]
}

var canonicalNameToModBit = map[string]uint8{
	"lctrl": ModLCtrl, "lshift": ModLShift, "lalt": ModLAlt, "lmeta": ModLGUI,
	"rctrl": ModRCtrl, "rshift": ModRShift, "ralt": ModRAlt, "rmeta": ModRGUI,
}

// modNameAliases keeps helper/tool historical spellings working.
var modNameAliases = map[string]string{
	"ctrl": "lctrl", "shift": "lshift", "alt": "lalt", "meta": "lmeta",
	"lgui": "lmeta", "rgui": "rmeta",
}

func normalizeModName(name string) string {
	key := strings.ToLower(strings.TrimSpace(name))
	if after, ok := strings.CutPrefix(key, "mod_"); ok {
		key = after
	}
	if alias, ok := modNameAliases[key]; ok {
		return alias
	}
	return key
}

// ModBit returns the modifier bit for a name, or 0 if the name is not
// a modifier. (0 is also a valid "no modifier" answer; callers should
// first check the boolean returned by IsModName.)
func ModBit(name string) uint8 {
	return canonicalNameToModBit[normalizeModName(name)]
}

// IsModName reports whether name is a modifier key name.
func IsModName(name string) bool {
	_, ok := canonicalNameToModBit[normalizeModName(name)]
	return ok
}

// modBitToName is the canonical inverse map for the forward converter.
var modBitToName = [8]string{
	0: "lctrl", 1: "lshift", 2: "lalt", 3: "lmeta",
	4: "rctrl", 5: "rshift", 6: "ralt", 7: "rmeta",
}

// ModName returns the canonical lowercase name for a modifier bit, or "" if out of range.
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
	b, ok := nameToButtonBit[strings.ToLower(strings.TrimSpace(name))]
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
