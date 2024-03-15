// SPDX-License-Identifier: Unlicense OR MIT

//go:build ((linux && !android) || freebsd || openbsd) && !noopengl
// +build linux,!android freebsd openbsd
// +build !noopengl

package app

import (
	"internal/egl"
)

type EGLint uint32
type EGLSurface uintptr
type EGLDisplay uintptr
type EGLContext uintptr
type EGLNativeDisplayType uintptr

const (
	EGL_SUCCESS                EGLint               = 0x3000
	EGL_NOT_INITIALIZED        EGLint               = 0x3001
	EGL_BAD_ACCESS             EGLint               = 0x3002
	EGL_BAD_ALLOC              EGLint               = 0x3003
	EGL_BAD_ATTRIBUTE          EGLint               = 0x3004
	EGL_BAD_CONFIG             EGLint               = 0x3005
	EGL_BAD_CONTEXT            EGLint               = 0x3006
	EGL_BAD_CURRENT_SURFACE    EGLint               = 0x3007
	EGL_BAD_DISPLAY            EGLint               = 0x3008
	EGL_BAD_MATCH              EGLint               = 0x3009
	EGL_BAD_NATIVE_PIXMAP      EGLint               = 0x300a
	EGL_BAD_NATIVE_WINDOW      EGLint               = 0x300b
	EGL_BAD_PARAMETER          EGLint               = 0x300c
	EGL_BAD_SURFACE            EGLint               = 0x300d
	EGL_CONTEXT_LOST           EGLint               = 0x300e
	EGL_COLOR_BUFFER_TYPE      EGLint               = 0x303f
	EGL_RGB_BUFFER             EGLint               = 0x308e
	EGL_SURFACE_TYPE           EGLint               = 0x3033
	EGL_WINDOW_BIT             EGLint               = 0x0004
	EGL_RENDERABLE_TYPE        EGLint               = 0x3040
	EGL_OPENGL_ES_BIT          EGLint               = 0x0001
	EGL_OPENGL_ES2_BIT         EGLint               = 0x0004
	EGL_OPENGL_BIT             EGLint               = 0x0008
	EGL_ALPHA_SIZE             EGLint               = 0x3021
	EGL_BLUE_SIZE              EGLint               = 0x3022
	EGL_GREEN_SIZE             EGLint               = 0x3023
	EGL_RED_SIZE               EGLint               = 0x3024
	EGL_DEPTH_SIZE             EGLint               = 0x3025
	EGL_STENCIL_SIZE           EGLint               = 0x3026
	EGL_SAMPLES                EGLint               = 0x3031
	EGL_OPENGL_ES_API          EGLint               = 0x30a0
	EGL_OPENGL_API             EGLint               = 0x30a2
	EGL_NONE                   EGLint               = 0x3038
	EGL_RENDER_BUFFER          EGLint               = 0x3086
	EGL_SINGLE_BUFFER          EGLint               = 0x3085
	EGL_EXTENSIONS             EGLint               = 0x3055
	EGL_CONTEXT_CLIENT_VERSION EGLint               = 0x3098
	EGL_NATIVE_VISUAL_ID       EGLint               = 0x302e
	EGL_NO_SURFACE             EGLSurface           = EGLSurface(0)
	EGL_NO_DISPLAY             EGLDisplay           = EGLDisplay(0)
	EGL_NO_CONTEXT             EGLContext           = EGLContext(0)
	EGL_DEFAULT_DISPLAY        EGLNativeDisplayType = EGLNativeDisplayType(0)

	EGL_CONTEXT_OPENGL_FORWARD_COMPATIBLE_BIT_KHR      EGLint = 0x00000002
	EGL_CONTEXT_OPENGL_CORE_PROFILE_BIT_KHR            EGLint = 0x00000001
	EGL_CONTEXT_OPENGL_COMPATIBILITY_PROFILE_BIT_KHR   EGLint = 0x00000002
	EGL_CONTEXT_OPENGL_DEBUG_BIT_KHR                   EGLint = 0x00000001
	EGL_CONTEXT_OPENGL_RESET_NOTIFICATION_STRATEGY_KHR EGLint = 0x31bd
	EGL_NO_RESET_NOTIFICATION_KHR                      EGLint = 0x31be
	EGL_LOSE_CONTEXT_ON_RESET_KHR                      EGLint = 0x31bf
	EGL_CONTEXT_OPENGL_ROBUST_ACCESS_BIT_KHR           EGLint = 0x00000004
	EGL_CONTEXT_MAJOR_VERSION_KHR                      EGLint = 0x3098
	EGL_CONTEXT_MINOR_VERSION_KHR                      EGLint = 0x30fb
	EGL_CONTEXT_OPENGL_PROFILE_MASK_KHR                EGLint = 0x30fd
	EGL_CONTEXT_FLAGS_KHR                              EGLint = 0x30fc
	EGL_CONTEXT_OPENGL_NO_ERROR_KHR                    EGLint = 0x31b3
	EGL_GL_COLORSPACE_KHR                              EGLint = 0x309d
	EGL_GL_COLORSPACE_SRGB_KHR                         EGLint = 0x3089
	EGL_CONTEXT_RELEASE_BEHAVIOR_KHR                   EGLint = 0x2097
	EGL_CONTEXT_RELEASE_BEHAVIOR_NONE_KHR              EGLint = 0
	EGL_CONTEXT_RELEASE_BEHAVIOR_FLUSH_KHR             EGLint = 0x2098
	EGL_PRESENT_OPAQUE_EXT                             EGLint = 0x31df
)

func (l EGLint) String() string {
	switch l {
	case EGL_SUCCESS:
		return "Success"
	case EGL_NOT_INITIALIZED:
		return "EGL is not or could not be initialized"
	case EGL_BAD_ACCESS:
		return "EGL cannot access a requested resource"
	case EGL_BAD_ALLOC:
		return "EGL failed to allocate resources for the requested operation"
	case EGL_BAD_ATTRIBUTE:
		return "An unrecognized attribute or attribute value was passed in the attribute list"
	case EGL_BAD_CONTEXT:
		return "An EGLContext argument does not name a valid EGL rendering context"
	case EGL_BAD_CONFIG:
		return "An EGLConfig argument does not name a valid EGL frame buffer configuration"
	case EGL_BAD_CURRENT_SURFACE:
		return "The current surface of the calling thread is a window, pixel buffer or pixmap that is no longer valid"
	case EGL_BAD_DISPLAY:
		return "An EGLDisplay argument does not name a valid EGL display connection"
	case EGL_BAD_SURFACE:
		return "An EGLSurface argument does not name a valid surface configured for GL rendering"
	case EGL_BAD_MATCH:
		return "Arguments are inconsistent"
	case EGL_BAD_PARAMETER:
		return "One or more argument values are invalid"
	case EGL_BAD_NATIVE_PIXMAP:
		return "A NativePixmapType argument does not refer to a valid native pixmap"
	case EGL_BAD_NATIVE_WINDOW:
		return "A NativeWindowType argument does not refer to a valid native window"
	case EGL_CONTEXT_LOST:
		return "The application must destroy all contexts and reinitialise"
	default:
		panic("ERROR: UNKNOWN EGL ERROR")
	}
}

func getEGLConfigAttrib(config EGLConfig, attrib int) int
{
    int value;
    egl.eglGetConfigAttrib(_glfw.egl.display, config, attrib, &value);
    return value;
}