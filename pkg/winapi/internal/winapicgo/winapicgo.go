package winapicgo

//nolint:unsafeptr

// #define VC_EXTRALEAN
// #define WIN32_LEAN_AND_MEAN
// #include <windows.h>
// #include <Winuser.h>
import "C"
import (
	"unsafe"

	"github.com/operdies/windows-nats-shell/pkg/nats/api/windows"
	"github.com/operdies/windows-nats-shell/pkg/wintypes"
)

func WindowFromPoint(point windows.Point) wintypes.HWND {
	// r, _, _ := windowFromPoint.Call(uintptr(point.X), uintptr(point.Y))
	var pt C.POINT
	pt.x = C.long(point.X)
	pt.y = C.long(point.Y)
	h := C.WindowFromPoint(pt)
	hack := unsafe.Pointer(h)
	return wintypes.HWND(hack)
}

// func toCType(hwnd wintypes.HWND) C.HWND {
// 	return C.HWND(unsafe.Pointer(hwnd))
// }

const (
	GWL_EXSTYLE      = -20
	WS_EX_NOACTIVATE = 0x8000000
)

// Tool windows don't appear in the app switcher or the task bar
func MakeToolWindow(hwnd unsafe.Pointer) {
	style := C.GetWindowLong((C.HWND)(hwnd), GWL_EXSTYLE)
	style |= C.WS_EX_NOACTIVATE
	style &= ^C.WS_EX_APPWINDOW
	style |= C.WS_EX_TOOLWINDOW
	C.SetWindowLong((C.HWND)(hwnd), GWL_EXSTYLE, style)
}

func SetBottomMost(hwnd unsafe.Pointer) {
	C.SetWindowPos((C.HWND)(hwnd), C.HWND_BOTTOM, 0, 0, 0, 0, C.SWP_NOMOVE|C.SWP_NOSIZE|C.SWP_NOACTIVATE)
}
