// SPDX-License-Identifier: Unlicense OR MIT

//go:build darwin && !ios

package app

import (
	"errors"
	"fmt"
	"runtime"
	"slices"
	"strings"

	"unsafe"

	"github.com/kanryu/mado"
	"github.com/kanryu/mado/gpu"
	"github.com/kanryu/mado/internal/gl"
)

/*
#cgo CFLAGS: -DGL_SILENCE_DEPRECATION -xobjective-c -fobjc-arc
#cgo LDFLAGS: -framework OpenGL

#include <CoreFoundation/CoreFoundation.h>
#include <CoreGraphics/CoreGraphics.h>
#include <AppKit/AppKit.h>
#include <dlfcn.h>

__attribute__ ((visibility ("hidden"))) void gio_swapInterval(CFTypeRef ctxRef, int interval);
__attribute__ ((visibility ("hidden"))) void gio_swapBuffers(CFTypeRef object);
__attribute__ ((visibility ("hidden"))) CFTypeRef gio_createGLContext(void);
__attribute__ ((visibility ("hidden"))) CFTypeRef gio_createGLContext2(NSOpenGLPixelFormatAttribute* attribs);
__attribute__ ((visibility ("hidden"))) void gio_setContextView(CFTypeRef ctx, CFTypeRef view);
__attribute__ ((visibility ("hidden"))) void gio_makeCurrentContext(CFTypeRef ctx);
__attribute__ ((visibility ("hidden"))) void gio_updateContext(CFTypeRef ctx);
__attribute__ ((visibility ("hidden"))) void gio_flushContextBuffer(CFTypeRef ctx);
__attribute__ ((visibility ("hidden"))) void gio_clearCurrentContext(void);
__attribute__ ((visibility ("hidden"))) void gio_lockContext(CFTypeRef ctxRef);
__attribute__ ((visibility ("hidden"))) void gio_unlockContext(CFTypeRef ctxRef);

typedef void (*PFN_glFlush)(void);

static void glFlush(PFN_glFlush f) {
	f();
}
*/
import "C"

type NSOpenGLPixelFormatAttribute uint32

const (
	NSOpenGLPFADoubleBuffer                   = 5
	NSOpenGLPFAAuxBuffers                     = 7
	NSOpenGLPFAColorSize                      = 8
	NSOpenGLPFAAlphaSize                      = 11
	NSOpenGLPFADepthSize                      = 12
	NSOpenGLPFAStencilSize                    = 13
	NSOpenGLPFAAccumSize                      = 14
	NSOpenGLPFASampleBuffers                  = 55
	NSOpenGLPFASamples                        = 56
	NSOpenGLPFAAccelerated                    = 73
	NSOpenGLPFAClosestPolicy                  = 74
	NSOpenGLPFAAllowOfflineRenderers          = 96
	NSOpenGLPFAOpenGLProfile                  = 99
	kCGLPFASupportsAutomaticGraphicsSwitching = 101
	NSOpenGLProfileVersion4_1Core             = 0x4100
	NSOpenGLProfileVersion3_2Core             = 0x3200
)

type PlatformContextState struct {
}

type PlatformLibraryContextState struct {
	inited bool
}

var _ mado.Context = (*glContext)(nil)

type glContext struct {
	c       *gl.Functions
	ctx     C.CFTypeRef
	view    C.CFTypeRef
	context GlfwContext

	glFlush C.PFN_glFlush
}

func init() {
	glfwconfiginit()
}

func newContext(w *window) (*glContext, error) {
	clib := C.CString("/System/Library/Frameworks/OpenGL.framework/OpenGL")
	defer C.free(unsafe.Pointer(clib))
	lib, err := C.dlopen(clib, C.RTLD_NOW|C.RTLD_LOCAL)
	if err != nil {
		return nil, err
	}
	csym := C.CString("glFlush")
	defer C.free(unsafe.Pointer(csym))
	glFlush := C.PFN_glFlush(C.dlsym(lib, csym))
	if glFlush == nil {
		return nil, errors.New("gl: missing symbol glFlush in the OpenGL framework")
	}
	view := w.contextView()
	//ctx := C.gio_createGLContext()
	ctx := newGlContext()
	if ctx == 0 {
		return nil, errors.New("gl: failed to create NSOpenGLContext")
	}
	C.gio_setContextView(ctx, view)
	c := &glContext{
		ctx:     ctx,
		view:    view,
		glFlush: glFlush,
	}
	c.Lock()
	// defer c.Unlock()
	//c.MakeCurrentContext()
	err = c.RefreshContextAttribs(&GlfwConfig.Hints.Context)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func newGlContext() C.CFTypeRef {
	if !IsEnablePollEvents() {
		return C.gio_createGLContext()
	}
	// in void glfwDefaultWindowHints(void)
	// // The default is 24 bits of color, 24 bits of depth and 8 bits of stencil,
	// // double buffered
	// memset(&_glfw.hints.framebuffer, 0, sizeof(_glfw.hints.framebuffer));
	// _glfw.hints.framebuffer.redBits      = 8;
	// _glfw.hints.framebuffer.greenBits    = 8;
	// _glfw.hints.framebuffer.blueBits     = 8;
	// _glfw.hints.framebuffer.alphaBits    = 8;
	// _glfw.hints.framebuffer.depthBits    = 24;
	// _glfw.hints.framebuffer.stencilBits  = 8;
	// _glfw.hints.framebuffer.doublebuffer = GLFW_TRUE;
	attribs := makeGlV2Attributes(&GlfwConfig.Hints.Context, &GlfwConfig.Hints.Framebuffer)
	return C.gio_createGLContext2((*C.uint)(unsafe.Pointer(&attribs[0])))
}

func (c *glContext) RefreshContextAttribs(ctxconfig *CtxConfig) error {
	f, err := gl.NewFunctions(nil, false)
	if err != nil {
		return err
	}
	exts := strings.Split(f.GetString(gl.EXTENSIONS), " ")
	fmt.Println("GetString(GL_EXTENSIONS)", exts)

	glVer := f.GetString(gl.VERSION)
	fmt.Println("GetString(GL_VERSION)", glVer)

	prefixes := []string{
		"OpenGL ES-CM ",
		"OpenGL ES-CL ",
		"OpenGL ES ",
	}
	for _, pref := range prefixes {
		if glVer[:len(pref)] == pref {
			c.context.Client = GLFW_OPENGL_ES_API
		}
	}
	ver, _, err := gl.ParseGLVersion(glVer)
	if err != nil {
		return err
	}
	c.context.Major, c.context.Minor = ver[0], ver[1]
	if c.context.Major < ctxconfig.Major ||
		(c.context.Major == ctxconfig.Major &&
			c.context.Minor < ctxconfig.Minor) {
		// The desired OpenGL version is greater than the actual version
		// This only happens if the machine lacks {GLX|WGL}_ARB_create_context
		// /and/ the user has requested an OpenGL version greater than 1.0

		// For API consistency, we emulate the behavior of the
		// {GLX|WGL}_ARB_create_context extension and fail here

		if c.context.Client == GLFW_OPENGL_API {
			return fmt.Errorf("Requested OpenGL version %i.%i, got version %i.%i",
				ctxconfig.Major, ctxconfig.Minor,
				c.context.Major, c.context.Minor,
			)
		} else {
			return fmt.Errorf("Requested OpenGL ES version %i.%i, got version %i.%i",
				ctxconfig.Major, ctxconfig.Minor,
				c.context.Major, c.context.Minor,
			)
		}
	}
	// Read back context flags (OpenGL 3.0 and above)
	if c.context.Major >= 3 {
		flags := f.GetInteger(gl.CONTEXT_FLAGS)
		if flags&gl.CONTEXT_FLAG_FORWARD_COMPATIBLE_BIT != 0 {
			c.context.Forward = true
		}
		if flags&gl.CONTEXT_FLAG_DEBUG_BIT != 0 {
			c.context.Debug = true
		} else if slices.Contains(exts, "GL_ARB_debug_output") && ctxconfig.Debug {
			// HACK: This is a workaround for older drivers (pre KHR_debug)
			//       not setting the debug bit in the context flags for
			//       debug contexts
			c.context.Debug = true
		}
		if flags&gl.CONTEXT_FLAG_NO_ERROR_BIT_KHR != 0 {
			c.context.Noerror = true
		}
	}
	// Read back OpenGL context profile (OpenGL 3.2 and above)
	if c.context.Major >= 4 ||
		(c.context.Major == 3 && c.context.Minor >= 2) {
		mask := f.GetInteger(gl.CONTEXT_PROFILE_MASK)
		if mask&gl.CONTEXT_COMPATIBILITY_PROFILE_BIT != 0 {
			c.context.Profile = GLFW_OPENGL_COMPAT_PROFILE
		} else if mask&gl.CONTEXT_CORE_PROFILE_BIT != 0 {
			c.context.Profile = GLFW_OPENGL_CORE_PROFILE
		} else if slices.Contains(exts, "GL_ARB_compatibility") {
			// HACK: This is a workaround for the compatibility profile bit
			//       not being set in the context flags if an OpenGL 3.2+
			//       context was created without having requested a specific
			//       version
			c.context.Profile = GLFW_OPENGL_COMPAT_PROFILE
		}
	}

	// Read back robustness strategy
	if slices.Contains(exts, "GL_ARB_robustness") {
		// NOTE: We avoid using the context flags for detection, as they are
		//       only present from 3.0 while the extension applies from 1.1
		strategy := f.GetInteger(gl.RESET_NOTIFICATION_STRATEGY_ARB)
		if strategy == gl.LOSE_CONTEXT_ON_RESET_ARB {
			c.context.Robustness = GLFW_LOSE_CONTEXT_ON_RESET
		} else if strategy == gl.NO_RESET_NOTIFICATION_ARB {
			c.context.Robustness = GLFW_NO_RESET_NOTIFICATION
		}
	}
	if slices.Contains(exts, "GL_KHR_context_flush_control") {
		behavior := f.GetInteger(gl.CONTEXT_RELEASE_BEHAVIOR)
		if behavior == gl.ZERO {
			c.context.Release = GLFW_RELEASE_BEHAVIOR_NONE
		} else if behavior == gl.CONTEXT_RELEASE_BEHAVIOR_FLUSH {
			c.context.Release = GLFW_RELEASE_BEHAVIOR_FLUSH
		}
	}
	return nil
}

func (c *glContext) RenderTarget() (gpu.RenderTarget, error) {
	return gpu.OpenGLRenderTarget{}, nil
}

func (c *glContext) API() gpu.API {
	return gpu.OpenGL{}
}

func (c *glContext) Release() {
	if c.ctx != 0 {
		C.gio_clearCurrentContext()
		C.CFRelease(c.ctx)
		c.ctx = 0
	}
}

func (c *glContext) Present() error {
	// Assume the caller already locked the context.
	C.glFlush(c.glFlush)
	return nil
}

func (c *glContext) Lock() error {
	// OpenGL contexts are implicit and thread-local. Lock the OS thread.
	runtime.LockOSThread()

	C.gio_lockContext(c.ctx)
	C.gio_makeCurrentContext(c.ctx)
	return nil
}

func (c *glContext) MakeCurrentContext() error {
	// OpenGL contexts are implicit and thread-local. Lock the OS thread.
	runtime.LockOSThread()

	C.gio_makeCurrentContext(c.ctx)
	return nil
}

func (c *glContext) Unlock() {
	C.gio_clearCurrentContext()
	C.gio_unlockContext(c.ctx)
}

func (c *glContext) Refresh() error {
	c.Lock()
	defer c.Unlock()
	C.gio_updateContext(c.ctx)
	return nil
}

func (c *glContext) SwapBuffers() error {
	//c.Lock()
	//defer c.Unlock()
	C.gio_swapBuffers(c.ctx)
	return nil
}

func (c *glContext) SwapInterval(interval int) {
	//c.Lock()
	//defer c.Unlock()
	C.gio_swapInterval(c.ctx, C.int(interval))
}

func (w *window) NewContext() (mado.Context, error) {
	return newContext(w)
}

func makeGlV2Attributes(ctxconfig *CtxConfig, fbconfig *FbConfig) []NSOpenGLPixelFormatAttribute {
	attribs := []NSOpenGLPixelFormatAttribute{
		NSOpenGLPFAAccelerated,
		NSOpenGLPFAClosestPolicy,
	}

	if ctxconfig.Nsgl.Offline {
		attribs = append(attribs, NSOpenGLPFAAllowOfflineRenderers)
		// NOTE: This replaces the NSSupportsAutomaticGraphicsSwitching key in
		//       Info.plist for unbundled applications
		// HACK: This assumes that NSOpenGLPixelFormat will remain
		//       a straightforward wrapper of its CGL counterpart
		attribs = append(attribs, kCGLPFASupportsAutomaticGraphicsSwitching)
	}

	if ctxconfig.Major >= 4 {
		attribs = append(attribs, NSOpenGLPFAOpenGLProfile, NSOpenGLProfileVersion4_1Core)
	} else if ctxconfig.Major >= 3 {
		attribs = append(attribs, NSOpenGLPFAOpenGLProfile, NSOpenGLProfileVersion3_2Core)
	}
	if ctxconfig.Major <= 2 {
		if fbconfig.AuxBuffers != GLFW_DONT_CARE {
			attribs = append(attribs, NSOpenGLPFAAuxBuffers, NSOpenGLPixelFormatAttribute(fbconfig.AuxBuffers))
		}

		if fbconfig.AccumRedBits != GLFW_DONT_CARE &&
			fbconfig.AccumGreenBits != GLFW_DONT_CARE &&
			fbconfig.AccumBlueBits != GLFW_DONT_CARE &&
			fbconfig.AccumAlphaBits != GLFW_DONT_CARE {
			accumBits := fbconfig.AccumRedBits +
				fbconfig.AccumGreenBits +
				fbconfig.AccumBlueBits +
				fbconfig.AccumAlphaBits

			attribs = append(attribs, NSOpenGLPFAAccumSize, NSOpenGLPixelFormatAttribute(accumBits))
		}
	}

	if fbconfig.RedBits != GLFW_DONT_CARE &&
		fbconfig.GreenBits != GLFW_DONT_CARE &&
		fbconfig.BlueBits != GLFW_DONT_CARE {
		colorBits := fbconfig.RedBits +
			fbconfig.GreenBits +
			fbconfig.BlueBits

		// macOS needs non-zero color size, so set reasonable values
		if colorBits == 0 {
			colorBits = 24

		} else if colorBits < 15 {
			colorBits = 15
		}

		attribs = append(attribs, NSOpenGLPFAColorSize, NSOpenGLPixelFormatAttribute(colorBits))
	}

	if fbconfig.AlphaBits != GLFW_DONT_CARE {
		attribs = append(attribs, NSOpenGLPFAAlphaSize, NSOpenGLPixelFormatAttribute(fbconfig.AlphaBits))
	}
	if fbconfig.DepthBits != GLFW_DONT_CARE {
		attribs = append(attribs, NSOpenGLPFADepthSize, NSOpenGLPixelFormatAttribute(fbconfig.DepthBits))
	}
	if fbconfig.StencilBits != GLFW_DONT_CARE {
		attribs = append(attribs, NSOpenGLPFAStencilSize, NSOpenGLPixelFormatAttribute(fbconfig.StencilBits))
	}
	if fbconfig.Doublebuffer {
		attribs = append(attribs, NSOpenGLPFADoubleBuffer)
	}

	if fbconfig.Samples != GLFW_DONT_CARE {
		if fbconfig.Samples == 0 {
			attribs = append(attribs, NSOpenGLPFASampleBuffers, 0)
		} else {
			attribs = append(attribs, NSOpenGLPFASampleBuffers, 1)
			attribs = append(attribs, NSOpenGLPFASamples, NSOpenGLPixelFormatAttribute(fbconfig.Samples))
		}
	}

	// NOTE: All NSOpenGLPixelFormats on the relevant cards support sRGB
	//       framebuffer, so there's no need (and no way) to request it

	attribs = append(attribs, 0)
	return attribs
}
