package main

import (
	"log"
	"runtime"
	"sync"
	"time"
	"unsafe"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/operdies/windows-nats-shell/pkg/gfx/colors"
	"github.com/operdies/windows-nats-shell/pkg/gfx/shaders"
	"github.com/operdies/windows-nats-shell/pkg/nats/api/shell"
	"github.com/operdies/windows-nats-shell/pkg/nats/client"
	wia "github.com/operdies/windows-nats-shell/pkg/winapi/winapiabstractions"
)

func init() {
	// This is needed to arrange that main() runs on main thread.
	// See documentation for functions that are only allowed to be called from the main thread.
	runtime.LockOSThread()
}

type config struct {
	Background struct {
		Path string
	}
	Render struct {
		Updaterate uint32
		Clearcolor string
	}
	Shader struct {
		Vert string
		Frag string
	}
}

/*
 * Creates the Vertex Array Object for a triangle.
 */
func createQuadVAO() uint32 {
	vertices := []float32{
		-1.0, -1.0, 0.0,
		1.0, -1.0, 0.0,
		-1.0, 1.0, 0.0,
		1.0, 1.0, 0.0,
	}

	var VAO uint32
	gl.GenVertexArrays(1, &VAO)

	var VBO uint32
	gl.GenBuffers(1, &VBO)

	// Bind the Vertex Array Object first, then bind and set vertex buffer(s) and attribute pointers()
	gl.BindVertexArray(VAO)

	// copy vertices data into VBO (it needs to be bound first)
	gl.BindBuffer(gl.ARRAY_BUFFER, VBO)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)

	// specify the format of our vertex input
	// (shader) input 0
	// vertex has size 3
	// vertex items are of type FLOAT
	// do not normalize (already done)
	// stride of 3 * sizeof(float) (separation of vertices)
	// offset of where the position data starts (0 for the beginning)
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 3*4, gl.PtrOffset(0))
	gl.EnableVertexAttribArray(0)

	// unbind the VAO (safe practice so we don't accidentally (mis)configure it later)
	gl.BindVertexArray(0)

	return VAO
}

func main() {
	nc := client.Default()
	defer nc.Close()
	cfg := client.GetConfig[config](nc.Request)
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

	glfw.WindowHint(glfw.Focused, glfw.False)
	glfw.WindowHint(glfw.FocusOnShow, glfw.False)
	glfw.WindowHint(glfw.Floating, glfw.True)
	glfw.WindowHint(glfw.TransparentFramebuffer, glfw.True)

	mon := glfw.GetPrimaryMonitor()
	_, _, width, height := mon.GetWorkarea()

	// Set the appropriate hints and gl context to render on the background
	// Returns a glfw window with its context set
	window, err := glfw.CreateWindow(width, height, "Background", nil, nil)
	if err != nil {
		panic(err)
	}

	window.MakeContextCurrent()
	if err := gl.Init(); err != nil {
		log.Fatalln(err)
	}

	hwnd := unsafe.Pointer(window.GetWin32Window())

	colors, _ := colors.StringToColor(cfg.Render.Clearcolor)

	gl.ClearColor(colors[0], colors[1], colors[2], colors[3])
	ticker := time.NewTicker(time.Millisecond * time.Duration(cfg.Render.Updaterate))
	poller := time.NewTicker(time.Millisecond * 20)

	quit := make(chan bool)
	window.SetCloseCallback(func(w *glfw.Window) {
		quit <- true
	})

	render := make(chan bool)

	_, err = nc.Subscribe.WH_SHELL(func(ci shell.ShellEventInfo) {
		wia.SetBottomMost(hwnd)
		render <- true
	})
	if err != nil {
		panic(err)
	}

	shown := true
	showLock := sync.Mutex{}
	nc.Subscribe.ToggleBackground(func() bool {
		showLock.Lock()
		defer showLock.Unlock()
		if shown {
			window.Hide()
		} else {
			window.Show()
		}
		shown = !shown
		return true
	})

	wia.MakeToolWindow(hwnd)
	wia.SetBottomMost(hwnd)
	window.SetMouseButtonCallback(func(w *glfw.Window, button glfw.MouseButton, action glfw.Action, mods glfw.ModifierKey) {
		// Writing to a channel seems to cause a deadlock.
		// It works fine when doing it from a goroutine. Strange
		go func() {
			wia.SetBottomMost(hwnd)
			render <- true
		}()
	})

	vertexShader, _ := shaders.NewShaderFromFile(cfg.Shader.Vert, gl.VERTEX_SHADER)
	fragmentShader, err := shaders.NewShaderFromFile(cfg.Shader.Frag, gl.FRAGMENT_SHADER)
	if err != nil {
		panic(err)
	}
	prog, err := shaders.NewProgram(vertexShader, fragmentShader)

	if err != nil {
		panic(err)
	}
	vao := createQuadVAO()

	fixResolution := func(force bool) {
		widthName := "ScreenWidth"
		heightName := "ScreenHeight"
		_, _, newWidth, newHeight := mon.GetWorkarea()
		if force == false && newWidth == width && newHeight == height {
			return
		}
		// Resize detected
		width = newWidth
		height = newHeight
		window.SetSize(newWidth, newHeight)
		gl.Viewport(0, 0, int32(newWidth), int32(newHeight))

		prog.Use()
		for _, u := range prog.GetUniforms() {
			if u.Name == widthName {
				prog.SetUniform(&u, float32(newWidth))
			} else if u.Name == heightName {
				prog.SetUniform(&u, float32(newHeight))
			}
		}
	}

	first := true
	doRender := func() {
		fixResolution(first)
		gl.ClearColor(colors[0], colors[1], colors[2], colors[3])
		gl.Clear(gl.COLOR_BUFFER_BIT)
		gl.BindVertexArray(vao)
		gl.DrawArrays(gl.TRIANGLE_STRIP, 0, 4)
		gl.BindVertexArray(0)
		glfw.PollEvents()
		gl.Finish()
	}

	doRender()
	first = false

	for {
		select {
		case <-quit:
			return
		case <-render:
			doRender()
		case <-ticker.C:
			doRender()
		case <-poller.C:
			glfw.PollEvents()
		}
	}
}
