package mkpgo

import "sync"

// Global keypad pressed-key caches.
// 全局键盘按下状态缓存。
var (
	// kpadStateMu protects keypad pressed-key caches below.
	// kpadStateMu 用于保护下面的键盘按下状态缓存。
	kpadStateMu sync.RWMutex

	// KpadPressedModKeysCache stores pressed modifier keys in press order.
	// KpadPressedModKeysCache 按按下先后顺序记录当前按下的修饰键。
	// Example key names / 示例键名: MOD_LCTRL, MOD_LSHIFT, MOD_RALT.
	KpadPressedModKeysCache = make([]string, 0)

	// KpadPressedKeysCache stores pressed non-modifier keys in press order.
	// KpadPressedKeysCache 按按下先后顺序记录当前按下的普通键。
	// Example key names / 示例键名: A, ENTER, F1, KP0.
	KpadPressedKeysCache = make([]string, 0)
)

// ResetKpadPressedCaches clears all pressed-key caches.
// ResetKpadPressedCaches 清空当前全部按键缓存（修饰键与普通键）。
func ResetKpadPressedCaches() {
	kpadStateMu.Lock()
	defer kpadStateMu.Unlock()

	KpadPressedModKeysCache = KpadPressedModKeysCache[:0]
	KpadPressedKeysCache = KpadPressedKeysCache[:0]
}
