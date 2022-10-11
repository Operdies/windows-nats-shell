package main

import (
	"log"
	"runtime"
	"time"
	"unsafe"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/nats-io/nats.go"
	"github.com/operdies/windows-nats-shell/cmd/background/colors"
	"github.com/operdies/windows-nats-shell/cmd/background/gfx"
	"github.com/operdies/windows-nats-shell/cmd/background/windowhelper"
	"github.com/operdies/windows-nats-shell/pkg/nats/api/shell"
	"github.com/operdies/windows-nats-shell/pkg/nats/client"
)

func init() {
	// This is needed to arrange that main() runs on main thread.
	// See documentation for functions that are only allowed to be called from the main thread.
	runtime.LockOSThread()
}

type customCfg struct {
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
	cfg, _ := client.Default().Request.Config("")
	custom, _ := shell.GetCustom[customCfg](cfg)
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

	colors, _ := colors.StringToColor(custom.Render.Clearcolor)

	gl.ClearColor(colors[0], colors[1], colors[2], colors[3])
	ticker := time.NewTicker(time.Millisecond * time.Duration(custom.Render.Updaterate))
	poller := time.NewTicker(time.Millisecond * 20)

	quit := make(chan bool)
	window.SetCloseCallback(func(w *glfw.Window) {
		quit <- true
	})

	nc, _ := client.New(nats.DefaultURL, time.Second)

	render := make(chan bool)

	_, err = nc.Subscribe.WH_SHELL(func(ci shell.ShellEventInfo) {
		if ci.ShellCode == shell.HSHELL_ACTIVATESHELLWINDOW {
			windowhelper.SetBottomMost(hwnd)
			render <- true
		}
	})
	if err != nil {
		panic(err)
	}

	windowhelper.MakeToolWindow(hwnd)
	windowhelper.SetBottomMost(hwnd)
	window.SetMouseButtonCallback(func(w *glfw.Window, button glfw.MouseButton, action glfw.Action, mods glfw.ModifierKey) {
		// Writing to a channel seems to cause a deadlock.
		// It works fine when doing it from a goroutine. Strange
		go func() {
			windowhelper.SetBottomMost(hwnd)
			render <- true
		}()
	})

	vertexShader, _ := gfx.NewShaderFromFile(custom.Shader.Vert, gl.VERTEX_SHADER)
	fragmentShader, err := gfx.NewShaderFromFile(custom.Shader.Frag, gl.FRAGMENT_SHADER)
	if err != nil {
		panic(err)
	}
	prog, err := gfx.NewProgram(vertexShader, fragmentShader)

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
