package mkpgo

import "sync"

var (
	// kpadStateMu protects global keypad press caches below.
	kpadStateMu sync.RWMutex

	// KpadPressedModKeysCache stores currently pressed modifier keys.
	// Example key names: MOD_LCTRL, MOD_LSHIFT, MOD_RALT.
	KpadPressedModKeysCache = make([]string, 0)

	// KpadPressedKeysCache stores currently pressed non-modifier keys.
	// Example key names: A, ENTER, F1, KP0.
	KpadPressedKeysCache = make([]string, 0)
)

func ResetKpadPressedCaches() {
	kpadStateMu.Lock()
	defer kpadStateMu.Unlock()

	KpadPressedModKeysCache = KpadPressedModKeysCache[:0]
	KpadPressedKeysCache = KpadPressedKeysCache[:0]
}
