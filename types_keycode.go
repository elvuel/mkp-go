package mkpgo

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/duke-git/lancet/v2/slice"
)

// Keyboard key-name helper sets.
// 键盘键名辅助集合。
var (
	// ModKeyNames lists supported modifier key names.
	// ModKeyNames 定义支持的修饰键名称集合。
	ModKeyNames KpadModKeys = []string{"MOD_LCTRL", "MOD_LSHIFT", "MOD_LALT", "MOD_LMETA", "MOD_RCTRL", "MOD_RSHIFT", "MOD_RALT", "MOD_RMETA"}
	// NodKeyShortNames is accepted short form list for modifier keys.
	// NodKeyShortNames 定义可接受的修饰键短名称。
	NodKeyShortNames = []string{"LCTRL", "LSHIFT", "LALT", "LMETA", "RCTRL", "RSHIFT", "RALT", "RMETA"}
)

// KpadModKeys is a helper type for modifier-key operations.
// KpadModKeys 是修饰键集合辅助类型。
type KpadModKeys []string

// ToStatus converts mod keys to HID status byte string.
// ToStatus 将修饰键集合转换为 HID 状态字节字符串。
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

// Contains reports whether any given key is a modifier key.
// Contains 判断输入键列表中是否包含修饰键。
func (m KpadModKeys) Contains(keys []string) bool {
	for _, key := range keys {
		if slice.Contain(m, key) {
			return true
		}
	}
	return false
}

// Extract splits keys into modifier keys and normal keys.
// Extract 将输入键拆分为修饰键与普通键。
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

// KpadOption models one keyboard (kpad) directive payload.
// KpadOption 描述一次键盘（kpad）指令的参数载荷。
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
	// 是否异步发送，true: SendDirectiveAsync, false: SendDirective
	Async bool `json:"async"`

	// commitKpadStateFn runs after successful kpad send.
	// commitKpadStateFn 在 kpad 发送成功后提交本地状态变更。
	commitKpadStateFn func()
}

// NewKpadOption creates default keyboard directive option.
// NewKpadOption 创建默认键盘指令参数对象。
func NewKpadOption() *KpadOption {
	return &KpadOption{
		ModKeys: []string{},
		Keys:    [6]string{},
		Release: 0,
		Delay:   0,
		Verbose: false,
		Async:   true,
	}
}

// kpadOptionJSON is the JSON helper structure used during custom decoding.
// kpadOptionJSON 是自定义解码时使用的 JSON 辅助结构。
type kpadOptionJSON struct {
	ModKeys KpadModKeys `json:"mod_keys"`
	Keys    [6]string   `json:"keys"`
	Release int         `json:"release"`
	Delay   int         `json:"delay"`
	Verbose bool        `json:"verbose"`
	Async   *bool       `json:"async"`
}

// UnmarshalJSON keeps async default compatible with historical behavior.
// UnmarshalJSON 在未显式提供 async 时保持历史默认异步行为。
func (opt *KpadOption) UnmarshalJSON(data []byte) error {
	raw := kpadOptionJSON{}
	*opt = *NewKpadOption()
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	opt.ModKeys = append(KpadModKeys(nil), raw.ModKeys...)
	opt.Keys = raw.Keys
	opt.Release = raw.Release
	opt.Delay = raw.Delay
	opt.Verbose = raw.Verbose
	if raw.Async != nil {
		opt.Async = *raw.Async
	}

	return nil
}

// ToString converts KpadOption into CLI argument fragment.
// ToString 将 KpadOption 转换为命令行参数片段。
func (opt *KpadOption) ToString() string {
	if opt == nil {
		return ""
	}

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

// WithKeys normalizes and sets up to 6 key slots (and mod keys).
// WithKeys 规范化并设置最多 6 个按键槽（包含修饰键拆分）。
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

// WithDelay sets directive delay value.
// WithDelay 设置指令延迟值。
func (opt *KpadOption) WithDelay(delay int) *KpadOption {
	opt.Delay = delay
	return opt
}

// WithRelease sets release mode value directly.
// WithRelease 直接设置释放模式值。
func (opt *KpadOption) WithRelease(release int) *KpadOption {
	opt.Release = release
	return opt
}

// WithHold sets hold mode (no auto release).
// WithHold 设置为按住模式（不自动释放）。
func (opt *KpadOption) WithHold() *KpadOption {
	opt.Release = 0
	return opt
}

// WithAutoRelease sets auto-release mode.
// WithAutoRelease 设置为自动释放模式。
func (opt *KpadOption) WithAutoRelease() *KpadOption {
	opt.Release = 1
	return opt
}

// WithDuration sets duration-based release when duration > 1.
// WithDuration 在 duration > 1 时设置时长释放模式。
func (opt *KpadOption) WithDuration(duration int) *KpadOption {
	if duration > 1 {
		opt.Release = duration
	}

	return opt
}

// WithVerbose controls verbose flag.
// WithVerbose 控制 verbose 输出开关。
func (opt *KpadOption) WithVerbose(verbose bool) *KpadOption {
	opt.Verbose = verbose
	return opt
}

// WithAsync controls whether kpad uses async send mode.
// WithAsync 控制 kpad 是否使用异步发送模式。
func (opt *KpadOption) WithAsync(async bool) *KpadOption {
	opt.Async = async
	return opt
}

// WithModKeys sets modifier keys directly.
// WithModKeys 直接设置修饰键集合。
func (opt *KpadOption) WithModKeys(modKeys []string) *KpadOption {
	opt.ModKeys = modKeys
	return opt
}

// WithKey sets single key into first key slot.
// WithKey 将单键写入第一个按键槽位。
func (opt *KpadOption) WithKey(key string) *KpadOption {
	key = strings.ToUpper(key)
	opt.Keys = [6]string{}
	opt.Keys[0] = key
	return opt
}

// KeyDown prepares hold packet from current cache + newly pressed key.
// KeyDown 基于当前缓存与新按下键计算按下包，并延迟提交状态。
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

// KeyUp prepares release packet and remain-hold packet for one key.
// KeyUp 计算单键抬起所需的释放包和剩余按住包。
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
		WithAsync(opt.Async).
		WithKeys(keys).
		WithAutoRelease()
	if hasKpadModKey(keys) {
		releaseOpt.WithModKeys(releaseModKeys)
	}

	remainHoldOpt := NewKpadOption().
		WithDelay(opt.Delay).
		WithVerbose(opt.Verbose).
		WithAsync(opt.Async)

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

// normalizeKpadKey normalizes aliases into canonical key names.
// normalizeKpadKey 将别名规范化为标准键名。
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

// expandKpadInputKey expands printable symbol into key-combo sequence.
// expandKpadInputKey 将可打印字符展开为键位组合。
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

// hasKpadModKey reports whether keys include modifier key.
// hasKpadModKey 判断键列表是否包含修饰键。
func hasKpadModKey(keys []string) bool {
	for _, key := range keys {
		if slice.Contain(ModKeyNames, key) {
			return true
		}
	}
	return false
}

// snapshotPressedModKeysLocked returns current pressed modifier keys snapshot.
// snapshotPressedModKeysLocked 返回当前修饰键按下快照（需持锁）。
func snapshotPressedModKeysLocked() []string {
	return clonePressedKeys(KpadPressedModKeysCache)
}

// snapshotPressedKpadKeysLocked returns current pressed key snapshot.
// snapshotPressedKpadKeysLocked 返回当前全部按键快照（需持锁）。
func snapshotPressedKpadKeysLocked() []string {
	keys := make([]string, 0, len(KpadPressedModKeysCache)+len(KpadPressedKeysCache))
	keys = append(keys, snapshotPressedModKeysLocked()...)
	keys = append(keys, clonePressedKeys(KpadPressedKeysCache)...)
	return keys
}

// clonePressedKeys clones a pressed-key slice.
// clonePressedKeys 复制按键切片。
func clonePressedKeys(keys []string) []string {
	return append([]string(nil), keys...)
}

// appendPressedKey appends key once while preserving press order.
// appendPressedKey 去重追加按键并保持按下顺序。
func appendPressedKey(keys []string, key string) []string {
	for _, cachedKey := range keys {
		if cachedKey == key {
			return keys
		}
	}
	return append(keys, key)
}

// removePressedKey removes key once while preserving remaining order.
// removePressedKey 删除按键并保持剩余顺序。
func removePressedKey(keys []string, key string) []string {
	for i, cachedKey := range keys {
		if cachedKey == key {
			return append(keys[:i], keys[i+1:]...)
		}
	}
	return keys
}

// setKpadStateCommit sets delayed state-commit callback.
// setKpadStateCommit 设置延迟状态提交回调。
func (opt *KpadOption) setKpadStateCommit(fn func()) {
	opt.commitKpadStateFn = fn
}

// commitKpadState executes delayed state commit once.
// commitKpadState 执行一次性延迟状态提交。
func (opt *KpadOption) commitKpadState() {
	if opt == nil || opt.commitKpadStateFn == nil {
		return
	}

	commitFn := opt.commitKpadStateFn
	opt.commitKpadStateFn = nil
	commitFn()
}
