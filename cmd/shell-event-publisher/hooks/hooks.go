package hooks

import (
	// #include <hooks.h>
	"C"
)

func Register() {
	C.RegisterHook()
}

func Unregister() {
	C.UnregisterHook()
}
