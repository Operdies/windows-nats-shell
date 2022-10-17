package winapicgo

// #define VC_EXTRALEAN
// #define WIN32_LEAN_AND_MEAN
// #include <windows.h>
import "C"
import (
	"unsafe"

	"github.com/operdies/windows-nats-shell/pkg/winapi/wintypes"
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
