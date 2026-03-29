package mkpgo

import "sync"

var (
	// kpadStateMu protects global keypad press caches below.
	kpadStateMu sync.RWMutex

	// KpadPressedModKeysCache stores currently pressed modifier keys.
	// Example key names: MOD_LCTRL, MOD_LSHIFT, MOD_RALT.
	KpadPressedModKeysCache = make(map[string]struct{})

	// KpadPressedKeysCache stores currently pressed non-modifier keys.
	// Example key names: A, ENTER, F1, KP0.
	KpadPressedKeysCache = make(map[string]struct{})
)

func ResetKpadPressedCaches() {
	kpadStateMu.Lock()
	defer kpadStateMu.Unlock()

	clear(KpadPressedModKeysCache)
	clear(KpadPressedKeysCache)
}
