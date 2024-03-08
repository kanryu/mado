// SPDX-License-Identifier: Unlicense OR MIT

//go:build windows
// +build windows

package app

import (
	"fmt"
	"math"
	"runtime"
	"strings"

	"unsafe"

	"github.com/kanryu/mado"
	"github.com/kanryu/mado/app/internal/windows"
	"github.com/kanryu/mado/gpu"
	"github.com/kanryu/mado/internal/gl"
	syscall "golang.org/x/sys/windows"
)

type (
	Action          int
	ErrorCode       int
	Hint            int
	InputMode       int
	Key             int
	ModifierKey     int
	MouseButton     int
	PeripheralEvent int
	StandardCursor  int
)

const (
	GLFW_DONT_CARE              = 0
	GLFW_NO_API                 = 0
	GLFW_OPENGL_API             = 0x00030001
	GLFW_OPENGL_ES_API          = 0x00030002
	GLFW_OPENGL_ANY_PROFILE     = 0
	GLFW_OPENGL_CORE_PROFILE    = 0x00032001
	GLFW_OPENGL_COMPAT_PROFILE  = 0x00032002
	GLFW_NO_ROBUSTNESS          = 0
	GLFW_NO_RESET_NOTIFICATION  = 0x00031001
	GLFW_LOSE_CONTEXT_ON_RESET  = 0x00031002
	GLFW_ANY_RELEASE_BEHAVIOR   = 0
	GLFW_RELEASE_BEHAVIOR_FLUSH = 0x00035001
	GLFW_RELEASE_BEHAVIOR_NONE  = 0x00035002

	intSize = 32 << (^uint(0) >> 63)

	_BI_BITFIELDS                                                = 3
	_CCHDEVICENAME                                               = 32
	_CCHFORMNAME                                                 = 32
	_CDS_TEST                                                    = 0x00000002
	_CDS_FULLSCREEN                                              = 0x00000004
	_CS_HREDRAW                                                  = 0x00000002
	_CS_OWNDC                                                    = 0x00000020
	_CS_VREDRAW                                                  = 0x00000001
	_CW_USEDEFAULT                                               = int32(^0x7fffffff)
	_DBT_DEVTYP_DEVICEINTERFACE                                  = 0x00000005
	_DEVICE_NOTIFY_WINDOW_HANDLE                                 = 0x00000000
	_DIB_RGB_COLORS                                              = 0
	_DISP_CHANGE_SUCCESSFUL                                      = 0
	_DISP_CHANGE_RESTART                                         = 1
	_DISP_CHANGE_FAILED                                          = -1
	_DISP_CHANGE_BADMODE                                         = -2
	_DISP_CHANGE_NOTUPDATED                                      = -3
	_DISP_CHANGE_BADFLAGS                                        = -4
	_DISP_CHANGE_BADPARAM                                        = -5
	_DISP_CHANGE_BADDUALVIEW                                     = -6
	_DISPLAY_DEVICE_ACTIVE                                       = 0x00000001
	_DISPLAY_DEVICE_MODESPRUNED                                  = 0x08000000
	_DISPLAY_DEVICE_PRIMARY_DEVICE                               = 0x00000004
	_DM_BITSPERPEL                                               = 0x00040000
	_DM_PELSWIDTH                                                = 0x00080000
	_DM_PELSHEIGHT                                               = 0x00100000
	_DM_DISPLAYFREQUENCY                                         = 0x00400000
	_DWM_BB_BLURREGION                                           = 0x00000002
	_DWM_BB_ENABLE                                               = 0x00000001
	_EDS_ROTATEDMODE                                             = 0x00000004
	_ENUM_CURRENT_SETTINGS                        uint32         = 0xffffffff
	_GCLP_HICON                                                  = -14
	_GCLP_HICONSM                                                = -34
	_GET_MODULE_HANDLE_EX_FLAG_FROM_ADDRESS                      = 0x00000004
	_GET_MODULE_HANDLE_EX_FLAG_UNCHANGED_REFCOUNT                = 0x00000002
	_GWL_EXSTYLE                                                 = -20
	_GWL_STYLE                                                   = -16
	_HTCLIENT                                                    = 1
	_HORZSIZE                                                    = 4
	_HWND_NOTOPMOST                               syscall.Handle = (1 << intSize) - 2
	_HWND_TOP                                     syscall.Handle = 0
	_HWND_TOPMOST                                 syscall.Handle = (1 << intSize) - 1
	_ICON_BIG                                                    = 1
	_ICON_SMALL                                                  = 0
	_IDC_ARROW                                                   = 32512
	_IDI_APPLICATION                                             = 32512
	_IMAGE_CURSOR                                                = 2
	_IMAGE_ICON                                                  = 1
	_KF_ALTDOWN                                                  = 0x2000
	_KF_DLGMODE                                                  = 0x0800
	_KF_EXTENDED                                                 = 0x0100
	_KF_MENUMODE                                                 = 0x1000
	_KF_REPEAT                                                   = 0x4000
	_KF_UP                                                       = 0x8000
	_LOGPIXELSX                                                  = 88
	_LOGPIXELSY                                                  = 90
	_LR_DEFAULTSIZE                                              = 0x0040
	_LR_SHARED                                                   = 0x8000
	_LWA_ALPHA                                                   = 0x00000002
	_MAPVK_VK_TO_VSC                                             = 0
	_MAPVK_VSC_TO_VK                                             = 1
	_MONITOR_DEFAULTTONEAREST                                    = 0x00000002
	_MOUSE_MOVE_ABSOLUTE                                         = 0x01
	_MSGFLT_ALLOW                                                = 1
	_OCR_CROSS                                                   = 32515
	_OCR_HAND                                                    = 32649
	_OCR_IBEAM                                                   = 32513
	_OCR_NO                                                      = 32648
	_OCR_NORMAL                                                  = 32512
	_OCR_SIZEALL                                                 = 32646
	_OCR_SIZENESW                                                = 32643
	_OCR_SIZENS                                                  = 32645
	_OCR_SIZENWSE                                                = 32642
	_OCR_SIZEWE                                                  = 32644
	_PM_NOREMOVE                                                 = 0x0000
	_PM_REMOVE                                                   = 0x0001
	_PFD_DRAW_TO_WINDOW                                          = 0x00000004
	_PFD_DOUBLEBUFFER                                            = 0x00000001
	_PFD_GENERIC_ACCELERATED                                     = 0x00001000
	_PFD_GENERIC_FORMAT                                          = 0x00000040
	_PFD_STEREO                                                  = 0x00000002
	_PFD_SUPPORT_OPENGL                                          = 0x00000020
	_PFD_TYPE_RGBA                                               = 0
	_QS_ALLEVENTS                                                = _QS_INPUT | _QS_POSTMESSAGE | _QS_TIMER | _QS_PAINT | _QS_HOTKEY
	_QS_HOTKEY                                                   = 0x0080
	_QS_INPUT                                                    = _QS_MOUSE | _QS_KEY | _QS_RAWINPUT
	_QS_KEY                                                      = 0x0001
	_QS_MOUSE                                                    = _QS_MOUSEMOVE | _QS_MOUSEBUTTON
	_QS_MOUSEBUTTON                                              = 0x0004
	_QS_MOUSEMOVE                                                = 0x0002
	_QS_PAINT                                                    = 0x0020
	_QS_POSTMESSAGE                                              = 0x0008
	_QS_RAWINPUT                                                 = 0x0400
	_QS_TIMER                                                    = 0x0010
	_RID_INPUT                                                   = 0x10000003
	_RIDEV_REMOVE                                                = 0x00000001
	_SC_KEYMENU                                                  = 0xf100
	_SC_MONITORPOWER                                             = 0xf170
	_SC_SCREENSAVE                                               = 0xf140
	_SIZE_MAXIMIZED                                              = 2
	_SIZE_MINIMIZED                                              = 1
	_SIZE_RESTORED                                               = 0
	_SM_CXICON                                                   = 11
	_SM_CXSMICON                                                 = 49
	_SM_CYCAPTION                                                = 4
	_SM_CYICON                                                   = 12
	_SM_CYSMICON                                                 = 50
	_SPI_GETFOREGROUNDLOCKTIMEOUT                                = 0x2000
	_SPI_GETMOUSETRAILS                                          = 94
	_SPI_SETFOREGROUNDLOCKTIMEOUT                                = 0x2001
	_SPI_SETMOUSETRAILS                                          = 93
	_SPIF_SENDCHANGE                                             = _SPIF_SENDWININICHANGE
	_SPIF_SENDWININICHANGE                                       = 2
	_SW_HIDE                                                     = 0
	_SW_MAXIMIZE                                                 = _SW_SHOWMAXIMIZED
	_SW_MINIMIZE                                                 = 6
	_SW_RESTORE                                                  = 9
	_SW_SHOWNA                                                   = 8
	_SW_SHOWMAXIMIZED                                            = 3
	_SWP_FRAMECHANGED                                            = 0x0020
	_SWP_NOACTIVATE                                              = 0x0010
	_SWP_NOCOPYBITS                                              = 0x0100
	_SWP_NOMOVE                                                  = 0x0002
	_SWP_NOOWNERZORDER                                           = 0x0200
	_SWP_NOSIZE                                                  = 0x0001
	_SWP_NOZORDER                                                = 0x0004
	_SWP_SHOWWINDOW                                              = 0x0040
	_TLS_OUT_OF_INDEXES                           uint32         = 0xffffffff
	_TME_LEAVE                                                   = 0x00000002
	_UNICODE_NOCHAR                                              = 0xffff
	_USER_DEFAULT_SCREEN_DPI                                     = 96
	_VERTSIZE                                                    = 6
	_VK_ADD                                                      = 0x6B
	_VK_CAPITAL                                                  = 0x14
	_VK_CONTROL                                                  = 0x11
	_VK_DECIMAL                                                  = 0x6E
	_VK_DIVIDE                                                   = 0x6F
	_VK_LSHIFT                                                   = 0xA0
	_VK_LWIN                                                     = 0x5B
	_VK_MENU                                                     = 0x12
	_VK_MULTIPLY                                                 = 0x6A
	_VK_NUMLOCK                                                  = 0x90
	_VK_NUMPAD0                                                  = 0x60
	_VK_NUMPAD1                                                  = 0x61
	_VK_NUMPAD2                                                  = 0x62
	_VK_NUMPAD3                                                  = 0x63
	_VK_NUMPAD4                                                  = 0x64
	_VK_NUMPAD5                                                  = 0x65
	_VK_NUMPAD6                                                  = 0x66
	_VK_NUMPAD7                                                  = 0x67
	_VK_NUMPAD8                                                  = 0x68
	_VK_NUMPAD9                                                  = 0x69
	_VK_PROCESSKEY                                               = 0xE5
	_VK_RSHIFT                                                   = 0xA1
	_VK_RWIN                                                     = 0x5C
	_VK_SHIFT                                                    = 0x10
	_VK_SNAPSHOT                                                 = 0x2C
	_VK_SUBTRACT                                                 = 0x6D
	_WAIT_FAILED                                                 = 0xffffffff
	_WHEEL_DELTA                                                 = 120
	_WGL_ACCUM_BITS_ARB                                          = 0x201D
	_WGL_ACCELERATION_ARB                                        = 0x2003
	_WGL_ACCUM_ALPHA_BITS_ARB                                    = 0x2021
	_WGL_ACCUM_BLUE_BITS_ARB                                     = 0x2020
	_WGL_ACCUM_GREEN_BITS_ARB                                    = 0x201F
	_WGL_ACCUM_RED_BITS_ARB                                      = 0x201E
	_WGL_AUX_BUFFERS_ARB                                         = 0x2024
	_WGL_ALPHA_BITS_ARB                                          = 0x201B
	_WGL_ALPHA_SHIFT_ARB                                         = 0x201C
	_WGL_BLUE_BITS_ARB                                           = 0x2019
	_WGL_BLUE_SHIFT_ARB                                          = 0x201A
	_WGL_COLOR_BITS_ARB                                          = 0x2014
	_WGL_COLORSPACE_EXT                                          = 0x309D
	_WGL_COLORSPACE_SRGB_EXT                                     = 0x3089
	_WGL_CONTEXT_COMPATIBILITY_PROFILE_BIT_ARB                   = 0x00000002
	_WGL_CONTEXT_CORE_PROFILE_BIT_ARB                            = 0x00000001
	_WGL_CONTEXT_DEBUG_BIT_ARB                                   = 0x0001
	_WGL_CONTEXT_ES2_PROFILE_BIT_EXT                             = 0x00000004
	_WGL_CONTEXT_FLAGS_ARB                                       = 0x2094
	_WGL_CONTEXT_FORWARD_COMPATIBLE_BIT_ARB                      = 0x0002
	_WGL_CONTEXT_MAJOR_VERSION_ARB                               = 0x2091
	_WGL_CONTEXT_MINOR_VERSION_ARB                               = 0x2092
	_WGL_CONTEXT_OPENGL_NO_ERROR_ARB                             = 0x31B3
	_WGL_CONTEXT_PROFILE_MASK_ARB                                = 0x9126
	_WGL_CONTEXT_RELEASE_BEHAVIOR_ARB                            = 0x2097
	_WGL_CONTEXT_RELEASE_BEHAVIOR_NONE_ARB                       = 0x0000
	_WGL_CONTEXT_RELEASE_BEHAVIOR_FLUSH_ARB                      = 0x2098
	_WGL_CONTEXT_RESET_NOTIFICATION_STRATEGY_ARB                 = 0x8256
	_WGL_CONTEXT_ROBUST_ACCESS_BIT_ARB                           = 0x00000004
	_WGL_DEPTH_BITS_ARB                                          = 0x2022
	_WGL_DRAW_TO_BITMAP_ARB                                      = 0x2002
	_WGL_DRAW_TO_WINDOW_ARB                                      = 0x2001
	_WGL_DOUBLE_BUFFER_ARB                                       = 0x2011
	_WGL_FRAMEBUFFER_SRGB_CAPABLE_ARB                            = 0x20A9
	_WGL_GREEN_BITS_ARB                                          = 0x2017
	_WGL_GREEN_SHIFT_ARB                                         = 0x2018
	_WGL_LOSE_CONTEXT_ON_RESET_ARB                               = 0x8252
	_WGL_NEED_PALETTE_ARB                                        = 0x2004
	_WGL_NEED_SYSTEM_PALETTE_ARB                                 = 0x2005
	_WGL_NO_ACCELERATION_ARB                                     = 0x2025
	_WGL_NO_RESET_NOTIFICATION_ARB                               = 0x8261
	_WGL_NUMBER_OVERLAYS_ARB                                     = 0x2008
	_WGL_NUMBER_PIXEL_FORMATS_ARB                                = 0x2000
	_WGL_NUMBER_UNDERLAYS_ARB                                    = 0x2009
	_WGL_PIXEL_TYPE_ARB                                          = 0x2013
	_WGL_RED_BITS_ARB                                            = 0x2015
	_WGL_RED_SHIFT_ARB                                           = 0x2016
	_WGL_SAMPLES_ARB                                             = 0x2042
	_WGL_SHARE_ACCUM_ARB                                         = 0x200E
	_WGL_SHARE_DEPTH_ARB                                         = 0x200C
	_WGL_SHARE_STENCIL_ARB                                       = 0x200D
	_WGL_STENCIL_BITS_ARB                                        = 0x2023
	_WGL_STEREO_ARB                                              = 0x2012
	_WGL_SUPPORT_GDI_ARB                                         = 0x200F
	_WGL_SUPPORT_OPENGL_ARB                                      = 0x2010
	_WGL_SWAP_LAYER_BUFFERS_ARB                                  = 0x2006
	_WGL_SWAP_METHOD_ARB                                         = 0x2007
	_WGL_TRANSPARENT_ARB                                         = 0x200A
	_WGL_TRANSPARENT_ALPHA_VALUE_ARB                             = 0x203A
	_WGL_TRANSPARENT_BLUE_VALUE_ARB                              = 0x2039
	_WGL_TRANSPARENT_GREEN_VALUE_ARB                             = 0x2038
	_WGL_TRANSPARENT_INDEX_VALUE_ARB                             = 0x203B
	_WGL_TRANSPARENT_RED_VALUE_ARB                               = 0x2037
	_WGL_TYPE_RGBA_ARB                                           = 0x202B

	AnyReleaseBehavior   = 0
	CursorDisabled       = 0x00034003
	CursorHidden         = 0x00034002
	CursorNormal         = 0x00034001
	EGLContextAPI        = 0x00036002
	LoseContextOnReset   = 0x00031002
	NativeContextAPI     = 0x00036001
	NoAPI                = 0
	NoResetNotification  = 0x00031001
	NoRobustness         = 0
	OpenGLAPI            = 0x00030001
	OpenGLAnyProfile     = 0
	OpenGLCompatProfile  = 0x00032002
	OpenGLCoreProfile    = 0x00032001
	OpenGLESAPI          = 0x00030002
	OSMesaContextAPI     = 0x00036003
	ReleaseBehaviorFlush = 0x00035001
	ReleaseBehaviorNone  = 0x00035002
)

type PlatformContextState struct {
	dc       syscall.Handle
	handle   syscall.Handle
	interval int
}

type PlatformLibraryContextState struct {
	inited bool

	EXT_swap_control               bool
	EXT_colorspace                 bool
	ARB_multisample                bool
	ARB_framebuffer_sRGB           bool
	EXT_framebuffer_sRGB           bool
	ARB_pixel_format               bool
	ARB_create_context             bool
	ARB_create_context_profile     bool
	EXT_create_context_es2_profile bool
	ARB_create_context_robustness  bool
	ARB_create_context_no_error    bool
	ARB_context_flush_control      bool
}

const (
	NotInitialized     = ErrorCode(0x00010001)
	NoCurrentContext   = ErrorCode(0x00010002)
	InvalidEnum        = ErrorCode(0x00010003)
	InvalidValue       = ErrorCode(0x00010004)
	OutOfMemory        = ErrorCode(0x00010005)
	APIUnavailable     = ErrorCode(0x00010006)
	VersionUnavailable = ErrorCode(0x00010007)
	PlatformError      = ErrorCode(0x00010008)
	FormatUnavailable  = ErrorCode(0x00010009)
	NoWindowContext    = ErrorCode(0x0001000A)
)

var _ mado.Context = (*glContext)(nil)

type glContext struct {
	win   *window
	hglrc syscall.Handle
}

func init() {

	glfwconfiginit()
	drivers = append(drivers, gpuAPI{
		priority: 2,
		name:     "opengl",
		initializer: func(w *window) (mado.Context, error) {
			ctx := &glContext{win: w}
			ctx.MakeCurrentContext()
			return ctx, nil
		},
	})
}

func (c *glContext) Release() {
}

func (c *glContext) Refresh() error {
	return nil
}

func (c *glContext) Lock() error {
	return nil
}

func (c *glContext) Unlock() {
}

func (c *glContext) MakeCurrentContext() error {
	// OpenGL contexts are implicit and thread-local. Lock the OS thread.
	runtime.LockOSThread()

	if err := c.win.initWGL(); err != nil {
		panic(err)
	}

	return nil
}

func (c *glContext) SwapBuffers() {
	windows.SwapBuffers(c.win.hdc)
}

func (c *glContext) SwapInterval(interval int) {
	if GlfwConfig.PlatformContext.EXT_swap_control {
		if err := gl.WglSwapIntervalEXT(int32(interval)); err != nil {
			panic(err)
		}
	}
}

func (w *window) initWGL() error {
	if GlfwConfig.PlatformContext.inited {
		return nil
	}
	if err := checkValidContextConfig(&GlfwConfig.Hints.Context); err != nil {
		return err
	}
	pfd := windows.PIXELFORMATDESCRIPTOR{
		Version:   1,
		Flags:     _PFD_DRAW_TO_WINDOW | _PFD_SUPPORT_OPENGL | _PFD_DOUBLEBUFFER,
		PixelType: _PFD_TYPE_RGBA,
		ColorBits: 24,
	}
	pfd.Size = uint16(unsafe.Sizeof(pfd))

	format, err := windows.ChoosePixelFormat(w.hdc, &pfd)
	if err != nil {
		return err
	}
	if err := windows.SetPixelFormat(w.hdc, format, &pfd); err != nil {
		return err
	}

	rc, err := gl.WglCreateContext(w.hdc)
	if err != nil {
		return err
	}

	pdc := gl.WglGetCurrentDC()
	prc := gl.WglGetCurrentContext()

	if err := gl.WglMakeCurrent(w.hdc, rc); err != nil {
		_ = gl.WglMakeCurrent(pdc, prc)
		_ = gl.WglDeleteContext(rc)
		return err
	}

	// NOTE: Functions must be loaded first as they're needed to retrieve the
	//       extension string that tells us whether the functions are supported
	//
	// Interestingly, gl.WglGetProcAddress might return 0 after extensionSupportedWGL is called.
	gl.InitWGLExtensionFunctions()

	// NOTE: gl.Wgl_ARB_extensions_string and gl.Wgl_EXT_extensions_string are not
	//       checked below as we are already using them
	GlfwConfig.PlatformContext.ARB_multisample = extensionSupportedWGL("WGL_ARB_multisample")
	GlfwConfig.PlatformContext.ARB_framebuffer_sRGB = extensionSupportedWGL("WGL_ARB_framebuffer_sRGB")
	GlfwConfig.PlatformContext.EXT_framebuffer_sRGB = extensionSupportedWGL("WGL_EXT_framebuffer_sRGB")
	GlfwConfig.PlatformContext.ARB_create_context = extensionSupportedWGL("WGL_ARB_create_context")
	GlfwConfig.PlatformContext.ARB_create_context_profile = extensionSupportedWGL("WGL_ARB_create_context_profile")
	GlfwConfig.PlatformContext.EXT_create_context_es2_profile = extensionSupportedWGL("WGL_EXT_create_context_es2_profile")
	GlfwConfig.PlatformContext.ARB_create_context_robustness = extensionSupportedWGL("WGL_ARB_create_context_robustness")
	GlfwConfig.PlatformContext.ARB_create_context_no_error = extensionSupportedWGL("WGL_ARB_create_context_no_error")
	GlfwConfig.PlatformContext.EXT_swap_control = extensionSupportedWGL("WGL_EXT_swap_control")
	GlfwConfig.PlatformContext.EXT_colorspace = extensionSupportedWGL("WGL_EXT_colorspace")
	GlfwConfig.PlatformContext.ARB_pixel_format = extensionSupportedWGL("WGL_ARB_pixel_format")
	GlfwConfig.PlatformContext.ARB_context_flush_control = extensionSupportedWGL("WGL_ARB_context_flush_control")

	if err := gl.WglMakeCurrent(pdc, prc); err != nil {
		return err
	}
	if err := gl.WglDeleteContext(rc); err != nil {
		return err
	}
	GlfwConfig.PlatformContext.inited = true
	return nil
}

func extensionSupportedWGL(extension string) bool {
	var extensions string

	if gl.WglGetExtensionsStringARB_Available() {
		extensions = gl.WglGetExtensionsStringARB(gl.WglGetCurrentDC())
	} else if gl.WglGetExtensionsStringEXT_Available() {
		extensions = gl.WglGetExtensionsStringEXT()
	}

	if len(extensions) == 0 {
		return false
	}

	for _, str := range strings.Split(extensions, " ") {
		if extension == str {
			return true
		}
	}
	return false
}

func (w *window) createContextWGL(ctxconfig *CtxConfig, fbconfig *FbConfig) error {
	share := syscall.Handle(0)
	// var share _HGLRC
	// if ctxconfig.share != nil {
	// 	share = ctxconfig.share.context.platform.handle
	// }

	pixelFormat, err := w.choosePixelFormat(ctxconfig, fbconfig)
	if err != nil {
		return err
	}

	var pfd windows.PIXELFORMATDESCRIPTOR
	if _, err := windows.DescribePixelFormat(w.hdc, int32(pixelFormat), uint32(unsafe.Sizeof(pfd)), &pfd); err != nil {
		return err
	}

	if err := windows.SetPixelFormat(w.hdc, int32(pixelFormat), &pfd); err != nil {
		return err
	}

	if ctxconfig.Client == OpenGLAPI {
		if ctxconfig.forward && !GlfwConfig.PlatformContext.ARB_create_context {
			return fmt.Errorf("glfw: a forward compatible OpenGL context requested but WGL_ARB_create_context is unavailable: %w", VersionUnavailable)
		}

		if ctxconfig.profile != 0 && !GlfwConfig.PlatformContext.ARB_create_context_profile {
			return fmt.Errorf("glfw: OpenGL profile requested but WGL_ARB_create_context_profile is unavailable: %w", VersionUnavailable)
		}
	} else {
		if !GlfwConfig.PlatformContext.ARB_create_context || !GlfwConfig.PlatformContext.ARB_create_context_profile || !GlfwConfig.PlatformContext.EXT_create_context_es2_profile {
			return fmt.Errorf("glfw: OpenGL ES requested but WGL_ARB_create_context_es2_profile is unavailable: %w", APIUnavailable)
		}
	}

	if GlfwConfig.PlatformContext.ARB_create_context {
		var flags int32
		var mask int32
		if ctxconfig.Client == OpenGLAPI {
			if ctxconfig.forward {
				flags |= _WGL_CONTEXT_FORWARD_COMPATIBLE_BIT_ARB
			}

			if ctxconfig.profile == OpenGLCoreProfile {
				mask |= _WGL_CONTEXT_CORE_PROFILE_BIT_ARB
			} else if ctxconfig.profile == OpenGLCompatProfile {
				mask |= _WGL_CONTEXT_COMPATIBILITY_PROFILE_BIT_ARB
			}
		} else {
			mask |= _WGL_CONTEXT_ES2_PROFILE_BIT_EXT
		}

		if ctxconfig.debug {
			flags |= _WGL_CONTEXT_DEBUG_BIT_ARB
		}

		var attribs []int32
		if ctxconfig.robustness != 0 {
			if GlfwConfig.PlatformContext.ARB_create_context_robustness {
				if ctxconfig.robustness == NoResetNotification {
					attribs = append(attribs, _WGL_CONTEXT_RESET_NOTIFICATION_STRATEGY_ARB, _WGL_NO_RESET_NOTIFICATION_ARB)
				} else if ctxconfig.robustness == LoseContextOnReset {
					attribs = append(attribs, _WGL_CONTEXT_RESET_NOTIFICATION_STRATEGY_ARB, _WGL_LOSE_CONTEXT_ON_RESET_ARB)
				}
				flags |= _WGL_CONTEXT_ROBUST_ACCESS_BIT_ARB
			}
		}

		if ctxconfig.release != 0 {
			if GlfwConfig.PlatformContext.ARB_context_flush_control {
				if ctxconfig.release == ReleaseBehaviorNone {
					attribs = append(attribs, _WGL_CONTEXT_RELEASE_BEHAVIOR_ARB, _WGL_CONTEXT_RELEASE_BEHAVIOR_NONE_ARB)
				} else if ctxconfig.release == ReleaseBehaviorFlush {
					attribs = append(attribs, _WGL_CONTEXT_RELEASE_BEHAVIOR_ARB, _WGL_CONTEXT_RELEASE_BEHAVIOR_FLUSH_ARB)
				}
			}
		}

		if ctxconfig.noerror {
			if GlfwConfig.PlatformContext.ARB_create_context_no_error {
				attribs = append(attribs, _WGL_CONTEXT_OPENGL_NO_ERROR_ARB, 1)
			}
		}

		// NOTE: Only request an explicitly versioned context when necessary, as
		//       explicitly requesting version 1.0 does not always return the
		//       highest version supported by the driver
		if ctxconfig.Major != 1 || ctxconfig.Minor != 0 {
			attribs = append(attribs, _WGL_CONTEXT_MAJOR_VERSION_ARB, int32(ctxconfig.Major))
			attribs = append(attribs, _WGL_CONTEXT_MINOR_VERSION_ARB, int32(ctxconfig.Minor))
		}

		if flags != 0 {
			attribs = append(attribs, _WGL_CONTEXT_FLAGS_ARB, flags)
		}

		if mask != 0 {
			attribs = append(attribs, _WGL_CONTEXT_PROFILE_MASK_ARB, mask)
		}

		attribs = append(attribs, 0, 0)

		var err error
		w.context.platform.handle, err = gl.WglCreateContextAttribsARB(w.hdc, share, &attribs[0])
		if err != nil {
			return err
		}
	} else {
		var err error
		w.context.platform.handle, err = gl.WglCreateContext(w.hdc)
		if err != nil {
			return err
		}

		if share != 0 {
			if err := gl.WglShareLists(share, w.context.platform.handle); err != nil {
				return err
			}
		}
	}

	w.Context.makeCurrent = makeContextCurrentWGL
	w.Context.swapBuffers = swapBuffersWGL
	w.Context.swapInterval = swapIntervalWGL
	w.Context.extensionSupported = extensionSupportedWGL
	w.Context.getProcAddress = getProcAddressWGL
	w.Context.destroy = destroyContextWGL

	return nil
}

func (w *window) choosePixelFormat(ctxconfig *CtxConfig, fbconfig_ *FbConfig) (int, error) {
	var nativeCount int32
	var attribs []int32

	if GlfwConfig.PlatformContext.ARB_pixel_format {
		var attrib int32 = _WGL_NUMBER_PIXEL_FORMATS_ARB
		if err := gl.WglGetPixelFormatAttribivARB(w.hdc, 1, 0, 1, &attrib, &nativeCount); err != nil {
			return 0, err
		}

		attribs = append(attribs,
			_WGL_SUPPORT_OPENGL_ARB,
			_WGL_DRAW_TO_WINDOW_ARB,
			_WGL_PIXEL_TYPE_ARB,
			_WGL_ACCELERATION_ARB,
			_WGL_RED_BITS_ARB,
			_WGL_RED_SHIFT_ARB,
			_WGL_GREEN_BITS_ARB,
			_WGL_GREEN_SHIFT_ARB,
			_WGL_BLUE_BITS_ARB,
			_WGL_BLUE_SHIFT_ARB,
			_WGL_ALPHA_BITS_ARB,
			_WGL_ALPHA_SHIFT_ARB,
			_WGL_DEPTH_BITS_ARB,
			_WGL_STENCIL_BITS_ARB,
			_WGL_ACCUM_BITS_ARB,
			_WGL_ACCUM_RED_BITS_ARB,
			_WGL_ACCUM_GREEN_BITS_ARB,
			_WGL_ACCUM_BLUE_BITS_ARB,
			_WGL_ACCUM_ALPHA_BITS_ARB,
			_WGL_AUX_BUFFERS_ARB,
			_WGL_STEREO_ARB,
			_WGL_DOUBLE_BUFFER_ARB)

		if GlfwConfig.PlatformContext.ARB_multisample {
			attribs = append(attribs, _WGL_SAMPLES_ARB)
		}

		if ctxconfig.Client == OpenGLAPI {
			if GlfwConfig.PlatformContext.ARB_framebuffer_sRGB || GlfwConfig.PlatformContext.EXT_framebuffer_sRGB {
				attribs = append(attribs, _WGL_FRAMEBUFFER_SRGB_CAPABLE_ARB)
			}
		} else {
			if GlfwConfig.PlatformContext.EXT_colorspace {
				attribs = append(attribs, _WGL_COLORSPACE_EXT)
			}
		}
	} else {
		c, err := windows.DescribePixelFormat(w.hdc, 1, uint32(unsafe.Sizeof(windows.PIXELFORMATDESCRIPTOR{})), nil)
		if err != nil {
			return 0, err
		}
		nativeCount = c
	}

	usableConfigs := make([]*FbConfig, 0, nativeCount)
	for i := int32(0); i < nativeCount; i++ {
		var u FbConfig
		pixelFormat := uintptr(i) + 1

		if GlfwConfig.PlatformContext.ARB_pixel_format {
			// Get pixel format attributes through "modern" extension
			values := make([]int32, len(attribs))
			if err := gl.WglGetPixelFormatAttribivARB(w.hdc, int32(pixelFormat), 0, uint32(len(attribs)), &attribs[0], &values[0]); err != nil {
				return 0, err
			}

			findAttribValue := func(attrib int32) int32 {
				return findPixelFormatAttribValue(attribs, values, attrib)
			}

			if findAttribValue(_WGL_SUPPORT_OPENGL_ARB) == 0 || findAttribValue(_WGL_DRAW_TO_WINDOW_ARB) == 0 {
				continue
			}

			if findAttribValue(_WGL_PIXEL_TYPE_ARB) != _WGL_TYPE_RGBA_ARB {
				continue
			}

			if findAttribValue(_WGL_ACCELERATION_ARB) == _WGL_NO_ACCELERATION_ARB {
				continue
			}

			if (findAttribValue(_WGL_DOUBLE_BUFFER_ARB) != 0) != fbconfig_.doublebuffer {
				continue
			}

			u.redBits = int(findAttribValue(_WGL_RED_BITS_ARB))
			u.greenBits = int(findAttribValue(_WGL_GREEN_BITS_ARB))
			u.blueBits = int(findAttribValue(_WGL_BLUE_BITS_ARB))
			u.alphaBits = int(findAttribValue(_WGL_ALPHA_BITS_ARB))

			u.depthBits = int(findAttribValue(_WGL_DEPTH_BITS_ARB))
			u.stencilBits = int(findAttribValue(_WGL_STENCIL_BITS_ARB))

			u.accumRedBits = int(findAttribValue(_WGL_ACCUM_RED_BITS_ARB))
			u.accumGreenBits = int(findAttribValue(_WGL_ACCUM_GREEN_BITS_ARB))
			u.accumBlueBits = int(findAttribValue(_WGL_ACCUM_BLUE_BITS_ARB))
			u.accumAlphaBits = int(findAttribValue(_WGL_ACCUM_ALPHA_BITS_ARB))

			u.auxBuffers = int(findAttribValue(_WGL_AUX_BUFFERS_ARB))

			if findAttribValue(_WGL_STEREO_ARB) != 0 {
				u.stereo = true
			}

			if GlfwConfig.PlatformContext.ARB_multisample {
				u.samples = int(findAttribValue(_WGL_SAMPLES_ARB))
			}

			if ctxconfig.Client == OpenGLAPI {
				if GlfwConfig.PlatformContext.ARB_framebuffer_sRGB || GlfwConfig.PlatformContext.EXT_framebuffer_sRGB {
					if findAttribValue(_WGL_FRAMEBUFFER_SRGB_CAPABLE_ARB) != 0 {
						u.sRGB = true
					}
				}
			} else {
				if GlfwConfig.PlatformContext.EXT_colorspace {
					if findAttribValue(_WGL_COLORSPACE_EXT) == _WGL_COLORSPACE_SRGB_EXT {
						u.sRGB = true
					}
				}
			}
		} else {
			// Get pixel format attributes through legacy PFDs

			var pfd windows.PIXELFORMATDESCRIPTOR
			if _, err := windows.DescribePixelFormat(w.hdc, int32(pixelFormat), uint32(unsafe.Sizeof(pfd)), &pfd); err != nil {
				return 0, err
			}

			if pfd.Flags&_PFD_DRAW_TO_WINDOW == 0 || pfd.Flags&_PFD_SUPPORT_OPENGL == 0 {
				continue
			}

			if pfd.Flags&_PFD_GENERIC_ACCELERATED == 0 && pfd.Flags&_PFD_GENERIC_FORMAT != 0 {
				continue
			}

			if pfd.PixelType != _PFD_TYPE_RGBA {
				continue
			}

			if (pfd.Flags&_PFD_DOUBLEBUFFER != 0) != fbconfig_.doublebuffer {
				continue
			}

			u.redBits = int(pfd.RedBits)
			u.greenBits = int(pfd.GreenBits)
			u.blueBits = int(pfd.BlueBits)
			u.alphaBits = int(pfd.AlphaBits)

			u.depthBits = int(pfd.DepthBits)
			u.stencilBits = int(pfd.StencilBits)

			u.accumRedBits = int(pfd.AccumRedBits)
			u.accumGreenBits = int(pfd.AccumGreenBits)
			u.accumBlueBits = int(pfd.AccumBlueBits)
			u.accumAlphaBits = int(pfd.AccumAlphaBits)

			u.auxBuffers = int(pfd.AuxBuffers)

			if pfd.Flags&_PFD_STEREO != 0 {
				u.stereo = true
			}
		}

		u.handle = pixelFormat
		usableConfigs = append(usableConfigs, &u)
	}

	if len(usableConfigs) == 0 {
		return 0, fmt.Errorf("glfw: the driver does not appear to support OpenGL")
	}

	closest := chooseFBConfig(fbconfig_, usableConfigs)
	if closest == nil {
		return 0, fmt.Errorf("glfw: failed to find a suitable pixel format")
	}

	return int(closest.handle), nil
}

func (c *glContext) RenderTarget() (gpu.RenderTarget, error) {
	return gpu.OpenGLRenderTarget{}, nil
}

func (c *glContext) API() gpu.API {
	return gpu.OpenGL{}
}

func (c *glContext) Present() error {
	// Assume the caller already locked the context.
	//C.glFlush(c.glFlush)
	return nil
}

// func (w *window) NewContext() (mado.Context, error) {
// 	return newContext(w)
// }

func findPixelFormatAttribValue(attribs []int32, values []int32, attrib int32) int32 {
	for i := range attribs {
		if attribs[i] == attrib {
			return values[i]
		}
	}
	return 0
}

func checkValidContextConfig(ctxconfig *CtxConfig) error {
	// if ctxconfig.share != nil {
	// 	if ctxconfig.Client == NoAPI || ctxconfig.share.context.client == NoAPI {
	// 		return NoWindowContext
	// 	}
	// }

	if ctxconfig.source != NativeContextAPI &&
		ctxconfig.source != EGLContextAPI &&
		ctxconfig.source != OSMesaContextAPI {
		return fmt.Errorf("glfw: invalid context creation API 0x%08X: %w", ctxconfig.source, InvalidEnum)
	}

	if ctxconfig.Client != NoAPI &&
		ctxconfig.Client != OpenGLAPI &&
		ctxconfig.Client != OpenGLESAPI {
		return fmt.Errorf("glfw: invalid client API 0x%08X: %w", ctxconfig.Client, InvalidEnum)
	}

	if ctxconfig.Client == OpenGLAPI {
		if (ctxconfig.Major < 1 || ctxconfig.Minor < 0) ||
			(ctxconfig.Major == 1 && ctxconfig.Minor > 5) ||
			(ctxconfig.Major == 2 && ctxconfig.Minor > 1) ||
			(ctxconfig.Major == 3 && ctxconfig.Minor > 3) {
			// OpenGL 1.0 is the smallest valid version
			// OpenGL 1.x series ended with version 1.5
			// OpenGL 2.x series ended with version 2.1
			// OpenGL 3.x series ended with version 3.3
			// For now, let everything else through

			return fmt.Errorf("glfw: invalid OpenGL version %d.%d: %w", ctxconfig.Major, ctxconfig.Minor, InvalidValue)
		}

		if ctxconfig.profile != 0 {
			if ctxconfig.profile != OpenGLCoreProfile && ctxconfig.profile != OpenGLCompatProfile {
				return fmt.Errorf("glfw: invalid OpenGL profile 0x%08X: %w", ctxconfig.profile, InvalidEnum)
			}

			if ctxconfig.Major <= 2 || (ctxconfig.Major == 3 && ctxconfig.Minor < 2) {
				// Desktop OpenGL context profiles are only defined for version 3.2
				// and above

				return fmt.Errorf("glfw: context profiles are only defined for OpenGL version 3.2 and above: %w", InvalidValue)
			}
		}

		if ctxconfig.forward && ctxconfig.Major <= 2 {
			// Forward-compatible contexts are only defined for OpenGL version 3.0 and above
			return fmt.Errorf("glfw: forward-compatibility is only defined for OpenGL version 3.0 and above: %w", InvalidValue)
		}
	} else if ctxconfig.Client == OpenGLESAPI {
		if ctxconfig.Major < 1 || ctxconfig.Minor < 0 ||
			(ctxconfig.Major == 1 && ctxconfig.Minor > 1) ||
			(ctxconfig.Major == 2 && ctxconfig.Minor > 0) {
			// OpenGL ES 1.0 is the smallest valid version
			// OpenGL ES 1.x series ended with version 1.1
			// OpenGL ES 2.x series ended with version 2.0
			// For now, let everything else through

			return fmt.Errorf("glfw: invalid OpenGL ES version %d.%d: %w", ctxconfig.Major, ctxconfig.Minor, InvalidValue)
		}
	}

	if ctxconfig.robustness != 0 {
		if ctxconfig.robustness != NoResetNotification && ctxconfig.robustness != LoseContextOnReset {
			return fmt.Errorf("glfw: invalid context robustness mode 0x%08X: %w", ctxconfig.robustness, InvalidEnum)
		}
	}

	if ctxconfig.release != 0 {
		if ctxconfig.release != ReleaseBehaviorNone && ctxconfig.release != ReleaseBehaviorFlush {
			return fmt.Errorf("glfw: invalid context release behavior 0x%08X: %w", ctxconfig.release, InvalidEnum)
		}
	}

	return nil
}

func chooseFBConfig(desired *FbConfig, alternatives []*FbConfig) *FbConfig {
	leastMissing := math.MaxInt32
	leastColorDiff := math.MaxInt32
	leastExtraDiff := math.MaxInt32

	var closest *FbConfig
	for _, current := range alternatives {
		if desired.stereo && !current.stereo {
			// Stereo is a hard constraint
			continue
		}

		// Count number of missing buffers
		missing := 0

		if desired.alphaBits > 0 && current.alphaBits == 0 {
			missing++
		}

		if desired.depthBits > 0 && current.depthBits == 0 {
			missing++
		}

		if desired.stencilBits > 0 && current.stencilBits == 0 {
			missing++
		}

		if desired.auxBuffers > 0 &&
			current.auxBuffers < desired.auxBuffers {
			missing += desired.auxBuffers - current.auxBuffers
		}

		if desired.samples > 0 && current.samples == 0 {
			// Technically, several multisampling buffers could be
			// involved, but that's a lower level implementation detail and
			// not important to us here, so we count them as one
			missing++
		}

		if desired.transparent != current.transparent {
			missing++
		}

		// These polynomials make many small channel size differences matter
		// less than one large channel size difference

		// Calculate color channel size difference value
		colorDiff := 0

		if desired.redBits != DontCare {
			colorDiff += (desired.redBits - current.redBits) *
				(desired.redBits - current.redBits)
		}

		if desired.greenBits != DontCare {
			colorDiff += (desired.greenBits - current.greenBits) *
				(desired.greenBits - current.greenBits)
		}

		if desired.blueBits != DontCare {
			colorDiff += (desired.blueBits - current.blueBits) *
				(desired.blueBits - current.blueBits)
		}

		// Calculate non-color channel size difference value
		extraDiff := 0

		if desired.alphaBits != DontCare {
			extraDiff += (desired.alphaBits - current.alphaBits) *
				(desired.alphaBits - current.alphaBits)
		}

		if desired.depthBits != DontCare {
			extraDiff += (desired.depthBits - current.depthBits) *
				(desired.depthBits - current.depthBits)
		}

		if desired.stencilBits != DontCare {
			extraDiff += (desired.stencilBits - current.stencilBits) *
				(desired.stencilBits - current.stencilBits)
		}

		if desired.accumRedBits != DontCare {
			extraDiff += (desired.accumRedBits - current.accumRedBits) *
				(desired.accumRedBits - current.accumRedBits)
		}

		if desired.accumGreenBits != DontCare {
			extraDiff += (desired.accumGreenBits - current.accumGreenBits) *
				(desired.accumGreenBits - current.accumGreenBits)
		}

		if desired.accumBlueBits != DontCare {
			extraDiff += (desired.accumBlueBits - current.accumBlueBits) *
				(desired.accumBlueBits - current.accumBlueBits)
		}

		if desired.accumAlphaBits != DontCare {
			extraDiff += (desired.accumAlphaBits - current.accumAlphaBits) *
				(desired.accumAlphaBits - current.accumAlphaBits)
		}

		if desired.samples != DontCare {
			extraDiff += (desired.samples - current.samples) *
				(desired.samples - current.samples)
		}

		if desired.sRGB && !current.sRGB {
			extraDiff++
		}

		// Figure out if the current one is better than the best one found so far
		// Least number of missing buffers is the most important heuristic,
		// then color buffer size match and lastly size match for other buffers

		if missing < leastMissing {
			closest = current
		} else if missing == leastMissing {
			if (colorDiff < leastColorDiff) || (colorDiff == leastColorDiff && extraDiff < leastExtraDiff) {
				closest = current
			}
		}

		if current == closest {
			leastMissing = missing
			leastColorDiff = colorDiff
			leastExtraDiff = extraDiff
		}
	}

	return closest
}
