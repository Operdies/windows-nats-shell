package winhacks

import (
	"log"
	"unsafe"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
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

type WindowPlus struct {
	GlfwWindow *glfw.Window
	Hwnd       unsafe.Pointer
}

// Set the appropriate hints and gl context to render on the background
// Returns a glfw window with its context set
func GetCanvas() *WindowPlus {
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

	w := WindowPlus{window, ptr}

	return &w
}
