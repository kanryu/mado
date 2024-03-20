// SPDX-License-Identifier: Unlicense OR MIT

//go:build linux || windows || freebsd || openbsd
// +build linux windows freebsd openbsd

package egl

var (
	NilEGLDisplay       EGLDisplay
	NilEGLSurface       EGLSurface
	NilEGLContext       EGLContext
	NilEGLConfig        EGLConfig
	EGL_DEFAULT_DISPLAY NativeDisplayType
)

const (
	EGL_CONFIG_CAVEAT = 0x3027

	EGL_SUCCESS                     = 0x3000
	EGL_NOT_INITIALIZED             = 0x3001
	EGL_BAD_ACCESS                  = 0x3002
	EGL_BAD_ALLOC                   = 0x3003
	EGL_BAD_ATTRIBUTE               = 0x3004
	EGL_BAD_CONFIG                  = 0x3005
	EGL_BAD_CONTEXT                 = 0x3006
	EGL_BAD_CURRENT_SURFACE         = 0x3007
	EGL_BAD_DISPLAY                 = 0x3008
	EGL_BAD_MATCH                   = 0x3009
	EGL_BAD_NATIVE_PIXMAP           = 0x300a
	EGL_BAD_NATIVE_WINDOW           = 0x300b
	EGL_BAD_PARAMETER               = 0x300c
	EGL_BAD_SURFACE                 = 0x300d
	EGL_CONTEXT_LOST                = 0x300e
	EGL_COLOR_BUFFER_TYPE           = 0x303f
	EGL_RGB_BUFFER                  = 0x308e
	EGL_SURFACE_TYPE                = 0x3033
	EGL_WINDOW_BIT                  = 0x0004
	EGL_RENDERABLE_TYPE             = 0x3040
	EGL_OPENGL_ES_BIT               = 0x0001
	EGL_OPENGL_ES2_BIT              = 0x0004
	EGL_OPENGL_BIT                  = 0x0008
	EGL_ALPHA_SIZE                  = 0x3021
	EGL_BLUE_SIZE                   = 0x3022
	EGL_GREEN_SIZE                  = 0x3023
	EGL_RED_SIZE                    = 0x3024
	EGL_DEPTH_SIZE                  = 0x3025
	EGL_STENCIL_SIZE                = 0x3026
	EGL_SAMPLES                     = 0x3031
	EGL_OPENGL_ES_API          uint = 0x30a0
	EGL_OPENGL_API             uint = 0x30a2
	EGL_NONE                        = 0x3038
	EGL_RENDER_BUFFER               = 0x3086
	EGL_SINGLE_BUFFER               = 0x3085
	EGL_EXTENSIONS                  = 0x3055
	EGL_CONTEXT_CLIENT_VERSION      = 0x3098
	EGL_NATIVE_VISUAL_ID            = 0x302e
	// EGL_NO_SURFACE ((EGLSurface) = 0)
	// EGL_NO_DISPLAY ((EGLDisplay) = 0)
	// EGL_NO_CONTEXT ((EGLContext) = 0)
	// EGL_DEFAULT_DISPLAY ((EGLNativeDisplayType) = 0)

	EGL_CONTEXT_OPENGL_FORWARD_COMPATIBLE_BIT_KHR      = 0x00000002
	EGL_CONTEXT_OPENGL_CORE_PROFILE_BIT_KHR            = 0x00000001
	EGL_CONTEXT_OPENGL_COMPATIBILITY_PROFILE_BIT_KHR   = 0x00000002
	EGL_CONTEXT_OPENGL_DEBUG_BIT_KHR                   = 0x00000001
	EGL_CONTEXT_OPENGL_RESET_NOTIFICATION_STRATEGY_KHR = 0x31bd
	EGL_NO_RESET_NOTIFICATION_KHR                      = 0x31be
	EGL_LOSE_CONTEXT_ON_RESET_KHR                      = 0x31bf
	EGL_CONTEXT_OPENGL_ROBUST_ACCESS_BIT_KHR           = 0x00000004
	EGL_CONTEXT_MAJOR_VERSION_KHR                      = 0x3098
	EGL_CONTEXT_MINOR_VERSION_KHR                      = 0x30fb
	EGL_CONTEXT_OPENGL_PROFILE_MASK_KHR                = 0x30fd
	EGL_CONTEXT_FLAGS_KHR                              = 0x30fc
	EGL_CONTEXT_OPENGL_NO_ERROR_KHR                    = 0x31b3
	EGL_GL_COLORSPACE_KHR                              = 0x309d
	EGL_GL_COLORSPACE_SRGB_KHR                         = 0x3089
	EGL_CONTEXT_RELEASE_BEHAVIOR_KHR                   = 0x2097
	EGL_CONTEXT_RELEASE_BEHAVIOR_NONE_KHR              = 0
	EGL_CONTEXT_RELEASE_BEHAVIOR_FLUSH_KHR             = 0x2098
	EGL_PRESENT_OPAQUE_EXT                             = 0x31df
)
