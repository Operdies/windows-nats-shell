package winhacks

import (
	// #include <Windows.h>
	// #include <Winuser.h>
	"C"
)

import (
	"log"
	"time"
	"unsafe"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
)

const (
	GWL_EXSTYLE      = -20
	WS_EX_NOACTIVATE = 0x8000000
)

func makeUnfocusable(hwnd2 C.HWND) {
	style := C.GetWindowLong(hwnd2, GWL_EXSTYLE)
	style |= C.WS_EX_NOACTIVATE
	style &= ^C.WS_EX_APPWINDOW
	style |= C.WS_EX_TOOLWINDOW
	C.SetWindowLong(hwnd2, GWL_EXSTYLE, style)
}

func setBottomMost(hwnd2 C.HWND) {
	C.SetWindowPos(hwnd2, C.HWND_BOTTOM, 0, 0, 0, 0, C.SWP_NOMOVE|C.SWP_NOSIZE|C.SWP_NOACTIVATE)
}


// Set the appropriate hints and gl context to render on the background
// Returns a glfw window with its context set
func GetCanvas() *glfw.Window {
	window, err := glfw.CreateWindow(1920, 1080, "Background", nil, nil)
	if err != nil {
		panic(err)
	}

	window.MakeContextCurrent()
	if err := gl.Init(); err != nil {
		log.Fatalln(err)
	}

	// Nice hack :)
	hwnd := window.GetWin32Window()
	ptr := unsafe.Pointer(hwnd)
	hwnd2 := C.HWND(ptr)

	go func() {
		for range time.NewTicker(time.Millisecond * 10).C {
			makeUnfocusable(hwnd2)
			setBottomMost(hwnd2)
		}
	}()

	return window

}
