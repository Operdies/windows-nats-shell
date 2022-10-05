package main

import (
	"runtime"
	"strconv"
	"time"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/nats-io/nats.go"
	"github.com/operdies/windows-nats-shell/cmd/background/winhacks"
	"github.com/operdies/windows-nats-shell/pkg/nats/api/shell"
	"github.com/operdies/windows-nats-shell/pkg/nats/client"
)

func init() {
	// This is needed to arrange that main() runs on main thread.
	// See documentation for functions that are only allowed to be called from the main thread.
	runtime.LockOSThread()
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
	glfw.WindowHint(glfw.DoubleBuffer, glfw.False)

	w2 := winhacks.GetCanvas()
	window := w2.GlfwWindow

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
	colStr := "00ac21c4"
	myClear := func(intensity float32) {
		colors := strToCol(colStr)
		for i := range colors {
			colors[i] *= intensity
		}
		gl.ClearColor(colors[0], colors[1], colors[2], colors[3])
	}

	intensity := 0.6
	myClear(float32(intensity))
	ticker := time.NewTicker(time.Millisecond * 5000)

	step := func() {
		// window.SwapBuffers()
		glfw.PollEvents()
		gl.Finish()
	}

	quit := make(chan bool)
	window.SetCloseCallback(func(w *glfw.Window) {
		quit <- true
	})

	nc, _ := client.New(nats.DefaultURL, time.Second)

	render := make(chan bool)

	nc.Subscribe.WH_CBT(func(ci shell.CBTEventInfo) {
		if ci.CBTCode == shell.HCBT_SETFOCUS {
			winhacks.SetBottomMost(w2.Hwnd)
			render <- true
		}
	})

	winhacks.MakeToolWindow(w2.Hwnd)
	winhacks.SetBottomMost(w2.Hwnd)
	window.SetMouseButtonCallback(func(w *glfw.Window, button glfw.MouseButton, action glfw.Action, mods glfw.ModifierKey) {
		winhacks.SetBottomMost(w2.Hwnd)
		render <- true
	})

	for {
		select {
		case <-quit:
			return
		case <-render:
			myClear(float32(intensity))
			gl.Clear(gl.COLOR_BUFFER_BIT)
			step()
		case <-ticker.C:
			myClear(float32(intensity))
			gl.Clear(gl.COLOR_BUFFER_BIT)
			step()
		}
	}
}
