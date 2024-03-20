// SPDX-License-Identifier: Unlicense OR MIT

//go:build linux || windows || freebsd || openbsd
// +build linux windows freebsd openbsd

package egl

import (
	"errors"
	"fmt"
	"runtime"
	"strings"

	"github.com/kanryu/mado/gpu"
)

type Context struct {
	disp          _EGLDisplay
	eglCtx        *eglContext
	eglSurf       _EGLSurface
	width, height int
}

type eglContext struct {
	config      _EGLConfig
	ctx         _EGLContext
	visualID    int
	srgb        bool
	surfaceless bool
}

var (
	nilEGLDisplay       _EGLDisplay
	nilEGLSurface       _EGLSurface
	nilEGLContext       _EGLContext
	nilEGLConfig        _EGLConfig
	EGL_DEFAULT_DISPLAY NativeDisplayType
)

const (
	_EGL_CONFIG_CAVEAT = 0x3027

	_EGL_SUCCESS                = 0x3000
	_EGL_NOT_INITIALIZED        = 0x3001
	_EGL_BAD_ACCESS             = 0x3002
	_EGL_BAD_ALLOC              = 0x3003
	_EGL_BAD_ATTRIBUTE          = 0x3004
	_EGL_BAD_CONFIG             = 0x3005
	_EGL_BAD_CONTEXT            = 0x3006
	_EGL_BAD_CURRENT_SURFACE    = 0x3007
	_EGL_BAD_DISPLAY            = 0x3008
	_EGL_BAD_MATCH              = 0x3009
	_EGL_BAD_NATIVE_PIXMAP      = 0x300a
	_EGL_BAD_NATIVE_WINDOW      = 0x300b
	_EGL_BAD_PARAMETER          = 0x300c
	_EGL_BAD_SURFACE            = 0x300d
	_EGL_CONTEXT_LOST           = 0x300e
	_EGL_COLOR_BUFFER_TYPE      = 0x303f
	_EGL_RGB_BUFFER             = 0x308e
	_EGL_SURFACE_TYPE           = 0x3033
	_EGL_WINDOW_BIT             = 0x0004
	_EGL_RENDERABLE_TYPE        = 0x3040
	_EGL_OPENGL_ES_BIT          = 0x0001
	_EGL_OPENGL_ES2_BIT         = 0x0004
	_EGL_OPENGL_BIT             = 0x0008
	_EGL_ALPHA_SIZE             = 0x3021
	_EGL_BLUE_SIZE              = 0x3022
	_EGL_GREEN_SIZE             = 0x3023
	_EGL_RED_SIZE               = 0x3024
	_EGL_DEPTH_SIZE             = 0x3025
	_EGL_STENCIL_SIZE           = 0x3026
	_EGL_SAMPLES                = 0x3031
	_EGL_OPENGL_ES_API          = 0x30a0
	_EGL_OPENGL_API             = 0x30a2
	_EGL_NONE                   = 0x3038
	_EGL_RENDER_BUFFER          = 0x3086
	_EGL_SINGLE_BUFFER          = 0x3085
	_EGL_EXTENSIONS             = 0x3055
	_EGL_CONTEXT_CLIENT_VERSION = 0x3098
	_EGL_NATIVE_VISUAL_ID       = 0x302e
	// _EGL_NO_SURFACE ((EGLSurface) = 0)
	// _EGL_NO_DISPLAY ((EGLDisplay) = 0)
	// _EGL_NO_CONTEXT ((EGLContext) = 0)
	// _EGL_DEFAULT_DISPLAY ((EGLNativeDisplayType) = 0)

	_EGL_CONTEXT_OPENGL_FORWARD_COMPATIBLE_BIT_KHR      = 0x00000002
	_EGL_CONTEXT_OPENGL_CORE_PROFILE_BIT_KHR            = 0x00000001
	_EGL_CONTEXT_OPENGL_COMPATIBILITY_PROFILE_BIT_KHR   = 0x00000002
	_EGL_CONTEXT_OPENGL_DEBUG_BIT_KHR                   = 0x00000001
	_EGL_CONTEXT_OPENGL_RESET_NOTIFICATION_STRATEGY_KHR = 0x31bd
	_EGL_NO_RESET_NOTIFICATION_KHR                      = 0x31be
	_EGL_LOSE_CONTEXT_ON_RESET_KHR                      = 0x31bf
	_EGL_CONTEXT_OPENGL_ROBUST_ACCESS_BIT_KHR           = 0x00000004
	_EGL_CONTEXT_MAJOR_VERSION_KHR                      = 0x3098
	_EGL_CONTEXT_MINOR_VERSION_KHR                      = 0x30fb
	_EGL_CONTEXT_OPENGL_PROFILE_MASK_KHR                = 0x30fd
	_EGL_CONTEXT_FLAGS_KHR                              = 0x30fc
	_EGL_CONTEXT_OPENGL_NO_ERROR_KHR                    = 0x31b3
	_EGL_GL_COLORSPACE_KHR                              = 0x309d
	_EGL_GL_COLORSPACE_SRGB_KHR                         = 0x3089
	_EGL_CONTEXT_RELEASE_BEHAVIOR_KHR                   = 0x2097
	_EGL_CONTEXT_RELEASE_BEHAVIOR_NONE_KHR              = 0
	_EGL_CONTEXT_RELEASE_BEHAVIOR_FLUSH_KHR             = 0x2098
	_EGL_PRESENT_OPAQUE_EXT                             = 0x31df
)

func (c *Context) Release() {
	c.ReleaseSurface()
	if c.eglCtx != nil {
		eglDestroyContext(c.disp, c.eglCtx.ctx)
		c.eglCtx = nil
	}
	eglTerminate(c.disp)
	c.disp = nilEGLDisplay
}

func (c *Context) Present() error {
	if !eglSwapBuffers(c.disp, c.eglSurf) {
		return fmt.Errorf("eglSwapBuffers failed (%x)", eglGetError())
	}
	return nil
}

func (c *Context) HasSurface() bool {
	return c.eglSurf != nilEGLSurface
}

func NewContext(disp NativeDisplayType, eglApi uint) (*Context, error) {
	if err := loadEGL(); err != nil {
		return nil, err
	}
	eglDisp := eglGetDisplay(disp)
	// eglGetDisplay can return EGL_NO_DISPLAY yet no error
	// (EGL_SUCCESS), in which case a default EGL display might be
	// available.
	if eglDisp == nilEGLDisplay {
		eglDisp = eglGetDisplay(EGL_DEFAULT_DISPLAY)
	}
	if eglDisp == nilEGLDisplay {
		return nil, fmt.Errorf("eglGetDisplay failed: 0x%x", eglGetError())
	}
	eglCtx, err := createContext(eglDisp, eglApi)
	if err != nil {
		return nil, err
	}
	c := &Context{
		disp:   eglDisp,
		eglCtx: eglCtx,
	}
	return c, nil
}

func (c *Context) RenderTarget() (gpu.RenderTarget, error) {
	return gpu.OpenGLRenderTarget{}, nil
}

func (c *Context) API() gpu.API {
	return gpu.OpenGL{}
}

func (c *Context) ReleaseSurface() {
	if c.eglSurf == nilEGLSurface {
		return
	}
	// Make sure any in-flight GL commands are complete.
	eglWaitClient()
	c.ReleaseCurrent()
	eglDestroySurface(c.disp, c.eglSurf)
	c.eglSurf = nilEGLSurface
}

func (c *Context) VisualID() int {
	return c.eglCtx.visualID
}

func (c *Context) CreateSurface(win NativeWindowType, width, height int) error {
	eglSurf, err := createSurface(c.disp, c.eglCtx, win)
	c.eglSurf = eglSurf
	c.width = width
	c.height = height
	return err
}

func (c *Context) ReleaseCurrent() {
	if c.disp != nilEGLDisplay {
		eglMakeCurrent(c.disp, nilEGLSurface, nilEGLSurface, nilEGLContext)
	}
}

func (c *Context) MakeCurrent() error {
	// OpenGL contexts are implicit and thread-local. Lock the OS thread.
	runtime.LockOSThread()

	if c.eglSurf == nilEGLSurface && !c.eglCtx.surfaceless {
		return errors.New("no surface created yet EGL_KHR_surfaceless_context is not supported")
	}
	if !eglMakeCurrent(c.disp, c.eglSurf, c.eglSurf, c.eglCtx.ctx) {
		return fmt.Errorf("eglMakeCurrent error 0x%x", eglGetError())
	}
	return nil
}

func (c *Context) EnableVSync(enable bool) {
	if enable {
		eglSwapInterval(c.disp, 1)
	} else {
		eglSwapInterval(c.disp, 0)
	}
}

func (c *Context) SwapBuffers() error {
	ok := eglSwapBuffers(c.disp, c.eglSurf)
	if ok {
		return nil
	}
	fmt.Println("eglSwapBuffers", eglGetError())
	return fmt.Errorf("eglSwapBuffers returned false")
}

func (c *Context) SwapInterval(interval int) error {
	ok := eglSwapInterval(c.disp, _EGLint(interval))
	if ok {
		return nil
	}
	return fmt.Errorf("eglSwapInterval returned false")
}

func hasExtension(exts []string, ext string) bool {
	for _, e := range exts {
		if ext == e {
			return true
		}
	}
	return false
}

func createContext(disp _EGLDisplay, eglApi uint) (*eglContext, error) {
	major, minor, ret := eglInitialize(disp)
	if !ret {
		return nil, fmt.Errorf("eglInitialize failed: 0x%x", eglGetError())
	}
	// sRGB framebuffer support on EGL 1.5 or if EGL_KHR_gl_colorspace is supported.
	exts := strings.Split(eglQueryString(disp, _EGL_EXTENSIONS), " ")
	srgb := major > 1 || minor >= 5 || hasExtension(exts, "EGL_KHR_gl_colorspace")
	attribs := []_EGLint{
		_EGL_RENDERABLE_TYPE, _EGL_OPENGL_ES2_BIT,
		_EGL_SURFACE_TYPE, _EGL_WINDOW_BIT,
		_EGL_BLUE_SIZE, 8,
		_EGL_GREEN_SIZE, 8,
		_EGL_RED_SIZE, 8,
		_EGL_CONFIG_CAVEAT, _EGL_NONE,
	}
	if srgb {
		if runtime.GOOS == "linux" || runtime.GOOS == "android" {
			// Some Mesa drivers crash if an sRGB framebuffer is requested without alpha.
			// https://bugs.freedesktop.org/show_bug.cgi?id=107782.
			//
			// Also, some Android devices (Samsung S9) need alpha for sRGB to work.
			attribs = append(attribs, _EGL_ALPHA_SIZE, 8)
		}
	}
	attribs = append(attribs, _EGL_NONE)

	ret = eglBindAPI(eglApi)
	if !ret {
		return nil, fmt.Errorf("eglBindAPI failed: 0x%x", eglGetError())
	}

	eglCfg, ret := eglChooseConfig(disp, attribs)
	if !ret {
		return nil, fmt.Errorf("eglChooseConfig failed: 0x%x", eglGetError())
	}
	if eglCfg == nilEGLConfig {
		supportsNoCfg := hasExtension(exts, "EGL_KHR_no_config_context")
		if !supportsNoCfg {
			return nil, errors.New("eglChooseConfig returned no configs")
		}
	}
	var visID _EGLint
	if eglCfg != nilEGLConfig {
		var ok bool
		visID, ok = eglGetConfigAttrib(disp, eglCfg, _EGL_NATIVE_VISUAL_ID)
		if !ok {
			return nil, errors.New("newContext: eglGetConfigAttrib for _EGL_NATIVE_VISUAL_ID failed")
		}
	}
	ctxAttribs := []_EGLint{
		_EGL_CONTEXT_CLIENT_VERSION, 3,
		_EGL_NONE,
	}
	eglCtx := eglCreateContext(disp, eglCfg, nilEGLContext, ctxAttribs)
	if eglCtx == nilEGLContext {
		// Fall back to OpenGL ES 2 and rely on extensions.
		ctxAttribs := []_EGLint{
			_EGL_CONTEXT_CLIENT_VERSION, 2,
			_EGL_NONE,
		}
		eglCtx = eglCreateContext(disp, eglCfg, nilEGLContext, ctxAttribs)
		if eglCtx == nilEGLContext {
			return nil, fmt.Errorf("eglCreateContext failed: 0x%x", eglGetError())
		}
	}
	return &eglContext{
		config:      _EGLConfig(eglCfg),
		ctx:         _EGLContext(eglCtx),
		visualID:    int(visID),
		srgb:        srgb,
		surfaceless: hasExtension(exts, "EGL_KHR_surfaceless_context"),
	}, nil
}

func createSurface(disp _EGLDisplay, eglCtx *eglContext, win NativeWindowType) (_EGLSurface, error) {
	var surfAttribs []_EGLint
	if eglCtx.srgb {
		surfAttribs = append(surfAttribs, _EGL_GL_COLORSPACE_KHR, _EGL_GL_COLORSPACE_SRGB_KHR)
	}
	surfAttribs = append(surfAttribs, _EGL_NONE)
	eglSurf := eglCreateWindowSurface(disp, eglCtx.config, win, surfAttribs)
	if eglSurf == nilEGLSurface && eglCtx.srgb {
		// Try again without sRGB.
		eglCtx.srgb = false
		surfAttribs = []_EGLint{_EGL_NONE}
		eglSurf = eglCreateWindowSurface(disp, eglCtx.config, win, surfAttribs)
	}
	if eglSurf == nilEGLSurface {
		return nilEGLSurface, fmt.Errorf("newContext: eglCreateWindowSurface failed 0x%x (sRGB=%v)", eglGetError(), eglCtx.srgb)
	}
	return eglSurf, nil
}
