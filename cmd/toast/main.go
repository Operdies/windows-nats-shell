package main

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/4ydx/gltext"
	v45 "github.com/4ydx/gltext/v4.5"
	"github.com/nats-io/nats.go"
	"golang.org/x/image/math/fixed"

	"github.com/go-gl/gl/v4.5-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"

	// "github.com/go-gl/gltext"
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

	fmt.Printf("Hello, toast\n")
	glfw.DefaultWindowHints()
	glfw.WindowHint(glfw.Resizable, glfw.False)
	glfw.WindowHint(glfw.Decorated, glfw.False)
	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)
	glfw.WindowHint(glfw.DoubleBuffer, glfw.False)

	glfw.WindowHint(glfw.Focused, glfw.False)
	glfw.WindowHint(glfw.FocusOnShow, glfw.False)
	glfw.WindowHint(glfw.Floating, glfw.True)
	glfw.WindowHint(glfw.TransparentFramebuffer, glfw.True)

	const width, height = 400, 400
	window, err := glfw.CreateWindow(width, height, "Toast", nil, nil)
	if err != nil {
		panic(err)
	}

	window.MakeContextCurrent()
	if err := gl.Init(); err != nil {
		log.Fatalln(err)
	}

	nc := client.Default()
	nc.Subscribe.ShellToast(func(t shell.Toast) {
		fmt.Printf("t: %+v\n", t)
	})
	c2, _ := nats.Connect(nats.DefaultURL)
	messages := make([]string, 0, 10)
	msgLock := sync.Mutex{}
	c2.Subscribe("stdout.>", func(msg *nats.Msg) {
		msgLock.Lock()
		defer msgLock.Unlock()
		if len(messages) >= 10 {
			messages = messages[1:]
		}
		messages = append(messages, msg.Subject+" : "+string(msg.Data))
	})

	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)

	font := loadFont()
	font.ResizeWindow(width, height)
	text := v45.NewText(font, 1.0, 1.0)
	text.SetString("Hello, toast")
	text.SetColor(mgl32.Vec3{0, 0, 0})

	doRender := make(chan bool)

	var render func()
	render = func() {
		width, height := window.GetSize()
		gl.ClearColor(0.5, 0.1, 0.7, 1)
		gl.Clear(gl.COLOR_BUFFER_BIT)
		var maxWidth float32
		pos := float32(height/2) - text.Height()/2
		startPos := pos
		text.SetColor(mgl32.Vec3{1, 1, 0})
		for k, m := range messages {
			str := m
			text.SetString(fmt.Sprintf("%v) %v", k, str))
			w := text.Width()
			if w > float32(maxWidth) {
				maxWidth = w
			}
			xPos := (w / 2) - (float32(width / 2))
			text.SetPosition(mgl32.Vec2{xPos, pos})
			text.Draw()
			pos -= (text.Height() / 2) + 1
		}
		newHeight := int(startPos-pos) + 10
		newWidth := int(maxWidth) + 10
		if newHeight != height || newWidth != width {
			font.ResizeWindow(float32(newWidth), float32(newHeight))
			gl.Viewport(0, 0, int32(newWidth), int32(newHeight))
			window.SetSize(newWidth, newHeight)
			render()
			return
		}
		gl.Finish()
		glfw.PollEvents()
	}

	render()

	ticker := time.NewTicker(time.Millisecond * 1000)
	poller := time.NewTicker(time.Millisecond)

	for !window.ShouldClose() {
		select {
		case <-doRender:
			render()
		case <-ticker.C:
			render()
		case <-poller.C:
			glfw.PollEvents()
		}
	}
}

func h[T any](r T, err error) T {
	if err != nil {
		log.Fatalln(err)
	}
	return r
}

func loadFont() *v45.Font {
	fmt.Println("Make font")
	fd := h(os.Open(`C:\Windows\Fonts\arial.ttf`))
	defer fd.Close()
	// Japanese character ranges
	// http://www.rikai.com/library/kanjitables/kanji_codes.unicode.shtml
	runeRanges := make(gltext.RuneRanges, 0)
	runeRanges = append(runeRanges, gltext.RuneRange{Low: 32, High: 128})
	runeRanges = append(runeRanges, gltext.RuneRange{Low: 0x3000, High: 0x3030})
	runeRanges = append(runeRanges, gltext.RuneRange{Low: 0x3040, High: 0x309f})
	runeRanges = append(runeRanges, gltext.RuneRange{Low: 0x30a0, High: 0x30ff})
	runeRanges = append(runeRanges, gltext.RuneRange{Low: 0x4e00, High: 0x9faf})
	runeRanges = append(runeRanges, gltext.RuneRange{Low: 0xff00, High: 0xffef})

	scale := fixed.Int26_6(18)
	runesPerRow := fixed.Int26_6(128)
	config := h(gltext.NewTruetypeFontConfig(fd, scale, runeRanges, runesPerRow, 5))
	font := h(v45.NewFont(config))
	fmt.Println("Font made")
	return font
}
