package main

import (
	"log"
	"runtime"
	"strconv"
	"time"
	"unsafe"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
)

import (
	// #include <Windows.h>
	// #include <Winuser.h>
	"C"
)

func init() {
	// This is needed to arrange that main() runs on main thread.
	// See documentation for functions that are only allowed to be called from the main thread.
	runtime.LockOSThread()
}

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

func main() {
	err := glfw.Init()
	if err != nil {
		panic(err)
	}
	defer glfw.Terminate()

	glfw.WindowHint(glfw.Resizable, glfw.False)
	glfw.WindowHint(glfw.Decorated, glfw.False)
	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)

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

	strToCol := func(s string) [4]float32 {
		col := func(c string) float32 {
			v, _ := strconv.ParseUint(c, 16, 8)
			return float32(v) / 255
		}
		a := s[:2]
		r := s[2:4]
		g := s[4:6]
		b := s[6:]
		return [4]float32{col(r), col(g), col(b), col(a)}
	}

	myClear := func(intensity float32) {
		colors := strToCol("00ac21c4")
		for i := range colors {
			colors[i] *= intensity
		}
		gl.ClearColor(colors[0], colors[1], colors[2], colors[3])
	}

	myClear(0.2)
	ticker := time.NewTicker(time.Millisecond * 1000)

	for range ticker.C {
		if window.ShouldClose() {
			break
		}

		// OpenGL START

		myClear(0.2)
		gl.Clear(gl.COLOR_BUFFER_BIT)

		// OpenGL END

		window.SwapBuffers()
		glfw.PollEvents()
	}

}
