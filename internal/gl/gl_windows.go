// SPDX-License-Identifier: Unlicense OR MIT

package gl

import (
	"errors"
	"fmt"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	OpenGL32              = windows.NewLazyDLL("opengl32.dll")
	_wglCreateContext     = OpenGL32.NewProc("wglCreateContext")
	_wglDeleteContext     = OpenGL32.NewProc("wglDeleteContext")
	_wglGetProcAddress    = OpenGL32.NewProc("wglGetProcAddress")
	_wglGetCurrentDC      = OpenGL32.NewProc("wglGetCurrentDC")
	_wglGetCurrentContext = OpenGL32.NewProc("wglGetCurrentContext")
	_wglMakeCurrent       = OpenGL32.NewProc("wglMakeCurrent")
	_wglShareLists        = OpenGL32.NewProc("wglShareLists")

	procWGLCreateContextAttribsARB   uintptr
	procWGLGetExtensionsStringARB    uintptr
	procWGLGetExtensionsStringEXT    uintptr
	procWGLGetPixelFormatAttribivARB uintptr
	procWGLSwapIntervalEXT           uintptr
)

func InitWGLExtensionFunctions() {
	procWGLCreateContextAttribsARB = WglGetProcAddress("wglCreateContextAttribsARB")
	procWGLGetExtensionsStringARB = WglGetProcAddress("wglGetExtensionsStringARB")
	procWGLGetExtensionsStringEXT = WglGetProcAddress("wglGetExtensionsStringEXT")
	procWGLGetPixelFormatAttribivARB = WglGetProcAddress("wglGetPixelFormatAttribivARB")
	procWGLSwapIntervalEXT = WglGetProcAddress("wglSwapIntervalEXT")
}

func WglCreateContext(hdc windows.Handle) (windows.Handle, error) {
	r, _, e := _wglCreateContext.Call(uintptr(hdc))
	if r == 0 {
		return 0, fmt.Errorf("glfw: wglCreateContext failed: %w", e)
	}
	return windows.Handle(r), nil
}

func WglDeleteContext(hglrc windows.Handle) error {
	r, _, e := _wglDeleteContext.Call(uintptr(hglrc))
	if int32(r) == 0 && !errors.Is(e, windows.ERROR_SUCCESS) {
		return fmt.Errorf("glfw: wglDeleteContext failed: %w", e)
	}
	return nil
}

func WglGetProcAddress(procname string) uintptr {
	ptr, err := windows.BytePtrFromString(procname)
	if err != nil {
		panic("glfw: unnamedParam1 must not include a NUL character")
	}
	r, _, _ := _wglGetProcAddress.Call(uintptr(unsafe.Pointer(ptr)))
	return r
}

func WglGetCurrentDC() windows.Handle {
	r, _, _ := _wglGetCurrentDC.Call()
	return windows.Handle(r)
}

func WglGetCurrentContext() windows.Handle {
	r, _, _ := _wglGetCurrentContext.Call()
	return windows.Handle(r)
}

func WglMakeCurrent(hdc windows.Handle, hglrc windows.Handle) error {
	r, _, e := _wglMakeCurrent.Call(uintptr(hdc), uintptr(hglrc))
	if int32(r) == 0 && !errors.Is(e, windows.ERROR_SUCCESS) {
		return fmt.Errorf("wglMakeCurrent failed: %w", e)
	}
	return nil
}

func WglShareListss(hglrc1 windows.Handle, hglrc2 windows.Handle) error {
	r, _, e := _wglShareLists.Call(uintptr(hglrc1), uintptr(hglrc2))
	if int32(r) == 0 && !errors.Is(e, windows.ERROR_SUCCESS) {
		return fmt.Errorf("wglShareLists failed: %w", e)
	}
	return nil
}

// extensions API

func WglCreateContextAttribsARB(hDC windows.Handle, hshareContext windows.Handle, attribList *int32) (windows.Handle, error) {
	r, _, e := syscall.Syscall(procWGLCreateContextAttribsARB, 3, uintptr(hDC), uintptr(hshareContext), uintptr(unsafe.Pointer(attribList)))
	if windows.Handle(r) == 0 {
		// TODO: Show more detailed error? See the original implementation.
		return 0, fmt.Errorf("wglCreateContextAttribsARB failed: %w", e)
	}
	return windows.Handle(r), nil
}

func WglGetExtensionsStringARB(hdc windows.Handle) string {
	r, _, _ := syscall.Syscall(procWGLGetExtensionsStringARB, 1, uintptr(hdc), 0, 0)
	return windows.BytePtrToString((*byte)(unsafe.Pointer(r)))
}

func WglGetExtensionsStringARB_Available() bool {
	return procWGLGetExtensionsStringARB != 0
}

func WglGetExtensionsStringEXT() string {
	r, _, _ := syscall.Syscall(procWGLGetExtensionsStringEXT, 0, 0, 0, 0)
	return windows.BytePtrToString((*byte)(unsafe.Pointer(r)))
}

func WglGetExtensionsStringEXT_Available() bool {
	return procWGLGetExtensionsStringEXT != 0
}

func WglGetPixelFormatAttribivARB(hdc windows.Handle, iPixelFormat int32, iLayerPlane int32, nAttributes uint32, piAttributes *int32, piValues *int32) error {
	r, _, e := syscall.Syscall6(procWGLGetPixelFormatAttribivARB, 6, uintptr(hdc), uintptr(iPixelFormat), uintptr(iLayerPlane), uintptr(nAttributes), uintptr(unsafe.Pointer(piAttributes)), uintptr(unsafe.Pointer(piValues)))
	if int32(r) == 0 && !errors.Is(e, windows.ERROR_SUCCESS) {
		return fmt.Errorf("wglGetPixelFormatAttribivARB failed: %w", e)
	}
	return nil
}

func WglSwapIntervalEXT(interval int32) error {
	r, _, e := syscall.Syscall(procWGLSwapIntervalEXT, 1, uintptr(interval), 0, 0)
	if int32(r) == 0 && !errors.Is(e, windows.ERROR_SUCCESS) {
		return fmt.Errorf("wglSwapIntervalEXT failed: %w", e)
	}
	return nil
}

func GetProcAddressWGL(procname string) uintptr {
	proc := WglGetProcAddress(procname)
	if proc != 0 {
		return proc
	}
	return OpenGL32.NewProc(procname).Addr()
}
