package main

import (
	"fmt"
	"runtime"

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

	window, err := glfw.CreateWindow(640, 480, "Testing", nil, nil)
	if err != nil {
		panic(err)
	}
	setCallbacks(window)

	window.MakeContextCurrent()

	for !window.ShouldClose() {
		// Do OpenGL stuff.
		window.SwapBuffers()
		glfw.PollEvents()
	}
}

func setCallbacks(w *glfw.Window) {
	w.SetMouseButtonCallback(fMouseButtonCb)
	w.SetKeyCallback(fKeyCb)
}

func fMouseButtonCb(w *glfw.Window, button glfw.MouseButton, action glfw.Action, mod glfw.ModifierKey) {
	fmt.Println("Mouse", button, action, mod)
}

func fKeyCb(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	fmt.Println("Key", key, scancode, action, mods)
}
