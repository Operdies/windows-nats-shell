package border

import (
	// #include <Windows.h>
	"C"
)
import (
	"unsafe"

	"github.com/operdies/windows-nats-shell/pkg/winapi/wintypes"
)

var (
	STYLES    C.long = C.WS_CAPTION | C.WS_THICKFRAME | C.WS_MINIMIZEBOX | C.WS_MAXIMIZEBOX | C.WS_SYSMENU
	EX_STYLES C.long = C.WS_EX_DLGMODALFRAME | C.WS_EX_CLIENTEDGE | C.WS_EX_STATICEDGE
	REDRAW    C.uint = C.SWP_FRAMECHANGED | C.SWP_NOMOVE | C.SWP_NOSIZE | C.SWP_NOZORDER | C.SWP_NOOWNERZORDER
)

func toCType(hwnd wintypes.HWND) C.HWND {
	return C.HWND(unsafe.Pointer(hwnd))
}

func Enable(h wintypes.HWND) {
	hwnd := toCType(h)
	var lStyle C.long
	lStyle = C.GetWindowLong(hwnd, C.GWL_STYLE)
	lStyle |= STYLES
	C.SetWindowLong(hwnd, C.GWL_STYLE, lStyle)

	var eStyle C.long
	eStyle = C.GetWindowLong(hwnd, C.GWL_EXSTYLE)
	eStyle |= EX_STYLES
	C.SetWindowLong(hwnd, C.GWL_EXSTYLE, eStyle)
	Redraw(h)
}

func Disable(h wintypes.HWND) {
	hwnd := toCType(h)
	lStyle := C.GetWindowLong(hwnd, C.GWL_STYLE)
	lStyle &= ^STYLES
	C.SetWindowLong(hwnd, C.GWL_STYLE, lStyle)

	eStyle := C.GetWindowLong(hwnd, C.GWL_EXSTYLE)
	eStyle &= ^EX_STYLES
	C.SetWindowLong(hwnd, C.GWL_EXSTYLE, eStyle)
	Redraw(h)
}

func Redraw(h wintypes.HWND) {
	hwnd := toCType(h)
	C.SetWindowPos(hwnd, nil, 0, 0, 0, 0, REDRAW)
}
