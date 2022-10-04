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

	window := winhacks.GetCanvas()

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

	intensity := 0.2
	myClear(float32(intensity))
	ticker := time.NewTicker(time.Millisecond * 30)

	step := func() {
		// window.SwapBuffers()
		glfw.PollEvents()
    gl.Finish()
	}

	quit := make(chan bool)
	window.SetCloseCallback(func(w *glfw.Window) {
		quit <- true
	})

	const (
		s = 83
		w = 87
	)
	clamp := func(v, min, max float64) float64 {
		if v < min {
			return min
		}
		if v > max {
			return max
		}
		return v
	}
	controls := make(chan bool)
	nc, _ := client.New(nats.DefaultURL, time.Second)
	nc.Subscribe.WH_KEYBOARD(func(kei shell.KeyboardEventInfo) {
		if kei.VirtualKeyCode == s && kei.PreviousKeyState == false {
			intensity -= 0.1
		} else if kei.VirtualKeyCode == w && kei.PreviousKeyState == false {
			intensity += 0.1
		} else {
			return
		}
		intensity = clamp(intensity, 0, 1)
		controls <- true
	})

	for {
		select {
		case <-quit:
			return
		case <-controls:
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
