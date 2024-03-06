package main

import (
	"fmt"
	"runtime"

	"github.com/go-gl/gl/v2.1/gl"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/kanryu/mado/glfw"
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
	// glfw.WindowHint(glfw.ClientAPI, glfw.OpenGLAPI)
	// glfw.WindowHint(glfw.ContextCreationAPI, glfw.NativeContextAPI)
	// glfw.WindowHint(glfw.ContextVersionMajor, 2)
	// glfw.WindowHint(glfw.ContextVersionMinor, 0)

	window, err := glfw.CreateWindow(640, 480, "Testing", nil, nil)
	if err != nil {
		panic(err)
	}
	//setCallbacks(window)

	window.MakeContextCurrent()
	glfw.SwapInterval(1)

	if err := gl.Init(); err != nil {
		panic(err)
	}

	for !window.ShouldClose() {
		// if ok, err := window.ShouldClose(); err != nil {
		// 	panic(err)
		// } else if ok {
		// 	break
		// }
		t := glfw.GetTime()
		fmt.Println("GetTime", t)
		width, height := window.GetFramebufferSize()
		gl.Viewport(0, 0, int32(width), int32(height))
		fmt.Println("Viewport", width, height, gl.GetError())
		// Clear color buffer to black
		gl.ClearColor(0.0, 0.0, 0.0, 0.0)
		fmt.Println("ClearColor", gl.GetError())
		gl.Clear(gl.COLOR_BUFFER_BIT)
		fmt.Println("Clear", gl.GetError())

		// Select and setup the projection matrix
		gl.MatrixMode(gl.PROJECTION)
		fmt.Println("MatrixMode", gl.GetError())
		gl.LoadIdentity()
		fmt.Println("LoadIdentity", gl.GetError())
		m1 := mgl32.Perspective(65.0, float32(width)/float32(height), 1.0, 100.0)
		gl.LoadMatrixf(&m1[0])
		fmt.Println("LoadMatrixf", gl.GetError())

		// Select and setup the modelview matrix
		gl.MatrixMode(gl.MODELVIEW)
		fmt.Println("MatrixMode", gl.GetError())
		gl.LoadIdentity()
		fmt.Println("LoadIdentity", gl.GetError())
		m2 := mgl32.LookAt(0.0, 5.0, 0.0, // Eye-position
			0.0, 20.0, 0.0, // View-point
			0.0, 0.0, 1.0) // Up-vector
		gl.LoadMatrixf(&m2[0])
		fmt.Println("LoadMatrixf", gl.GetError())

		// Draw a rotating colorful triangle
		gl.Translatef(0.0, 14.0, 0.0)
		fmt.Println("Translatef", gl.GetError())
		gl.Rotatef(float32(t*100.0), 0.0, 0.0, 1.0)
		fmt.Println("Rotatef", gl.GetError())

		gl.Begin(gl.TRIANGLES)
		fmt.Println("Begin", gl.GetError())
		gl.Color3f(1.0, 0.0, 0.0)
		gl.Vertex3f(-5.0, 0.0, -4.0)
		gl.Color3f(0.0, 1.0, 0.0)
		gl.Vertex3f(5.0, 0.0, -4.0)
		gl.Color3f(0.0, 0.0, 1.0)
		gl.Vertex3f(0.0, 0.0, 6.0)
		gl.End()
		fmt.Println("End", gl.GetError())

		// Do OpenGL stuff.
		window.SwapBuffers()
		glfw.PollEvents()
	}
}
