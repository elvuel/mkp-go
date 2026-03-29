package mkpgo

import (
	"fmt"
	"strings"

	"github.com/duke-git/lancet/v2/slice"
)

var (
	ModKeyNames      KpadModKeys = []string{"MOD_LCTRL", "MOD_LSHIFT", "MOD_LALT", "MOD_LMETA", "MOD_RCTRL", "MOD_RSHIFT", "MOD_RALT", "MOD_RMETA"}
	NodKeyShortNames             = []string{"LCTRL", "LSHIFT", "LALT", "LMETA", "RCTRL", "RSHIFT", "RALT", "RMETA"}
)

type KpadModKeys []string

func (m KpadModKeys) ToStatus() string {
	if len(m) == 0 {
		return ""
	}

	code := KeyNameToNCode[m[0]]
	resetKeys := m[1:]

	for _, key := range resetKeys {
		code = code | KeyNameToNCode[key]
	}

	return fmt.Sprintf("0x%02x", code)
}

func (m KpadModKeys) Contains(keys []string) bool {
	for _, key := range keys {
		if slice.Contain(m, key) {
			return true
		}
	}
	return false
}

func (m KpadModKeys) Extract(keys []string) (modKeys []string, normalKeys []string) {
	for _, key := range keys {
		if slice.Contain(m, key) {
			modKeys = append(modKeys, key)
		} else {
			normalKeys = append(normalKeys, key)
		}
	}
	return
}

// 键盘的cancel, 就是 发送 kpad --port 2 --rel 0 这样默认就是发送 全0的8个字节。

type KpadOption struct {
	// MOD_LCTRL 0x01 MOD_LSHIFT 0x02 MOD_LALT 0x04 MOD_LMETA 0x08 MOD_RCTRL 0x10 MOD_RSHIFT 0x20 MOD_RALT 0x40 MOD_RMETA 0x80
	ModKeys KpadModKeys `json:"mod_keys"`
	// keynames 键名
	Keys [6]string `json:"keys"`
	// 释放模式 默认为 0: 不释放 | 1: 自动释放 | > 1 表示持续时间，单位ms
	Release int `json:"release"`
	// 按键执行延迟时间 默认为5 单位s(常规使用默认为0)
	Delay int `json:"delay"`
	// 输出详情
	Verbose bool `json:"verbose"`

	commitKpadStateFn func()
}

func NewKpadOption() *KpadOption {
	return &KpadOption{
		ModKeys: []string{},
		Keys:    [6]string{},
		Release: 0,
		Delay:   0,
		Verbose: false,
	}
}

func (opt *KpadOption) ToString() string {
	opts := make([]string, 0)
	if status := opt.ModKeys.ToStatus(); status != "" {
		opts = append(opts, "--s", status)
	}

	for i, key := range opt.Keys {
		if key != "" {
			opts = append(opts, fmt.Sprintf("--x%d", i+1), KeyNameToHexCode[key])
		}
	}

	if opt.Release >= 0 {
		opts = append(opts, "--rel", fmt.Sprintf("%d", opt.Release))
	}

	if opt.Delay >= 0 {
		opts = append(opts, "--d", fmt.Sprintf("%d", opt.Delay))
	}

	if opt.Verbose {
		opts = append(opts, "--v", "1")
	}

	return strings.Join(opts, " ")
}

func (opt *KpadOption) WithKeys(keys []string) *KpadOption {
	keys = slice.Map(keys, func(_ int, key string) string {
		key = strings.ToUpper(strings.TrimSpace(key))
		switch key {
		case "CTRL":
			return "MOD_LCTRL"
		case "SHIFT":
			return "MOD_LSHIFT"
		case "ALT":
			return "MOD_LALT"
		case "META":
			return "MOD_LMETA"
		}
		if slice.Contain(NodKeyShortNames, key) {
			key = "MOD_" + key
		}
		return key
	})
	modKey, normalKeys := ModKeyNames.Extract(keys)
	opt.ModKeys = modKey
	opt.Keys = [6]string{}
	if len(normalKeys) > 0 {
		for i, nk := range normalKeys {
			if i < 6 {
				opt.Keys[i] = nk
			}
		}
	}

	return opt
}

func (opt *KpadOption) WithDelay(delay int) *KpadOption {
	opt.Delay = delay
	return opt
}

func (opt *KpadOption) WithRelease(release int) *KpadOption {
	opt.Release = release
	return opt
}

func (opt *KpadOption) WithHold() *KpadOption {
	opt.Release = 0
	return opt
}

func (opt *KpadOption) WithAutoRelease() *KpadOption {
	opt.Release = 1
	return opt
}

func (opt *KpadOption) WithDuration(duration int) *KpadOption {
	if duration > 1 {
		opt.Release = duration
	}

	return opt
}

func (opt *KpadOption) WithVerbose(verbose bool) *KpadOption {
	opt.Verbose = verbose
	return opt
}

func (opt *KpadOption) WithModKeys(modKeys []string) *KpadOption {
	opt.ModKeys = modKeys
	return opt
}

func (opt *KpadOption) WithKey(key string) *KpadOption {
	key = strings.ToUpper(key)
	opt.Keys = [6]string{}
	opt.Keys[0] = key
	return opt
}

func (opt *KpadOption) KeyDown(key string) *KpadOption {
	keys := expandKpadInputKey(key)
	if len(keys) == 0 {
		return opt
	}

	kpadStateMu.RLock()
	nextModKeys := clonePressedKeys(KpadPressedModKeysCache)
	nextNormalKeys := clonePressedKeys(KpadPressedKeysCache)

	for _, key := range keys {
		if slice.Contain(ModKeyNames, key) {
			nextModKeys = appendPressedKey(nextModKeys, key)
		} else {
			nextNormalKeys = appendPressedKey(nextNormalKeys, key)
		}
	}
	pressedModKeys := clonePressedKeys(nextModKeys)
	pressedKeys := make([]string, 0, len(nextModKeys)+len(nextNormalKeys))
	pressedKeys = append(pressedKeys, nextModKeys...)
	pressedKeys = append(pressedKeys, nextNormalKeys...)
	kpadStateMu.RUnlock()

	opt.WithKeys(pressedKeys).WithHold()
	opt.WithModKeys(pressedModKeys)
	commitKeys := append([]string(nil), keys...)
	opt.setKpadStateCommit(func() {
		kpadStateMu.Lock()
		defer kpadStateMu.Unlock()
		for _, key := range commitKeys {
			if slice.Contain(ModKeyNames, key) {
				KpadPressedModKeysCache = appendPressedKey(KpadPressedModKeysCache, key)
			} else {
				KpadPressedKeysCache = appendPressedKey(KpadPressedKeysCache, key)
			}
		}
	})
	return opt
}

func (opt *KpadOption) KeyUp(key string) (*KpadOption, *KpadOption) {
	keys := expandKpadInputKey(key)
	if len(keys) == 0 {
		return nil, nil
	}

	var releaseModKeys []string
	var remainModKeys []string
	kpadStateMu.RLock()
	releaseModKeys = snapshotPressedModKeysLocked()
	nextModKeys := clonePressedKeys(KpadPressedModKeysCache)
	nextNormalKeys := clonePressedKeys(KpadPressedKeysCache)
	for _, key := range keys {
		if slice.Contain(ModKeyNames, key) {
			nextModKeys = removePressedKey(nextModKeys, key)
		} else {
			nextNormalKeys = removePressedKey(nextNormalKeys, key)
		}
	}
	remainModKeys = clonePressedKeys(nextModKeys)
	remainKeys := make([]string, 0, len(nextModKeys)+len(nextNormalKeys))
	remainKeys = append(remainKeys, remainModKeys...)
	remainKeys = append(remainKeys, nextNormalKeys...)
	kpadStateMu.RUnlock()

	releaseOpt := NewKpadOption().
		WithDelay(opt.Delay).
		WithVerbose(opt.Verbose).
		WithKeys(keys).
		WithAutoRelease()
	if hasKpadModKey(keys) {
		releaseOpt.WithModKeys(releaseModKeys)
	}

	remainHoldOpt := NewKpadOption().
		WithDelay(opt.Delay).
		WithVerbose(opt.Verbose)

	if len(remainKeys) == 0 {
		remainHoldOpt.WithKey("NONE").WithHold()
	} else {
		remainHoldOpt.WithKeys(remainKeys).WithHold()
		if hasKpadModKey(keys) {
			remainHoldOpt.WithModKeys(remainModKeys)
		}
	}

	commitKeys := append([]string(nil), keys...)
	remainHoldOpt.setKpadStateCommit(func() {
		kpadStateMu.Lock()
		defer kpadStateMu.Unlock()
		for _, key := range commitKeys {
			if slice.Contain(ModKeyNames, key) {
				KpadPressedModKeysCache = removePressedKey(KpadPressedModKeysCache, key)
			} else {
				KpadPressedKeysCache = removePressedKey(KpadPressedKeysCache, key)
			}
		}
	})
	return releaseOpt, remainHoldOpt
}

func normalizeKpadKey(key string) string {
	key = strings.ToUpper(strings.TrimSpace(key))
	if key == "" {
		return ""
	}

	switch key {
	case "CTRL":
		return "MOD_LCTRL"
	case "SHIFT":
		return "MOD_LSHIFT"
	case "ALT":
		return "MOD_LALT"
	case "META":
		return "MOD_LMETA"
	}

	if slice.Contain(NodKeyShortNames, key) {
		return "MOD_" + key
	}

	// Keep compatibility with helper historical naming.
	if key == "LEFTBRACKET" {
		return "LEFTBRACE"
	}
	if key == "RIGHTBRACKET" {
		return "RIGHTBRACE"
	}

	return key
}

func expandKpadInputKey(key string) []string {
	switch key {
	case "":
		return nil
	case "!":
		return []string{"MOD_LSHIFT", "1"}
	case "@":
		return []string{"MOD_LSHIFT", "2"}
	case "#":
		return []string{"MOD_LSHIFT", "3"}
	case "$":
		return []string{"MOD_LSHIFT", "4"}
	case "%":
		return []string{"MOD_LSHIFT", "5"}
	case "^":
		return []string{"MOD_LSHIFT", "6"}
	case "&":
		return []string{"MOD_LSHIFT", "7"}
	case "*":
		return []string{"MOD_LSHIFT", "8"}
	case "(":
		return []string{"MOD_LSHIFT", "9"}
	case ")":
		return []string{"MOD_LSHIFT", "0"}
	case " ":
		return []string{"SPACE"}
	case "-":
		return []string{"MINUS"}
	case "_":
		return []string{"MOD_LSHIFT", "MINUS"}
	case "=":
		return []string{"EQUAL"}
	case "+":
		return []string{"MOD_LSHIFT", "EQUAL"}
	case "[":
		return []string{"LEFTBRACE"}
	case "{":
		return []string{"MOD_LSHIFT", "LEFTBRACE"}
	case "]":
		return []string{"RIGHTBRACE"}
	case "}":
		return []string{"MOD_LSHIFT", "RIGHTBRACE"}
	case "\\":
		return []string{"BACKSLASH"}
	case "|":
		return []string{"MOD_LSHIFT", "BACKSLASH"}
	case ";":
		return []string{"SEMICOLON"}
	case ":":
		return []string{"MOD_LSHIFT", "SEMICOLON"}
	case "'":
		return []string{"APOSTROPHE"}
	case "\"":
		return []string{"MOD_LSHIFT", "APOSTROPHE"}
	case "`":
		return []string{"GRAVE"}
	case "~":
		return []string{"MOD_LSHIFT", "GRAVE"}
	case ",":
		return []string{"COMMA"}
	case "<":
		return []string{"MOD_LSHIFT", "COMMA"}
	case ".":
		return []string{"DOT"}
	case ">":
		return []string{"MOD_LSHIFT", "DOT"}
	case "/":
		return []string{"SLASH"}
	case "?":
		return []string{"MOD_LSHIFT", "SLASH"}
	default:
		normalized := normalizeKpadKey(key)
		if normalized == "" {
			return nil
		}
		return []string{normalized}
	}
}

func hasKpadModKey(keys []string) bool {
	for _, key := range keys {
		if slice.Contain(ModKeyNames, key) {
			return true
		}
	}
	return false
}

func snapshotPressedModKeysLocked() []string {
	return clonePressedKeys(KpadPressedModKeysCache)
}

func snapshotPressedKpadKeysLocked() []string {
	keys := make([]string, 0, len(KpadPressedModKeysCache)+len(KpadPressedKeysCache))
	keys = append(keys, snapshotPressedModKeysLocked()...)
	keys = append(keys, clonePressedKeys(KpadPressedKeysCache)...)
	return keys
}

func clonePressedKeys(keys []string) []string {
	return append([]string(nil), keys...)
}

func appendPressedKey(keys []string, key string) []string {
	for _, cachedKey := range keys {
		if cachedKey == key {
			return keys
		}
	}
	return append(keys, key)
}

func removePressedKey(keys []string, key string) []string {
	for i, cachedKey := range keys {
		if cachedKey == key {
			return append(keys[:i], keys[i+1:]...)
		}
	}
	return keys
}

func (opt *KpadOption) setKpadStateCommit(fn func()) {
	opt.commitKpadStateFn = fn
}

func (opt *KpadOption) commitKpadState() {
	if opt == nil || opt.commitKpadStateFn == nil {
		return
	}

	commitFn := opt.commitKpadStateFn
	opt.commitKpadStateFn = nil
	commitFn()
}
