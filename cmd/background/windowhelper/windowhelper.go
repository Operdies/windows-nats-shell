package windowhelper

import (
	"unsafe"
)

// #include <Windows.h>
// #include <Winuser.h>
import "C"

const (
	GWL_EXSTYLE      = -20
	WS_EX_NOACTIVATE = 0x8000000
)

// Tool windows don't appear in the app switcher or the task bar
func MakeToolWindow(hwnd unsafe.Pointer) {
	style := C.GetWindowLong((C.HWND)(hwnd), GWL_EXSTYLE)
	// style |= C.WS_EX_NOACTIVATE
	style &= ^C.WS_EX_APPWINDOW
	style |= C.WS_EX_TOOLWINDOW
	C.SetWindowLong((C.HWND)(hwnd), GWL_EXSTYLE, style)
}

func SetBottomMost(hwnd unsafe.Pointer) {
	C.SetWindowPos((C.HWND)(hwnd), C.HWND_BOTTOM, 0, 0, 0, 0, C.SWP_NOMOVE|C.SWP_NOSIZE|C.SWP_NOACTIVATE)
}
