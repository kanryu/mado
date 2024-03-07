//go:build darwin
// +build darwin

package gl

import (
	"fmt"
	"runtime"
	"strings"
	"unsafe"
)

/*
#cgo CFLAGS: -Werror
#include <stdint.h>
#include <stdlib.h>
#include <sys/types.h>
#define __USE_GNU
#include <dlfcn.h>

typedef unsigned int GLenum;
typedef unsigned int GLuint;
typedef char GLchar;
typedef float GLfloat;
typedef ssize_t GLsizeiptr;
typedef intptr_t GLintptr;
typedef unsigned int GLbitfield;
typedef int GLint;
typedef unsigned char GLboolean;
typedef int GLsizei;
typedef uint8_t GLubyte;

typedef const GLubyte *(*_glGetString)(GLenum name);
typedef const GLubyte* (*_glGetStringi)(GLenum name, GLuint index);
typedef void (*_glGetIntegerv)(GLenum pname, GLint *data);
typedef void (*_glGetIntegeri_v)(GLenum pname, GLuint idx, GLint *data);

static const GLubyte *glGetStringDarwin(_glGetString f, GLenum name) {
	return f(name);
}
static const GLubyte* glGetStringiDarwin(_glGetStringi f, GLenum name, GLuint index) {
	return f(name, index);
}
static void glGetIntegervDarwin(_glGetIntegerv f, GLenum pname, GLint *data) {
	f(pname, data);
}
static void glGetIntegeri_vDarwin(_glGetIntegeri_v f, GLenum pname, GLuint idx, GLint *data) {
	f(pname, idx, data);
}
__attribute__ ((visibility ("hidden"))) void * gio_getProcAddress(const char* procname);
*/
import "C"

type FunctionsDarwin struct {
	// Query caches.
	uints  [100]C.GLuint
	ints   [100]C.GLint
	floats [100]C.GLfloat

	glGetString     C._glGetString
	glGetStringi    C._glGetStringi
	glGetIntegerv   C._glGetIntegerv
	glGetIntegeri_v C._glGetIntegeri_v
}

func NewFunctionsDarwin(ctx Context, forceES bool) (*FunctionsDarwin, error) {
	if ctx != nil {
		panic("non-nil context")
	}
	f := new(FunctionsDarwin)
	err := f.load(forceES)
	if err != nil {
		return nil, err
	}
	return f, nil
}

func getProcAddress(lib string) unsafe.Pointer {
	clib := C.CString(lib)
	defer C.free(unsafe.Pointer(clib))
	return C.gio_getProcAddress(clib)
}

func (f *FunctionsDarwin) load(forceES bool) error {
	var (
		loadErr error
	)

	load := func(s string) *[0]byte {
		if f := getProcAddress(s); f != nil {
			return (*[0]byte)(f)
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
	f.glGetIntegerv = must("glGetIntegerv")
	f.glGetString = must("glGetString")
	f.glGetIntegeri_v = load("glGetIntegeri_v")
	f.glGetStringi = load("glGetStringi")

	return loadErr
}

func (f *FunctionsDarwin) GetInteger4(pname Enum) [4]int {
	C.glGetIntegervDarwin(f.glGetIntegerv, C.GLenum(pname), &f.ints[0])
	var r [4]int
	for i := range r {
		r[i] = int(f.ints[i])
	}
	return r
}

func (f *FunctionsDarwin) GetInteger(pname Enum) int {
	C.glGetIntegervDarwin(f.glGetIntegerv, C.GLenum(pname), &f.ints[0])
	return int(f.ints[0])
}

func (f *FunctionsDarwin) GetIntegeri(pname Enum, idx int) int {
	C.glGetIntegeri_vDarwin(f.glGetIntegeri_v, C.GLenum(pname), C.GLuint(idx), &f.ints[0])
	return int(f.ints[0])
}

func (f *FunctionsDarwin) getStringi(pname Enum, index int) string {
	str := C.glGetStringiDarwin(f.glGetStringi, C.GLenum(pname), C.GLuint(index))
	if str == nil {
		return ""
	}
	return C.GoString((*C.char)(unsafe.Pointer(str)))
}

func (f *FunctionsDarwin) GetString(pname Enum) string {
	switch {
	case runtime.GOOS == "darwin" && pname == EXTENSIONS:
		// macOS OpenGL 3 core profile doesn't support glGetString(GL_EXTENSIONS).
		// Use glGetStringi(GL_EXTENSIONS, <index>).
		var exts []string
		nexts := f.GetInteger(NUM_EXTENSIONS)
		for i := 0; i < nexts; i++ {
			ext := f.getStringi(EXTENSIONS, i)
			exts = append(exts, ext)
		}
		return strings.Join(exts, " ")
	default:
		str := C.glGetStringDarwin(f.glGetString, C.GLenum(pname))
		return C.GoString((*C.char)(unsafe.Pointer(str)))
	}
}
