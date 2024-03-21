// SPDX-License-Identifier: Unlicense OR MIT

//go:build ((linux && !android) || freebsd || openbsd) && !nox11 && !noopengl
// +build linux,!android freebsd openbsd
// +build !nox11
// +build !noopengl

package unix

import (
	"errors"
	"fmt"
	"runtime"
	"strings"

	"github.com/kanryu/mado"
	"github.com/kanryu/mado/gpu"
	"github.com/kanryu/mado/unix/internal/egl"
)

type Context struct {
	disp          egl.EGLDisplay
	eglCtx        *eglContext
	eglSurf       egl.EGLSurface
	attribs       []egl.EGLint
	width, height int
}

type eglContext struct {
	config      egl.EGLConfig
	ctx         egl.EGLContext
	visualID    int
	srgb        bool
	surfaceless bool
}

func (c *Context) Release() {
	c.ReleaseSurface()
	if c.eglCtx != nil {
		egl.EglDestroyContext(c.disp, c.eglCtx.ctx)
		c.eglCtx = nil
	}
	egl.EglTerminate(c.disp)
	c.disp = egl.NilEGLDisplay
}

func (c *Context) Present() error {
	if !egl.EglSwapBuffers(c.disp, c.eglSurf) {
		return fmt.Errorf("eglSwapBuffers failed (%x)", egl.EglGetError())
	}
	return nil
}

func (c *Context) HasSurface() bool {
	return c.eglSurf != egl.NilEGLSurface
}

func NewContext(disp egl.NativeDisplayType, eglApi uint) (*Context, error) {
	if err := egl.LoadEGL(); err != nil {
		return nil, err
	}
	eglDisp := egl.EglGetDisplay(disp)
	// eglGetDisplay can return EGL_NO_DISPLAY yet no error
	// (EGL_SUCCESS), in which case a default EGL display might be
	// available.
	if eglDisp == egl.NilEGLDisplay {
		eglDisp = egl.EglGetDisplay(egl.EGL_DEFAULT_DISPLAY)
	}
	if eglDisp == egl.NilEGLDisplay {
		return nil, fmt.Errorf("eglGetDisplay failed: 0x%x", egl.EglGetError())
	}
	var eglCtx *eglContext
	var attribs []egl.EGLint
	var err error
	if mado.GlfwConfig.Enable {
		eglCtx, attribs, err = glfwCreateContext(eglDisp, eglApi)

	} else {
		eglCtx, err = createContext(eglDisp)
	}
	if err != nil {
		return nil, err
	}
	c := &Context{
		disp:    eglDisp,
		eglCtx:  eglCtx,
		attribs: attribs,
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
	if c.eglSurf == egl.NilEGLSurface {
		return
	}
	// Make sure any in-flight GL commands are complete.
	egl.EglWaitClient()
	c.ReleaseCurrent()
	egl.EglDestroySurface(c.disp, c.eglSurf)
	c.eglSurf = egl.NilEGLSurface
}

func (c *Context) VisualID() int {
	return c.eglCtx.visualID
}

func (c *Context) CreateSurface(win egl.NativeWindowType, width, height int) error {
	eglSurf, err := createSurface(c.disp, c.eglCtx, c.attribs, win)
	c.eglSurf = eglSurf
	c.width = width
	c.height = height
	return err
}

func (c *Context) ReleaseCurrent() {
	if c.disp != egl.NilEGLDisplay {
		egl.EglMakeCurrent(c.disp, egl.NilEGLSurface, egl.NilEGLSurface, egl.NilEGLContext)
	}
}

func (c *Context) MakeCurrent() error {
	// OpenGL contexts are implicit and thread-local. Lock the OS thread.
	runtime.LockOSThread()

	if c.eglSurf == egl.NilEGLSurface && !c.eglCtx.surfaceless {
		return errors.New("no surface created yet EGL_KHR_surfaceless_context is not supported")
	}
	if !egl.EglMakeCurrent(c.disp, c.eglSurf, c.eglSurf, c.eglCtx.ctx) {
		return fmt.Errorf("eglMakeCurrent error 0x%x", egl.EglGetError())
	}
	return nil
}

func (c *Context) EnableVSync(enable bool) {
	if enable {
		egl.EglSwapInterval(c.disp, 1)
	} else {
		egl.EglSwapInterval(c.disp, 0)
	}
}

func (c *Context) SwapBuffers() error {
	ok := egl.EglSwapBuffers(c.disp, c.eglSurf)
	if ok {
		return nil
	}
	fmt.Println("eglSwapBuffers", egl.EglGetError())
	return fmt.Errorf("eglSwapBuffers returned false")
}

func (c *Context) SwapInterval(interval int) error {
	ok := egl.EglSwapInterval(c.disp, egl.EGLint(interval))
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

func createContext(disp egl.EGLDisplay) (*eglContext, error) {
	major, minor, ret := egl.EglInitialize(disp)
	if !ret {
		return nil, fmt.Errorf("eglInitialize failed: 0x%x", egl.EglGetError())
	}
	// sRGB framebuffer support on EGL 1.5 or if EGL_KHR_gl_colorspace is supported.
	exts := strings.Split(egl.EglQueryString(disp, egl.EGL_EXTENSIONS), " ")
	srgb := major > 1 || minor >= 5 || hasExtension(exts, "EGL_KHR_gl_colorspace")
	attribs := []egl.EGLint{
		egl.EGL_RENDERABLE_TYPE, egl.EGL_OPENGL_ES2_BIT,
		egl.EGL_SURFACE_TYPE, egl.EGL_WINDOW_BIT,
		egl.EGL_BLUE_SIZE, 8,
		egl.EGL_GREEN_SIZE, 8,
		egl.EGL_RED_SIZE, 8,
		egl.EGL_CONFIG_CAVEAT, egl.EGL_NONE,
	}
	if srgb {
		if runtime.GOOS == "linux" || runtime.GOOS == "android" {
			// Some Mesa drivers crash if an sRGB framebuffer is requested without alpha.
			// https://bugs.freedesktop.org/show_bug.cgi?id=107782.
			//
			// Also, some Android devices (Samsung S9) need alpha for sRGB to work.
			attribs = append(attribs, egl.EGL_ALPHA_SIZE, 8)
		}
	}
	attribs = append(attribs, egl.EGL_NONE)

	eglCfg, ret := egl.EglChooseConfig(disp, attribs)
	if !ret {
		return nil, fmt.Errorf("eglChooseConfig failed: 0x%x", egl.EglGetError())
	}
	if eglCfg == egl.NilEGLConfig {
		supportsNoCfg := hasExtension(exts, "EGL_KHR_no_config_context")
		if !supportsNoCfg {
			return nil, errors.New("eglChooseConfig returned no configs")
		}
	}
	var visID egl.EGLint
	if eglCfg != egl.NilEGLConfig {
		var ok bool
		visID, ok = egl.EglGetConfigAttrib(disp, eglCfg, egl.EGL_NATIVE_VISUAL_ID)
		if !ok {
			return nil, errors.New("newContext: eglGetConfigAttrib for _EGL_NATIVE_VISUAL_ID failed")
		}
	}
	ctxAttribs := []egl.EGLint{
		egl.EGL_CONTEXT_CLIENT_VERSION, 3,
		egl.EGL_NONE,
	}
	eglCtx := egl.EglCreateContext(disp, eglCfg, egl.NilEGLContext, ctxAttribs)
	if eglCtx == egl.NilEGLContext {
		// Fall back to OpenGL ES 2 and rely on extensions.
		ctxAttribs := []egl.EGLint{
			egl.EGL_CONTEXT_CLIENT_VERSION, 2,
			egl.EGL_NONE,
		}
		eglCtx = egl.EglCreateContext(disp, eglCfg, egl.NilEGLContext, ctxAttribs)
		if eglCtx == egl.NilEGLContext {
			return nil, fmt.Errorf("eglCreateContext failed: 0x%x", egl.EglGetError())
		}
	}
	return &eglContext{
		config:      egl.EGLConfig(eglCfg),
		ctx:         egl.EGLContext(eglCtx),
		visualID:    int(visID),
		srgb:        srgb,
		surfaceless: hasExtension(exts, "EGL_KHR_surfaceless_context"),
	}, nil
}

func glfwCreateContext(disp egl.EGLDisplay, eglApi uint) (*eglContext, []egl.EGLint, error) {
	major, minor, ret := egl.EglInitialize(disp)
	if !ret {
		return nil, nil, fmt.Errorf("eglInitialize failed: 0x%x", egl.EglGetError())
	}
	// sRGB framebuffer support on EGL 1.5 or if EGL_KHR_gl_colorspace is supported.
	exts := strings.Split(egl.EglQueryString(disp, egl.EGL_EXTENSIONS), " ")
	srgb := major > 1 || minor >= 5 || hasExtension(exts, "EGL_KHR_gl_colorspace")
	mado.GlfwConfig.PlatformContext.Major = int(major)
	mado.GlfwConfig.PlatformContext.Minor = int(minor)
	mado.GlfwConfig.PlatformContext.EGL_KHR_create_context = hasExtension(exts, "EGL_KHR_create_context")
	mado.GlfwConfig.PlatformContext.KHR_create_context_no_error = hasExtension(exts, "KHR_create_context_no_error")
	mado.GlfwConfig.PlatformContext.KHR_gl_colorspace = hasExtension(exts, "KHR_gl_colorspace")
	mado.GlfwConfig.PlatformContext.KHR_get_all_proc_addresses = hasExtension(exts, "KHR_get_all_proc_addresses")
	mado.GlfwConfig.PlatformContext.KHR_context_flush_control = hasExtension(exts, "KHR_context_flush_control")
	mado.GlfwConfig.PlatformContext.EXT_present_opaque = hasExtension(exts, "EXT_present_opaque")
	cfg := mado.GlfwConfig.PlatformContext
	ctxconfig := &mado.GlfwConfig.Hints.Context
	fbconfig := &mado.GlfwConfig.Hints.Framebuffer

	var eglCfg egl.EGLConfig

	if _config, err := chooseEGLConfig(disp, ctxconfig, fbconfig); err != nil {
		return nil, nil, fmt.Errorf("EGL: Failed to find a suitable EGLConfig")
	} else {
		eglCfg = _config
	}

	if ok := egl.EglBindAPI(eglApi); !ok {
		return nil, nil, errors.New("eglBindAPI: bind EGL Api failed")
	}

	var attribs []egl.EGLint
	if cfg.EGL_KHR_create_context {
		mask := 0
		flags := 0

		if ctxconfig.Client == mado.GLFW_OPENGL_API {
			if ctxconfig.Forward {
				flags |= egl.EGL_CONTEXT_OPENGL_FORWARD_COMPATIBLE_BIT_KHR
			}
			if ctxconfig.Profile == mado.GLFW_OPENGL_CORE_PROFILE {
				mask |= egl.EGL_CONTEXT_OPENGL_CORE_PROFILE_BIT_KHR
			} else if ctxconfig.Profile == mado.GLFW_OPENGL_COMPAT_PROFILE {
				mask |= egl.EGL_CONTEXT_OPENGL_COMPATIBILITY_PROFILE_BIT_KHR
			}
		}

		if ctxconfig.Debug {
			flags |= egl.EGL_CONTEXT_OPENGL_DEBUG_BIT_KHR
		}

		if ctxconfig.Robustness != 0 {
			if ctxconfig.Robustness == mado.GLFW_NO_RESET_NOTIFICATION {
				attribs = append(attribs, egl.EGL_CONTEXT_OPENGL_RESET_NOTIFICATION_STRATEGY_KHR,
					egl.EGL_NO_RESET_NOTIFICATION_KHR)
			} else if ctxconfig.Robustness == mado.GLFW_LOSE_CONTEXT_ON_RESET {
				attribs = append(attribs, egl.EGL_CONTEXT_OPENGL_RESET_NOTIFICATION_STRATEGY_KHR,
					egl.EGL_LOSE_CONTEXT_ON_RESET_KHR)
			}

			flags |= egl.EGL_CONTEXT_OPENGL_ROBUST_ACCESS_BIT_KHR
		}

		if ctxconfig.Major != 1 || ctxconfig.Minor != 0 {
			attribs = append(attribs, egl.EGL_CONTEXT_MAJOR_VERSION_KHR, egl.EGLint(ctxconfig.Major))
			attribs = append(attribs, egl.EGL_CONTEXT_MINOR_VERSION_KHR, egl.EGLint(ctxconfig.Minor))
		}

		if ctxconfig.Noerror {
			if cfg.KHR_create_context_no_error {
				flags |= egl.EGL_CONTEXT_OPENGL_NO_ERROR_KHR
			}
		}

		if mask != 0 {
			attribs = append(attribs, egl.EGL_CONTEXT_OPENGL_PROFILE_MASK_KHR, egl.EGLint(mask))

		}

		if flags != 0 {
			attribs = append(attribs, egl.EGL_CONTEXT_FLAGS_KHR, egl.EGLint(flags))
		}
	} else {
		if ctxconfig.Client == mado.GLFW_OPENGL_ES_API {
			attribs = append(attribs, egl.EGL_CONTEXT_CLIENT_VERSION, egl.EGLint(ctxconfig.Major))
		}
	}

	if cfg.KHR_context_flush_control {
		if ctxconfig.Release == mado.GLFW_RELEASE_BEHAVIOR_NONE {
			attribs = append(attribs, egl.EGL_CONTEXT_RELEASE_BEHAVIOR_KHR,
				egl.EGL_CONTEXT_RELEASE_BEHAVIOR_NONE_KHR)
		} else if ctxconfig.Release == mado.GLFW_RELEASE_BEHAVIOR_FLUSH {
			attribs = append(attribs, egl.EGL_CONTEXT_RELEASE_BEHAVIOR_KHR,
				egl.EGL_CONTEXT_RELEASE_BEHAVIOR_FLUSH_KHR)
		}
	}

	attribs = append(attribs, egl.EGL_NONE, egl.EGL_NONE)

	var visID egl.EGLint
	if eglCfg != egl.NilEGLConfig {
		var ok bool
		visID, ok = egl.EglGetConfigAttrib(disp, eglCfg, egl.EGL_NATIVE_VISUAL_ID)
		if !ok {
			return nil, nil, errors.New("newContext: eglGetConfigAttrib for _EGL_NATIVE_VISUAL_ID failed")
		}
	}
	eglCtx := egl.EglCreateContext(disp, eglCfg, egl.NilEGLContext, attribs)

	if eglCtx == egl.NilEGLContext {
		return nil, nil, fmt.Errorf("EGL: Failed to create context: %d", egl.EglGetError())
	}

	// // Set up attributes for surface creation
	attribs = []egl.EGLint{}

	if fbconfig.SRGB {
		if cfg.KHR_gl_colorspace {
			attribs = append(attribs, egl.EGL_GL_COLORSPACE_KHR, egl.EGL_GL_COLORSPACE_SRGB_KHR)
		}
	}

	if !fbconfig.Doublebuffer {
		attribs = append(attribs, egl.EGL_RENDER_BUFFER, egl.EGL_SINGLE_BUFFER)
	}

	// _GLFW_WAYLAND
	if mado.GlfwConfig.WindowType == mado.WindowTypeWayland {
		if cfg.EXT_present_opaque {
			var transparentInt egl.EGLint
			if !fbconfig.Transparent {
				transparentInt = 1
			}
			attribs = append(attribs, egl.EGL_PRESENT_OPAQUE_EXT, transparentInt)
		}
	}

	attribs = append(attribs, egl.EGL_NONE, egl.EGL_NONE)

	// // Load the appropriate client library
	// if !cfg.EGL_KKHR_get_all_proc_addresses {
	// }

	return &eglContext{
		config:      egl.EGLConfig(eglCfg),
		ctx:         egl.EGLContext(eglCtx),
		visualID:    int(visID),
		srgb:        srgb,
		surfaceless: hasExtension(exts, "EGL_KHR_surfaceless_context"),
	}, attribs, nil
}

func chooseEGLConfig(disp egl.EGLDisplay, ctxconfig *mado.CtxConfig, fbconfig *mado.FbConfig) (egl.EGLConfig, error) {
	var wrongApiAvailable bool
	apiBit := egl.EGL_OPENGL_BIT
	if ctxconfig.Client == mado.GLFW_OPENGL_ES_API {
		if ctxconfig.Major == 1 {
			apiBit = egl.EGL_OPENGL_ES_BIT

		} else {
			apiBit = egl.EGL_OPENGL_ES2_BIT
		}
	}

	if fbconfig.Stereo {
		return egl.EGLConfig(0), fmt.Errorf("EGL: Stereo rendering not supported")
	}

	var nativeCount int
	egl.EglGetConfigs(disp, nil, 0, &nativeCount)
	if nativeCount == 0 {
		return egl.EGLConfig(0), fmt.Errorf("EGL: No EGLConfigs returned")

	}

	nativeConfigs := make([]egl.EGLConfig, nativeCount)
	egl.EglGetConfigs(disp, nativeConfigs, nativeCount, &nativeCount)

	usableConfigs := []*mado.FbConfig{}

	for i := 0; i < nativeCount; i++ {
		n := nativeConfigs[i]
		u := &mado.FbConfig{}

		// Only consider RGB(A) EGLConfigs
		if val, ok := egl.EglGetConfigAttrib(disp, n, egl.EGL_COLOR_BUFFER_TYPE); ok {
			if val != egl.EGL_RGB_BUFFER {
				continue
			}
		}

		// Only consider window EGLConfigs
		if val, ok := egl.EglGetConfigAttrib(disp, n, egl.EGL_SURFACE_TYPE); ok {
			if val&egl.EGL_WINDOW_BIT == 0 {
				continue
			}
		}

		// _GLFW_X11
		if mado.GlfwConfig.WindowType == mado.WindowTypeX11 {
			// Only consider EGLConfigs with associated Visuals
			if val, ok := egl.EglGetConfigAttrib(disp, n, egl.EGL_NATIVE_VISUAL_ID); ok {
				if val == 0 {
					continue
				}
			}
		}

		if ctxconfig.Client == mado.GLFW_OPENGL_ES_API {
			if ctxconfig.Major == 1 {
				if val, ok := egl.EglGetConfigAttrib(disp, n, egl.EGL_RENDERABLE_TYPE); ok {
					if val&egl.EGL_OPENGL_ES_BIT == 0 {
						continue
					}
				}
			} else {
				if val, ok := egl.EglGetConfigAttrib(disp, n, egl.EGL_RENDERABLE_TYPE); ok {
					if val&egl.EGL_OPENGL_ES2_BIT == 0 {
						continue
					}
				}
			}
		} else if ctxconfig.Client == mado.GLFW_OPENGL_API {
			if val, ok := egl.EglGetConfigAttrib(disp, n, egl.EGL_RENDERABLE_TYPE); ok {
				if val&egl.EGL_OPENGL_BIT == 0 {
					continue
				}
			}
		}

		if val, ok := egl.EglGetConfigAttrib(disp, n, egl.EGL_RENDERABLE_TYPE); ok {
			if int(val)&apiBit == 0 {
				wrongApiAvailable = true
				continue
			}
		}

		if val, ok := egl.EglGetConfigAttrib(disp, n, egl.EGL_RED_SIZE); ok {
			u.RedBits = int(val)
		}
		if val, ok := egl.EglGetConfigAttrib(disp, n, egl.EGL_GREEN_SIZE); ok {
			u.GreenBits = int(val)
		}
		if val, ok := egl.EglGetConfigAttrib(disp, n, egl.EGL_BLUE_SIZE); ok {
			u.BlueBits = int(val)
		}
		if val, ok := egl.EglGetConfigAttrib(disp, n, egl.EGL_ALPHA_SIZE); ok {
			u.AlphaBits = int(val)
		}
		if val, ok := egl.EglGetConfigAttrib(disp, n, egl.EGL_DEPTH_SIZE); ok {
			u.DepthBits = int(val)
		}
		if val, ok := egl.EglGetConfigAttrib(disp, n, egl.EGL_STENCIL_SIZE); ok {
			u.StencilBits = int(val)
		}

		// GLFW_WAYLAND
		if mado.GlfwConfig.WindowType == mado.WindowTypeWayland {
			// NOTE: The wl_surface opaque region is no guarantee that its buffer
			//       is presented as opaque, if it also has an alpha channel
			// HACK: If EGL_EXT_present_opaque is unavailable, ignore any config
			//       with an alpha channel to ensure the buffer is opaque
			if !mado.GlfwConfig.PlatformContext.EXT_present_opaque {
				if !fbconfig.Transparent && u.AlphaBits > 0 {
					continue
				}
			}
		}
		if val, ok := egl.EglGetConfigAttrib(disp, n, egl.EGL_SAMPLES); ok {
			u.Samples = int(val)
		}

		u.Doublebuffer = fbconfig.Doublebuffer

		u.Handle = uintptr(n)
		usableConfigs = append(usableConfigs, u)
	}

	closest := mado.ChooseFBConfig(fbconfig, usableConfigs)
	if closest == nil {
		if wrongApiAvailable {
			if ctxconfig.Client == mado.GLFW_OPENGL_ES_API {
				if ctxconfig.Major == 1 {
					return egl.EGLConfig(0), fmt.Errorf("EGL: Failed to find support for OpenGL ES 1.x")
				} else {
					return egl.EGLConfig(0), fmt.Errorf("EGL: Failed to find support for OpenGL ES 2 or later")
				}
			} else {
				return egl.EGLConfig(0), fmt.Errorf("EGL: Failed to find support for OpenGL")
			}
		} else {
			return egl.EGLConfig(0), fmt.Errorf("EGL: Failed to find a suitable EGLConfig")
		}
	}

	return egl.EGLConfig(closest.Handle), nil
}

func createSurface(disp egl.EGLDisplay, eglCtx *eglContext, attribs []egl.EGLint, win egl.NativeWindowType) (egl.EGLSurface, error) {
	var surfAttribs []egl.EGLint
	if attribs == nil {
		if eglCtx.srgb {
			surfAttribs = append(surfAttribs, egl.EGL_GL_COLORSPACE_KHR, egl.EGL_GL_COLORSPACE_SRGB_KHR)
		}
		surfAttribs = append(surfAttribs, egl.EGL_NONE)
	} else {
		surfAttribs = attribs
	}
	eglSurf := egl.EglCreateWindowSurface(disp, eglCtx.config, win, surfAttribs)
	if eglSurf == egl.NilEGLSurface && eglCtx.srgb {
		// Try again without sRGB.
		eglCtx.srgb = false
		surfAttribs = []egl.EGLint{egl.EGL_NONE}
		eglSurf = egl.EglCreateWindowSurface(disp, eglCtx.config, win, surfAttribs)
	}
	if eglSurf == egl.NilEGLSurface {
		return egl.NilEGLSurface, fmt.Errorf("newContext: eglCreateWindowSurface failed 0x%x (sRGB=%v)", egl.EglGetError(), eglCtx.srgb)
	}
	return eglSurf, nil
}
