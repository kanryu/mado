// SPDX-License-Identifier: Unlicense OR MIT

//go:build darwin && !ios && nometal
// +build darwin,!ios,nometal

package app

import (
	"errors"
	"runtime"

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

var _ mado.Context = (*glContext)(nil)

type glContext struct {
	c    *gl.Functions
	ctx  C.CFTypeRef
	view C.CFTypeRef

	glFlush C.PFN_glFlush
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
	fbconfig := FbConfig{
		redBits:      8,
		greenBits:    8,
		blueBits:     8,
		alphaBits:    8,
		depthBits:    24,
		stencilBits:  8,
		doublebuffer: true,
	}
	attribs := makeGlV2Attributes(fbconfig)
	return C.gio_createGLContext2((*C.uint)(unsafe.Pointer(&attribs[0])))
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

func (c *glContext) SwapBuffers() {
	c.Lock()
	defer c.Unlock()
	C.gio_swapBuffers(c.ctx)
}

func (w *window) NewContext() (mado.Context, error) {
	return newContext(w)
}

type FbConfig struct {
	major          int
	minor          int
	offline        bool
	redBits        int
	greenBits      int
	blueBits       int
	alphaBits      int
	depthBits      int
	stencilBits    int
	accumRedBits   int
	accumGreenBits int
	accumBlueBits  int
	accumAlphaBits int
	auxBuffers     int
	stereo         bool
	samples        int
	sRGB           bool
	doublebuffer   bool
	transparent    bool
}
type NSOpenGLPixelFormatAttribute uint32

const (
	GLFW_DONT_CARE                            = 0
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

func makeGlV2Attributes(fbconfig FbConfig) []NSOpenGLPixelFormatAttribute {
	attribs := []NSOpenGLPixelFormatAttribute{
		NSOpenGLPFAAccelerated,
		NSOpenGLPFAClosestPolicy,
	}

	if fbconfig.offline {
		attribs = append(attribs, NSOpenGLPFAAllowOfflineRenderers)
		// NOTE: This replaces the NSSupportsAutomaticGraphicsSwitching key in
		//       Info.plist for unbundled applications
		// HACK: This assumes that NSOpenGLPixelFormat will remain
		//       a straightforward wrapper of its CGL counterpart
		attribs = append(attribs, kCGLPFASupportsAutomaticGraphicsSwitching)
	}

	if fbconfig.major >= 4 {
		attribs = append(attribs, NSOpenGLPFAOpenGLProfile, NSOpenGLProfileVersion4_1Core)
	} else if fbconfig.major >= 3 {
		attribs = append(attribs, NSOpenGLPFAOpenGLProfile, NSOpenGLProfileVersion3_2Core)
	}
	if fbconfig.major <= 2 {
		if fbconfig.auxBuffers != GLFW_DONT_CARE {
			attribs = append(attribs, NSOpenGLPFAAuxBuffers, NSOpenGLPixelFormatAttribute(fbconfig.auxBuffers))
		}

		if fbconfig.accumRedBits != GLFW_DONT_CARE &&
			fbconfig.accumGreenBits != GLFW_DONT_CARE &&
			fbconfig.accumBlueBits != GLFW_DONT_CARE &&
			fbconfig.accumAlphaBits != GLFW_DONT_CARE {
			accumBits := fbconfig.accumRedBits +
				fbconfig.accumGreenBits +
				fbconfig.accumBlueBits +
				fbconfig.accumAlphaBits

			attribs = append(attribs, NSOpenGLPFAAccumSize, NSOpenGLPixelFormatAttribute(accumBits))
		}
	}

	if fbconfig.redBits != GLFW_DONT_CARE &&
		fbconfig.greenBits != GLFW_DONT_CARE &&
		fbconfig.blueBits != GLFW_DONT_CARE {
		colorBits := fbconfig.redBits +
			fbconfig.greenBits +
			fbconfig.blueBits

		// macOS needs non-zero color size, so set reasonable values
		if colorBits == 0 {
			colorBits = 24

		} else if colorBits < 15 {
			colorBits = 15
		}

		attribs = append(attribs, NSOpenGLPFAColorSize, NSOpenGLPixelFormatAttribute(colorBits))
	}

	if fbconfig.alphaBits != GLFW_DONT_CARE {
		attribs = append(attribs, NSOpenGLPFAAlphaSize, NSOpenGLPixelFormatAttribute(fbconfig.alphaBits))
	}
	if fbconfig.depthBits != GLFW_DONT_CARE {
		attribs = append(attribs, NSOpenGLPFADepthSize, NSOpenGLPixelFormatAttribute(fbconfig.depthBits))
	}
	if fbconfig.stencilBits != GLFW_DONT_CARE {
		attribs = append(attribs, NSOpenGLPFAStencilSize, NSOpenGLPixelFormatAttribute(fbconfig.stencilBits))
	}
	if fbconfig.doublebuffer {
		attribs = append(attribs, NSOpenGLPFADoubleBuffer)
	}

	if fbconfig.samples != GLFW_DONT_CARE {
		if fbconfig.samples == 0 {
			attribs = append(attribs, NSOpenGLPFASampleBuffers, 0)
		} else {
			attribs = append(attribs, NSOpenGLPFASampleBuffers, 1)
			attribs = append(attribs, NSOpenGLPFASamples, NSOpenGLPixelFormatAttribute(fbconfig.samples))
		}
	}

	// NOTE: All NSOpenGLPixelFormats on the relevant cards support sRGB
	//       framebuffer, so there's no need (and no way) to request it

	attribs = append(attribs, 0)
	return attribs
}
