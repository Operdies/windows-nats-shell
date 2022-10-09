package gfx

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/go-gl/gl/v4.1-core/gl"
)

type Shader struct {
	Handle uint32
}

type Program struct {
	handle  uint32
	shaders []*Shader
}

func (shader *Shader) Delete() {
	gl.DeleteShader(shader.Handle)
}

func (prog *Program) Delete() {
	for _, shader := range prog.shaders {
		shader.Delete()
	}
	gl.DeleteProgram(prog.handle)
}

type Uniform struct {
	Name     string
	GlType   GLTYPE
	Size     int32
	Location int32
}

func (program *Program) SetUniform(uniform *Uniform, value float32) {
	switch uniform.GlType {
	case FLOAT:
		fmt.Printf("Setting %v to %v\n", uniform, value)
		gl.Uniform1f(uniform.Location, value)
	}
}

func (program *Program) GetUniforms() []Uniform {
	var count int32
	gl.GetProgramiv(program.handle, gl.ACTIVE_UNIFORMS, &count)
	result := make([]Uniform, 0, count)
	buf := make([]byte, 100)
	var i int32
	for i = 0; i < count; i++ {
		var nameLength int32
		var uniform Uniform
		var tp uint32
		gl.GetActiveUniform(program.handle, uint32(i), int32(len(buf)), &nameLength, &uniform.Size, &tp, &buf[0])
		uniform.Name = string(buf[:nameLength])
		uniform.Location = i
		uniform.GlType = int(tp)
		loc := program.GetUniformLocation(uniform.Name)
		fmt.Printf("Index: %d, loc: %d\n", i, loc)
		result = append(result, uniform)
	}

	return result
}

func (prog *Program) Attach(shaders ...*Shader) {
	for _, shader := range shaders {
		gl.AttachShader(prog.handle, shader.Handle)
		prog.shaders = append(prog.shaders, shader)
	}
}

func (prog *Program) Use() {
	gl.UseProgram(prog.handle)
}

func (prog *Program) Link() error {
	gl.LinkProgram(prog.handle)
	return getGlError(prog.handle, gl.LINK_STATUS, gl.GetProgramiv, gl.GetProgramInfoLog,
		"PROGRAM::LINKING_FAILURE")
}

func (prog *Program) GetUniformLocation(name string) int32 {
	return gl.GetUniformLocation(prog.handle, gl.Str(name+"\x00"))
}

func NewProgram(shaders ...*Shader) (*Program, error) {
	prog := &Program{handle: gl.CreateProgram()}
	prog.Attach(shaders...)

	if err := prog.Link(); err != nil {
		return nil, err
	}

	return prog, nil
}

func NewShader(src string, sType uint32) (*Shader, error) {

	handle := gl.CreateShader(sType)
	glSrcs, freeFn := gl.Strs(src + "\x00")
	defer freeFn()
	gl.ShaderSource(handle, 1, glSrcs, nil)
	gl.CompileShader(handle)
	err := getGlError(handle, gl.COMPILE_STATUS, gl.GetShaderiv, gl.GetShaderInfoLog,
		"SHADER::COMPILE_FAILURE::")
	if err != nil {
		return nil, err
	}
	return &Shader{Handle: handle}, nil
}

func NewShaderFromFile(file string, sType uint32) (*Shader, error) {
	r, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	src, err := io.ReadAll(r)

	if err != nil {
		return nil, err
	}
	handle := gl.CreateShader(sType)
	glSrc, freeFn := gl.Strs(string(src) + "\x00")
	defer freeFn()
	gl.ShaderSource(handle, 1, glSrc, nil)
	gl.CompileShader(handle)
	err = getGlError(handle, gl.COMPILE_STATUS, gl.GetShaderiv, gl.GetShaderInfoLog,
		"SHADER::COMPILE_FAILURE::"+file)
	if err != nil {
		return nil, err
	}
	return &Shader{Handle: handle}, nil
}

type getObjIv func(uint32, uint32, *int32)
type getObjInfoLog func(uint32, int32, *int32, *uint8)

func getGlError(glHandle uint32, checkTrueParam uint32, getObjIvFn getObjIv,
	getObjInfoLogFn getObjInfoLog, failMsg string) error {

	var success int32
	getObjIvFn(glHandle, checkTrueParam, &success)

	if success == gl.FALSE {
		var logLength int32
		getObjIvFn(glHandle, gl.INFO_LOG_LENGTH, &logLength)

		log := gl.Str(strings.Repeat("\x00", int(logLength)))
		getObjInfoLogFn(glHandle, logLength, nil, log)

		return fmt.Errorf("%s: %s", failMsg, gl.GoStr(log))
	}

	return nil
}

type GLTYPE = int

const (
	FLOAT                                     GLTYPE = gl.FLOAT
	FLOAT_VEC2                                       = gl.FLOAT_VEC2
	FLOAT_VEC3                                       = gl.FLOAT_VEC3
	FLOAT_VEC4                                       = gl.FLOAT_VEC4
	DOUBLE                                           = gl.DOUBLE
	DOUBLE_VEC2                                      = gl.DOUBLE_VEC2
	DOUBLE_VEC3                                      = gl.DOUBLE_VEC3
	DOUBLE_VEC4                                      = gl.DOUBLE_VEC4
	INT                                              = gl.INT
	INT_VEC2                                         = gl.INT_VEC2
	INT_VEC3                                         = gl.INT_VEC3
	INT_VEC4                                         = gl.INT_VEC4
	UNSIGNED_INT                                     = gl.UNSIGNED_INT
	UNSIGNED_INT_VEC2                                = gl.UNSIGNED_INT_VEC2
	UNSIGNED_INT_VEC3                                = gl.UNSIGNED_INT_VEC3
	UNSIGNED_INT_VEC4                                = gl.UNSIGNED_INT_VEC4
	BOOL                                             = gl.BOOL
	BOOL_VEC2                                        = gl.BOOL_VEC2
	BOOL_VEC3                                        = gl.BOOL_VEC3
	BOOL_VEC4                                        = gl.BOOL_VEC4
	FLOAT_MAT2                                       = gl.FLOAT_MAT2
	FLOAT_MAT3                                       = gl.FLOAT_MAT3
	FLOAT_MAT4                                       = gl.FLOAT_MAT4
	FLOAT_MAT2x3                                     = gl.FLOAT_MAT2x3
	FLOAT_MAT2x4                                     = gl.FLOAT_MAT2x4
	FLOAT_MAT3x2                                     = gl.FLOAT_MAT3x2
	FLOAT_MAT3x4                                     = gl.FLOAT_MAT3x4
	FLOAT_MAT4x2                                     = gl.FLOAT_MAT4x2
	FLOAT_MAT4x3                                     = gl.FLOAT_MAT4x3
	DOUBLE_MAT2                                      = gl.DOUBLE_MAT2
	DOUBLE_MAT3                                      = gl.DOUBLE_MAT3
	DOUBLE_MAT4                                      = gl.DOUBLE_MAT4
	DOUBLE_MAT2x3                                    = gl.DOUBLE_MAT2x3
	DOUBLE_MAT2x4                                    = gl.DOUBLE_MAT2x4
	DOUBLE_MAT3x2                                    = gl.DOUBLE_MAT3x2
	DOUBLE_MAT3x4                                    = gl.DOUBLE_MAT3x4
	DOUBLE_MAT4x2                                    = gl.DOUBLE_MAT4x2
	DOUBLE_MAT4x3                                    = gl.DOUBLE_MAT4x3
	SAMPLER_1D                                       = gl.SAMPLER_1D
	SAMPLER_2D                                       = gl.SAMPLER_2D
	SAMPLER_3D                                       = gl.SAMPLER_3D
	SAMPLER_CUBE                                     = gl.SAMPLER_CUBE
	SAMPLER_1D_SHADOW                                = gl.SAMPLER_1D_SHADOW
	SAMPLER_2D_SHADOW                                = gl.SAMPLER_2D_SHADOW
	SAMPLER_1D_ARRAY                                 = gl.SAMPLER_1D_ARRAY
	SAMPLER_2D_ARRAY                                 = gl.SAMPLER_2D_ARRAY
	SAMPLER_1D_ARRAY_SHADOW                          = gl.SAMPLER_1D_ARRAY_SHADOW
	SAMPLER_2D_ARRAY_SHADOW                          = gl.SAMPLER_2D_ARRAY_SHADOW
	SAMPLER_2D_MULTISAMPLE                           = gl.SAMPLER_2D_MULTISAMPLE
	SAMPLER_2D_MULTISAMPLE_ARRAY                     = gl.SAMPLER_2D_MULTISAMPLE_ARRAY
	SAMPLER_CUBE_SHADOW                              = gl.SAMPLER_CUBE_SHADOW
	SAMPLER_BUFFER                                   = gl.SAMPLER_BUFFER
	SAMPLER_2D_RECT                                  = gl.SAMPLER_2D_RECT
	SAMPLER_2D_RECT_SHADOW                           = gl.SAMPLER_2D_RECT_SHADOW
	INT_SAMPLER_1D                                   = gl.INT_SAMPLER_1D
	INT_SAMPLER_2D                                   = gl.INT_SAMPLER_2D
	INT_SAMPLER_3D                                   = gl.INT_SAMPLER_3D
	INT_SAMPLER_CUBE                                 = gl.INT_SAMPLER_CUBE
	INT_SAMPLER_1D_ARRAY                             = gl.INT_SAMPLER_1D_ARRAY
	INT_SAMPLER_2D_ARRAY                             = gl.INT_SAMPLER_2D_ARRAY
	INT_SAMPLER_2D_MULTISAMPLE                       = gl.INT_SAMPLER_2D_MULTISAMPLE
	INT_SAMPLER_2D_MULTISAMPLE_ARRAY                 = gl.INT_SAMPLER_2D_MULTISAMPLE_ARRAY
	INT_SAMPLER_BUFFER                               = gl.INT_SAMPLER_BUFFER
	INT_SAMPLER_2D_RECT                              = gl.INT_SAMPLER_2D_RECT
	UNSIGNED_INT_SAMPLER_1D                          = gl.UNSIGNED_INT_SAMPLER_1D
	UNSIGNED_INT_SAMPLER_2D                          = gl.UNSIGNED_INT_SAMPLER_2D
	UNSIGNED_INT_SAMPLER_3D                          = gl.UNSIGNED_INT_SAMPLER_3D
	UNSIGNED_INT_SAMPLER_CUBE                        = gl.UNSIGNED_INT_SAMPLER_CUBE
	UNSIGNED_INT_SAMPLER_1D_ARRAY                    = gl.UNSIGNED_INT_SAMPLER_1D_ARRAY
	UNSIGNED_INT_SAMPLER_2D_ARRAY                    = gl.UNSIGNED_INT_SAMPLER_2D_ARRAY
	UNSIGNED_INT_SAMPLER_2D_MULTISAMPLE              = gl.UNSIGNED_INT_SAMPLER_2D_MULTISAMPLE
	UNSIGNED_INT_SAMPLER_2D_MULTISAMPLE_ARRAY        = gl.UNSIGNED_INT_SAMPLER_2D_MULTISAMPLE_ARRAY
	UNSIGNED_INT_SAMPLER_BUFFER                      = gl.UNSIGNED_INT_SAMPLER_BUFFER
	UNSIGNED_INT_SAMPLER_2D_RECT                     = gl.UNSIGNED_INT_SAMPLER_2D_RECT
	IMAGE_1D                                         = gl.IMAGE_1D
	IMAGE_2D                                         = gl.IMAGE_2D
	IMAGE_3D                                         = gl.IMAGE_3D
	IMAGE_2D_RECT                                    = gl.IMAGE_2D_RECT
	IMAGE_CUBE                                       = gl.IMAGE_CUBE
	IMAGE_BUFFER                                     = gl.IMAGE_BUFFER
	IMAGE_1D_ARRAY                                   = gl.IMAGE_1D_ARRAY
	IMAGE_2D_ARRAY                                   = gl.IMAGE_2D_ARRAY
	IMAGE_2D_MULTISAMPLE                             = gl.IMAGE_2D_MULTISAMPLE
	IMAGE_2D_MULTISAMPLE_ARRAY                       = gl.IMAGE_2D_MULTISAMPLE_ARRAY
	INT_IMAGE_1D                                     = gl.INT_IMAGE_1D
	INT_IMAGE_2D                                     = gl.INT_IMAGE_2D
	INT_IMAGE_3D                                     = gl.INT_IMAGE_3D
	INT_IMAGE_2D_RECT                                = gl.INT_IMAGE_2D_RECT
	INT_IMAGE_CUBE                                   = gl.INT_IMAGE_CUBE
	INT_IMAGE_BUFFER                                 = gl.INT_IMAGE_BUFFER
	INT_IMAGE_1D_ARRAY                               = gl.INT_IMAGE_1D_ARRAY
	INT_IMAGE_2D_ARRAY                               = gl.INT_IMAGE_2D_ARRAY
	INT_IMAGE_2D_MULTISAMPLE                         = gl.INT_IMAGE_2D_MULTISAMPLE
	INT_IMAGE_2D_MULTISAMPLE_ARRAY                   = gl.INT_IMAGE_2D_MULTISAMPLE_ARRAY
	UNSIGNED_INT_IMAGE_1D                            = gl.UNSIGNED_INT_IMAGE_1D
	UNSIGNED_INT_IMAGE_2D                            = gl.UNSIGNED_INT_IMAGE_2D
	UNSIGNED_INT_IMAGE_3D                            = gl.UNSIGNED_INT_IMAGE_3D
	UNSIGNED_INT_IMAGE_2D_RECT                       = gl.UNSIGNED_INT_IMAGE_2D_RECT
	UNSIGNED_INT_IMAGE_CUBE                          = gl.UNSIGNED_INT_IMAGE_CUBE
	UNSIGNED_INT_IMAGE_BUFFER                        = gl.UNSIGNED_INT_IMAGE_BUFFER
	UNSIGNED_INT_IMAGE_1D_ARRAY                      = gl.UNSIGNED_INT_IMAGE_1D_ARRAY
	UNSIGNED_INT_IMAGE_2D_ARRAY                      = gl.UNSIGNED_INT_IMAGE_2D_ARRAY
	UNSIGNED_INT_IMAGE_2D_MULTISAMPLE                = gl.UNSIGNED_INT_IMAGE_2D_MULTISAMPLE
	UNSIGNED_INT_IMAGE_2D_MULTISAMPLE_ARRAY          = gl.UNSIGNED_INT_IMAGE_2D_MULTISAMPLE_ARRAY
	UNSIGNED_INT_ATOMIC_COUNTER                      = gl.UNSIGNED_INT_ATOMIC_COUNTER
)
