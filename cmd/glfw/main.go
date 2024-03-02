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
	// window callbacks
	w.SetPosCallback(fPosCallback)
	w.SetSizeCallback(fSizeCallback)
	w.SetFramebufferSizeCallback(fFramebufferSizeCallback)
	w.SetCloseCallback(fCloseCallback)
	w.SetMaximizeCallback(fMaximizeCallback)
	w.SetContentScaleCallback(fContentScaleCallback)
	w.SetRefreshCallback(fRefreshCallback)
	w.SetFocusCallback(fFocusCallback)
	w.SetIconifyCallback(fIconifyCallback)

	// input callbacks
	//	w.SetMouseButtonCallback(fMouseButtonCallback)
	//	w.SetCursorPosCallback(fCursorPosCallback)
	w.SetCursorEnterCallback(fCursorEnterCallback)
	w.SetScrollCallback(fScrollCallback)
	w.SetCharCallback(fCharCallback)
	w.SetCharModsCallback(fCharModsCallback)
	w.SetDropCallback(fDropCallback)
	w.SetKeyCallback(fKeyCallback)

	// ime input callbacks
	w.SetPreeditCallback(fPreeditCallback)
	w.SetImeStatusCallback(fImeStatusCallback)
	w.SetPreeditCandidateCallback(fPreeditCandidateHolder)
}

func fPosCallback(w *glfw.Window, xpos int, ypos int) {
	fmt.Println("Pos", xpos, ypos)

}
func fSizeCallback(w *glfw.Window, width int, height int) {
	fmt.Println("Size", width, height)

}
func fFramebufferSizeCallback(w *glfw.Window, width int, height int) {
	fmt.Println("FramebufferSize", width, height)

}
func fCloseCallback(w *glfw.Window) {
	fmt.Println("Close")

}
func fMaximizeCallback(w *glfw.Window, maximized bool) {
	fmt.Println("Maximize", maximized)

}
func fContentScaleCallback(w *glfw.Window, x float32, y float32) {
	fmt.Println("ContentScale", x, y)

}
func fRefreshCallback(w *glfw.Window) {
	fmt.Println("Refresh")

}
func fFocusCallback(w *glfw.Window, focused bool) {
	fmt.Println("Focus", focused)

}
func fIconifyCallback(w *glfw.Window, iconified bool) {
	fmt.Println("Iconify", iconified)

}

// input
func fMouseButtonCallback(w *glfw.Window, button glfw.MouseButton, action glfw.Action, mod glfw.ModifierKey) {
	fmt.Println("Mouse", button, action, mod)
}
func fCursorPosCallback(w *glfw.Window, xpos float64, ypos float64) {
	fmt.Println("CursorPos")
}
func fCursorEnterCallback(w *glfw.Window, entered bool) {
	fmt.Println("CursorEnter")
}
func fScrollCallback(w *glfw.Window, xoff float64, yoff float64) {
	fmt.Println("Scroll")
}
func fCharCallback(w *glfw.Window, char rune) {
	fmt.Println("Char", char)
}
func fCharModsCallback(w *glfw.Window, char rune, mods glfw.ModifierKey) {
	fmt.Println("CharMods", char, mods)
}
func fDropCallback(w *glfw.Window, names []string) {
	fmt.Println("Drop", names)
}
func fKeyCallback(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	fmt.Println("Key", key, scancode, action, mods)
}

// ime input
func fPreeditCallback(
	w *glfw.Window,
	preeditCount int,
	preeditString string,
	blockCount int,
	blockSizes string,
	focusedBlock int,
	caret int,
) {
	fmt.Println("Preedit")
}
func fImeStatusCallback(w *glfw.Window) {
	fmt.Println("ImeStatus")
}
func fPreeditCandidateHolder(
	w *glfw.Window,
	candidatesCount int,
	selectedIndex int,
	pageStart int,
	pageSize int,
) {
	fmt.Println("PreeditCandidate")
}
