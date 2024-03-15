// SPDX-License-Identifier: Unlicense OR MIT

//go:build ((linux && !android) || freebsd || openbsd) && !noopengl
// +build linux,!android freebsd openbsd
// +build !noopengl

package glx

import (
	"fmt"
	"runtime"
	"unsafe"
)

/*
#cgo CFLAGS: -Werror
#cgo linux freebsd LDFLAGS: -ldl

// #include <stdint.h>
// #include <stdlib.h>
// #include <sys/types.h>
// #define __USE_GNU
// #include <dlfcn.h>
//
// #ifndef APIENTRY
// #define APIENTRY
// #endif
// #ifndef APIENTRYP
// #define APIENTRYP APIENTRY *
// #endif
// #ifndef GLAPI
// #define GLAPI extern
// #endif
// 
// typedef unsigned int GLenum;
// typedef unsigned char GLboolean;
// typedef unsigned int GLbitfield;
// typedef signed char GLbyte;
// typedef short GLshort;
// typedef int GLint;
// typedef int GLsizei;
// typedef unsigned char GLubyte;
// typedef unsigned short GLushort;
// typedef unsigned int GLuint;
// typedef unsigned short GLhalf;
// typedef float GLfloat;
// typedef float GLclampf;
// typedef double GLdouble;
// typedef double GLclampd;
// typedef void GLvoid;
// 
// #ifndef GLX_ARB_get_proc_address
// typedef void (*__GLXextFuncPtr)(void);
// #endif
// 
// #ifndef GLX_SGIX_video_source
// typedef XID GLXVideoSourceSGIX;
// #endif
// 
// #ifndef GLX_SGIX_fbconfig
// typedef XID GLXFBConfigIDSGIX;
// typedef struct __GLXFBConfigRec *GLXFBConfigSGIX;
// #endif
// 
// #ifndef GLX_SGIX_pbuffer
// typedef XID GLXPbufferSGIX;
// typedef struct {
// int type;
// unsigned long serial;	  /* # of last request processed by server */
// Bool send_event;		  /* true if this came for SendEvent request */
// Display *display;		  /* display the event was read from */
// GLXDrawable drawable;	  /* i.d. of Drawable */
// int event_type;		  /* GLX_DAMAGED_SGIX or GLX_SAVED_SGIX */
// int draw_type;		  /* GLX_WINDOW_SGIX or GLX_PBUFFER_SGIX */
// unsigned int mask;	  /* mask indicating which buffers are affected*/
// int x, y;
// int width, height;
// int count;		  /* if nonzero, at least this many more */
// } GLXBufferClobberEventSGIX;
// #endif
// 
// #ifndef GLX_NV_video_output
// typedef unsigned int GLXVideoDeviceNV;
// #endif
// 
// #ifndef GLX_NV_video_capture
// typedef XID GLXVideoCaptureDeviceNV;
// #endif

// //  VERSION_1_1
// char * (* _glxGetClientString)(Display * dpy,int name);
// //  VERSION_1_3
// GLXContext (* _glxCreateNewContext)(Display* dpy, GLXFBConfig config, int render_type, GLXContext share_list, Bool direct);
// GLXWindow (* _glxCreateWindow)(Display* dpy, GLXFBConfig config, Window win, int* attrib_list);
// void (* _glxDestroyWindow)(Display* dpy, GLXWindow win);
// GLXFBConfig * (* _glxGetFBConfigs)(Display* dpy, int screen, int* nelements);
// XVisualInfo * (* _glxGetVisualFromFBConfig)(Display* dpy, GLXFBConfig config);
// int (* _glxGetFBConfigAttrib)(Display* dpy, GLXFBConfig config, int attribute, int* value);
// //  VERSION_1_4
// void (* _glxDestroyContext)(GLint context);
// void (* _glxMakeCurrent)(GLint drawable, GLint context);
// __GLXextFuncPtr (* _glxGetProcAddress)(GLubyte* procName);
// Bool (* _glxQueryExtension)(Display *dpy, int *errorBase, int *eventBase);
// void (* _glxQueryExtensionsString)(GLint screen);
// void (* _glxQueryVersion)(GLint* major, GLint* minor);
// //  ARB_create_context
// GLXContext (* _glxCreateContextAttribsARB)(Display* dpy, GLXFBConfig config, GLXContext share_context, Bool direct, int* attrib_list);
// //  ARB_get_proc_address
// __GLXextFuncPtr (* _glxGetProcAddressARB)(GLubyte* procName);
// //  EXT_swap_control
// void (* _glxSwapIntervalEXT)(Display* dpy, GLXDrawable drawable, int interval);
// //  OML_sync_control
// int64_t (* _glxSwapBuffersMscOML)(Display* dpy, GLXDrawable drawable, int64_t target_msc, int64_t divisor, int64_t remainder);
// //  SGI_swap_control
// int (* _glxSwapIntervalSGI)(int interval);


// //  VERSION_1_1
// static char * goglxGetClientString(_glxGetClientString f, Display * dpy,int name) {
//	return f(dpy, name);
// }
// //  VERSION_1_3
// static GLXContext goglxCreateNewContext(_glxCreateNewContext f, Display* dpy, GLXFBConfig config, int render_type, GLXContext share_list, Bool direct) {
// 	return f(dpy, config, render_type, share_list, direct);
// }
// static GLXWindow goglxCreateWindow(_glxCreateWindow f, Display* dpy, GLXFBConfig config, Window win, int* attrib_list) {
// 	return f(dpy, config, win, attrib_list);
// }
// static void goglxDestroyWindow(_glxDestroyWindow f, Display* dpy, GLXWindow win) {
// 	f(dpy, win);
// }
// static GLXFBConfig * goglxGetFBConfigs(_glxGetFBConfigs f, Display* dpy, int screen, int* nelements) {
// 	return f(dpy, screen, nelements);
// }
// static XVisualInfo * goglxGetVisualFromFBConfig(_glxGetVisualFromFBConfig f, Display* dpy, GLXFBConfig config) {
// 	return f(dpy, config);
// }
// static int goglxGetFBConfigAttrib(_glxGetFBConfigAttrib f, Display* dpy, GLXFBConfig config, int attribute, int* value) {
// 	return f(dpy, config, attribute, value);
// }
// //  VERSION_1_4
// static void goglxDestroyContext(_glxDestroyContext f, GLint context) {
// 	f(context);
// }
// static void goglxMakeCurrent(_glxMakeCurrent f, GLint drawable, GLint context) {
// 	f(drawable, context);
// }
// static __GLXextFuncPtr goglxGetProcAddress(_glxGetProcAddress f, GLubyte* procName) {
// 	return f(procName);
// }
// Bool glxQueryExtension)_glxQueryExtension f, (Display *dpy, int *errorBase, int *eventBase)
// 	return f(dpy, errBase, eventBase);
// }
// static void goglxQueryExtensionsString(_glxQueryExtensionsString f, GLint screen) {
// 	f(screen);
// }
// static void goglxQueryVersion(_glxQueryVersion f, GLint* major, GLint* minor) {
// 	f(major, minor);
// }
// //  ARB_create_context
// static GLXContext goglxCreateContextAttribsARB(_glxCreateContextAttribsARB f, Display* dpy, GLXFBConfig config, GLXContext share_context, Bool direct, int* attrib_list) {
// 	return f(dpy, config, share_context, direct, attrib_list);
// }
// //  ARB_get_proc_address
// static __GLXextFuncPtr goglxGetProcAddressARB(_glxGetProcAddressARB f, GLubyte* procName) {
// 	return f(procName);
// }
// //  EXT_swap_control
// static void goglxSwapIntervalEXT(_glxSwapIntervalEXT f, Display* dpy, GLXDrawable drawable, int interval) {
// 	f(dpy, drawable, interval);
// }
// //  OML_sync_control
// static int64_t goglxSwapBuffersMscOML(_glxSwapBuffersMscOML f, Display* dpy, GLXDrawable drawable, int64_t target_msc, int64_t divisor, int64_t remainder) {
// 	return f(dpy, drawable, target_msc, divisor, remainder);
// }
// //  SGI_swap_control
// static int goglxSwapIntervalSGI(_glxSwapIntervalSGI f, int interval) {
// 	return f(interval);
// }


*/
import "C"

type (
	Enum     C.GLenum
	Boolean  C.GLboolean
	Bitfield C.GLbitfield
	Byte     C.GLbyte
	Short    C.GLshort
	Int      C.GLint
	Sizei    C.GLsizei
	Ubyte    C.GLubyte
	Ushort   C.GLushort
	Uint     C.GLuint
	Half     C.GLhalf
	Float    C.GLfloat
	Clampf   C.GLclampf
	Double   C.GLdouble
	Clampd   C.GLclampd
	Char     C.GLchar
	Pointer  unsafe.Pointer
	Sync     C.GLsync
	Int64    C.GLint64
	Uint64   C.GLuint64
	Intptr   C.GLintptr
	Sizeiptr C.GLsizeiptr
)

type Context interface{}

type GlxFunctions struct {
	// Query caches.
	uints  [100]C.GLuint
	ints   [100]C.GLint
	floats [100]C.GLfloat

	glxGetFBConfigs          C._glxGetFBConfigs
	glxGetFBConfigAttrib     C._glxGetFBConfigAttrib
	glxGetClientString       C._glxGetClientString
	glxQueryExtension        C._glxQueryExtension
	glxQueryVersion          C._glxQueryVersion
	glxDestroyContext        C._glxDestroyContext
	glxMakeCurrent           C._glxMakeCurrent
	glxSwapBuffers           C._glxSwapBuffers
	glxQueryExtensionsString C._glxQueryExtensionsString
	glxCreateNewContext      C._glxCreateNewContext
	glxGetVisualFromFBConfig C._glxGetVisualFromFBConfig
	glxCreateWindow          C._glxCreateWindow
	glxDestroyWindow         C._glxDestroyWindow
	glxGetProcAddress        C._glxGetProcAddress
	glxGetProcAddressARB     C._glxGetProcAddressARB
	glxSwapIntervalSGI       C._glxSwapIntervalSGI
	glxSwapIntervalEXT       C._glxSwapIntervalEXT
	glxCreateContextAttribsARB C._glxCreateContextAttribsARB
	//glxSwapIntervalMESA        C.goglxSwapIntervalMESA
}

func NewGlxFunctions(ctx Context, forceES bool) (*GlxFunctions, error) {
	if ctx != nil {
		panic("non-nil context")
	}
	f := new(GlxFunctions)
	err := f.load(forceES)
	if err != nil {
		return nil, err
	}
	return f, nil
}

func dlsym(handle unsafe.Pointer, s string) unsafe.Pointer {
	cs := C.CString(s)
	defer C.free(unsafe.Pointer(cs))
	return C.dlsym(handle, cs)
}

func dlopen(lib string) unsafe.Pointer {
	clib := C.CString(lib)
	defer C.free(unsafe.Pointer(clib))
	return C.dlopen(clib, C.RTLD_NOW|C.RTLD_LOCAL)
}

func (f *GlxFunctions) load() error {
	var (
		loadErr  error
		libNames []string
		handles  []unsafe.Pointer
	)
	switch {
	case runtime.GOOS == "openbsd", runtime.GOOS == "netbsd":
		libNames = []string{"libGL.so"}
	default:
		libNames = []string{"libGLX.so.0", "libGL.so.1", "libGL.so"}
	}
	for _, lib := range libNames {
		if h := dlopen(lib); h != nil {
			handles = append(handles, h)
		}
	}
	if len(handles) == 0 {
		return fmt.Errorf("glx: no GLX implementation could be loaded (tried %q)", libNames)
	}
	load := func(s string) *[0]byte {
		for _, h := range handles {
			if f := dlsym(h, s); f != nil {
				return (*[0]byte)(f)
			}
		}
		return nil
	}
	must := func(s string) *[0]byte {
		ptr := load(s)
		if ptr == nil {
			loadErr = fmt.Errorf("gl: failed to load symbol %q", s)
		}
		return ptr
	}
	f.glxGetFBConfigs = must("glxGetFBConfigs")
	f.glxGetFBConfigAttrib = must("glxGetFBConfigAttrib")
	f.glxGetClientString = must("glxGetClientString")
	f.glxQueryExtension = must("glxQueryExtension")
	f.glxQueryVersion = must("glxQueryVersion")
	f.glxDestroyContext = must("glxDestroyContext")
	f.glxMakeCurrent = must("glxMakeCurrent")
	f.glxSwapBuffers = must("glxSwapBuffers")
	f.glxQueryExtensionsString = must("glxQueryExtensionsString")
	f.glxCreateNewContext = must("glxCreateNewContext")
	f.glxGetVisualFromFBConfig = must("glxGetVisualFromFBConfig")
	f.glxCreateWindow = must("glxCreateWindow")
	f.glxDestroyWindow = must("glxDestroyWindow")
	f.glxGetProcAddress = must("glxGetProcAddress")
	f.glxGetProcAddressARB = must("glxGetProcAddressARB")
	f.glxSwapIntervalSGI = must("glxSwapIntervalSGI")
	f.glxSwapIntervalEXT = must("glxSwapIntervalEXT")
	f.glxCreateContextAttribsARB = must("glxCreateContextAttribsARB")

	return loadErr
}

func (g *GlxFunctions) GetFBConfigs(dpy Pointer, screen Int, nelements *Int) Pointer {
	return (Pointer)(C.goglxGetFBConfigs(g.GetFBConfigs, (*C.Display)(dpy), (C.int)(screen), (*C.int)(nelements)))
}

func (g *GlxFunctions) GetFBConfigAttrib(dpy Pointer, config Pointer, attribute Int, value *Int) Int {
	return (Int)(C.goglxGetFBConfigAttrib(g.GetFBConfigAttrib, (*C.Display)(dpy), (C.GLXFBConfig)(config), (C.int)(attribute), (*C.int)(value)))
}

func (g *GlxFunctions) QueryVersion(major *Int, minor *Int)  {
	C.goglxQueryVersion(g.QueryVersion, (*C.GLint)(major), (*C.GLint)(minor))
}
func (g *GlxFunctions) DestroyContext(context Int)  {
	C.goglxDestroyContext(g.DestroyContext, (C.GLint)(context))
}
func (g *GlxFunctions) MakeCurrent(drawable Int, context Int)  {
	C.goglxMakeCurrent(g.MakeCurrent, (C.GLint)(drawable), (C.GLint)(context))
}
func (g *GlxFunctions) SwapBuffers(drawable Int)  {
	C.goglxSwapBuffers(g.SwapBuffers, (C.GLint)(drawable))
}
func (g *GlxFunctions) QueryExtensionsString(screen Int)  {
	C.goglxQueryExtensionsString(g.QueryExtensionsString, (C.GLint)(screen))
}
func (g *GlxFunctions) CreateNewContext(config Int, render_type Int, share_list Int, direct Int)  {
	C.goglxCreateNewContext(g.CreateNewContext, (C.GLint)(config), (C.GLint)(render_type), (C.GLint)(share_list), (C.GLint)(direct))
}
func (g *GlxFunctions) GetVisualFromFBConfig(dpy Pointer, config Pointer) Pointer {
	return (Pointer)(C.goglxGetVisualFromFBConfig(g.GetVisualFromFBConfig, (*C.Display)(dpy), (C.GLXFBConfig)(config)))
}
func (g *GlxFunctions) CreateWindow(dpy Pointer, config Pointer, win Pointer, attrib_list *Int) Pointer {
	return (Pointer)(C.goglxCreateWindow(g.CreateWindow, (*C.Display)(dpy), (C.GLXFBConfig)(config), (C.Window)(win), (*C.int)(attrib_list)))
}
func (g *GlxFunctions) DestroyWindow(dpy Pointer, win Pointer)  {
	C.goglxDestroyWindow(g.DestroyWindow, (*C.Display)(dpy), (C.GLXWindow)(win))
}
func (g *GlxFunctions) GetProcAddress(procName *Ubyte) Pointer {
	return (Pointer)(C.goglxGetProcAddress(g.GetProcAddress, (*C.GLubyte)(procName)))
}
// GetProcAddressARB ARB_get_proc_address
func (g *GlxFunctions) GetProcAddressARB(procName *Ubyte) Pointer {
	return (Pointer)(C.goglxGetProcAddressARB(g.GetProcAddressARB, (*C.GLubyte)(procName)))
}
// SwapIntervalSGI SGI_swap_control
func (g *GlxFunctions) SwapIntervalSGI(interval Int) Int {
	return (Int)(C.goglxSwapIntervalSGI(g.SwapIntervalSGI, (C.int)(interval)))
}
// SwapIntervalEXT EXT_swap_control
func (g *GlxFunctions) SwapIntervalEXT(dpy Pointer, drawable Pointer, interval Int)  {
	C.goglxSwapIntervalEXT(g.SwapIntervalEXT, (*C.Display)(dpy), (C.GLXDrawable)(drawable), (C.int)(interval))
}
// CreateContextAttribsARB ARB_create_context
func (g *GlxFunctions) CreateContextAttribsARB(dpy Pointer, config Pointer, share_context Pointer, direct int, attrib_list *Int) Pointer {
	return (Pointer)(C.goglxCreateContextAttribsARB(g.CreateContextAttribsARB, (*C.Display)(dpy), (C.GLXFBConfig)(config), (C.GLXContext)(share_context), (C.int)(direct), (*C.int)(attrib_list)))
}
