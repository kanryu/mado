// SPDX-License-Identifier: Unlicense OR MIT

//go:build ((linux && !android) || freebsd || openbsd) && !nox11 && !noopengl
// +build linux,!android freebsd openbsd
// +build !nox11
// +build !noopengl

package unix

import (
	"unsafe"

	"github.com/kanryu/mado"
	"github.com/kanryu/mado/unix/internal/egl"
)

var _ mado.Context = (*x11Context)(nil)

type x11Context struct {
	win *x11Window
	*Context
}

type PlatformContextState struct{}
type PlatformLibraryContextState struct{}

func init() {
	newX11EGLContext = func(w *x11Window) (mado.Context, error) {
		disp := egl.NativeDisplayType(unsafe.Pointer(w.display()))
		eglApi := egl.EGL_OPENGL_ES_API
		if mado.GlfwConfig.Enable {
			eglApi = egl.EGL_OPENGL_API
		}
		ctx, err := NewContext(disp, eglApi)
		if err != nil {
			return nil, err
		}
		return &x11Context{win: w, Context: ctx}, nil
	}
}

func (c *x11Context) Release() {
	if c.Context != nil {
		c.Context.Release()
		c.Context = nil
	}
}

func (c *x11Context) SwapBuffers() error {
	if c.Context != nil {
		return c.Context.SwapBuffers()
	}
	return nil
}

func (c *x11Context) SwapInterval(interval int) {
	if c.Context != nil {
		c.Context.SwapInterval(interval)
	}
}

func (c *x11Context) Refresh() error {
	c.Context.ReleaseSurface()
	win, width, height := c.win.window()
	eglSurf := egl.NativeWindowType(uintptr(win))
	if err := c.Context.CreateSurface(eglSurf, width, height); err != nil {
		return err
	}
	if err := c.Context.MakeCurrent(); err != nil {
		return err
	}
	c.Context.EnableVSync(true)
	c.Context.ReleaseCurrent()
	return nil
}

func (c *x11Context) Lock() error {
	return c.Context.MakeCurrent()
}

func (c *x11Context) Unlock() {
	c.Context.ReleaseCurrent()
}
