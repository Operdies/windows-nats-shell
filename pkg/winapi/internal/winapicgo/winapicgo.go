package winapicgo

//nolint:unsafeptr

// #define VC_EXTRALEAN
// #define WIN32_LEAN_AND_MEAN
// #include <windows.h>
// #include <Winuser.h>
import "C"
import (
	"unsafe"

	"github.com/operdies/windows-nats-shell/pkg/wintypes"
)

func WindowFromPoint(point wintypes.POINT) wintypes.HWND {
	// r, _, _ := windowFromPoint.Call(uintptr(point.X), uintptr(point.Y))
	var pt C.POINT
	pt.x = C.long(point.X)
	pt.y = C.long(point.Y)
	h := C.WindowFromPoint(pt)
	hack := unsafe.Pointer(h)
	return wintypes.HWND(hack)
}

var (
	BorderlessStyles   C.long = C.WS_CAPTION | C.WS_THICKFRAME | C.WS_MINIMIZEBOX | C.WS_MAXIMIZEBOX | C.WS_SYSMENU
	BorderlessExStyles C.long = C.WS_EX_DLGMODALFRAME | C.WS_EX_CLIENTEDGE | C.WS_EX_STATICEDGE
	RedrawFlags        C.uint = C.SWP_FRAMECHANGED | C.SWP_NOMOVE | C.SWP_NOSIZE | C.SWP_NOZORDER | C.SWP_NOOWNERZORDER
)

func toCType(hwnd wintypes.HWND) C.HWND {
	return C.HWND(unsafe.Pointer(hwnd))
}

func EnableBorders(h wintypes.HWND) {
	hwnd := toCType(h)
	var lStyle C.long
	lStyle = C.GetWindowLong(hwnd, C.GWL_STYLE)
	lStyle |= BorderlessStyles
	C.SetWindowLong(hwnd, C.GWL_STYLE, lStyle)

	var eStyle C.long
	eStyle = C.GetWindowLong(hwnd, C.GWL_EXSTYLE)
	eStyle |= BorderlessExStyles
	C.SetWindowLong(hwnd, C.GWL_EXSTYLE, eStyle)
	redrawWindow(h)
	borderMap[h] = true
}

func DisableBorders(h wintypes.HWND) {
	hwnd := toCType(h)
	lStyle := C.GetWindowLong(hwnd, C.GWL_STYLE)
	lStyle &= ^BorderlessStyles
	C.SetWindowLong(hwnd, C.GWL_STYLE, lStyle)

	eStyle := C.GetWindowLong(hwnd, C.GWL_EXSTYLE)
	eStyle &= ^BorderlessExStyles
	C.SetWindowLong(hwnd, C.GWL_EXSTYLE, eStyle)
	redrawWindow(h)
	borderMap[h] = false
}

func redrawWindow(h wintypes.HWND) {
	hwnd := toCType(h)
	C.SetWindowPos(hwnd, nil, 0, 0, 0, 0, RedrawFlags)
}

var (
	borderMap = map[wintypes.HWND]bool{}
)

func BordersEnabled(h wintypes.HWND) bool {
	b, ok := borderMap[h]
	if ok {
		return b
	}
	// todo: check if the border is set with styles or somethign?
	return false
}

func ToggleBorders(h wintypes.HWND) {
	if BordersEnabled(h) {
		DisableBorders(h)
	} else {
		EnableBorders(h)
	}
}

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
