// SPDX-License-Identifier: Unlicense OR MIT

//go:build windows
// +build windows

package mswindows

import (
	"fmt"
	"math"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"syscall"

	"unsafe"

	"github.com/kanryu/mado"
	"github.com/kanryu/mado/gpu"
	"github.com/kanryu/mado/internal/gl"
	"github.com/kanryu/mado/mswindows/internal/windows"
	winsyscall "golang.org/x/sys/windows"
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
	intSize = 32 << (^uint(0) >> 63)

	_BI_BITFIELDS                                                   = 3
	_CCHDEVICENAME                                                  = 32
	_CCHFORMNAME                                                    = 32
	_CDS_TEST                                                       = 0x00000002
	_CDS_FULLSCREEN                                                 = 0x00000004
	_CS_HREDRAW                                                     = 0x00000002
	_CS_OWNDC                                                       = 0x00000020
	_CS_VREDRAW                                                     = 0x00000001
	_CW_USEDEFAULT                                                  = int32(^0x7fffffff)
	_DBT_DEVTYP_DEVICEINTERFACE                                     = 0x00000005
	_DEVICE_NOTIFY_WINDOW_HANDLE                                    = 0x00000000
	_DIB_RGB_COLORS                                                 = 0
	_DISP_CHANGE_SUCCESSFUL                                         = 0
	_DISP_CHANGE_RESTART                                            = 1
	_DISP_CHANGE_FAILED                                             = -1
	_DISP_CHANGE_BADMODE                                            = -2
	_DISP_CHANGE_NOTUPDATED                                         = -3
	_DISP_CHANGE_BADFLAGS                                           = -4
	_DISP_CHANGE_BADPARAM                                           = -5
	_DISP_CHANGE_BADDUALVIEW                                        = -6
	_DISPLAY_DEVICE_ACTIVE                                          = 0x00000001
	_DISPLAY_DEVICE_MODESPRUNED                                     = 0x08000000
	_DISPLAY_DEVICE_PRIMARY_DEVICE                                  = 0x00000004
	_DM_BITSPERPEL                                                  = 0x00040000
	_DM_PELSWIDTH                                                   = 0x00080000
	_DM_PELSHEIGHT                                                  = 0x00100000
	_DM_DISPLAYFREQUENCY                                            = 0x00400000
	_DWM_BB_BLURREGION                                              = 0x00000002
	_DWM_BB_ENABLE                                                  = 0x00000001
	_EDS_ROTATEDMODE                                                = 0x00000004
	_ENUM_CURRENT_SETTINGS                        uint32            = 0xffffffff
	_GCLP_HICON                                                     = -14
	_GCLP_HICONSM                                                   = -34
	_GET_MODULE_HANDLE_EX_FLAG_FROM_ADDRESS                         = 0x00000004
	_GET_MODULE_HANDLE_EX_FLAG_UNCHANGED_REFCOUNT                   = 0x00000002
	_GWL_EXSTYLE                                                    = -20
	_GWL_STYLE                                                      = -16
	_HTCLIENT                                                       = 1
	_HORZSIZE                                                       = 4
	_HWND_NOTOPMOST                               winsyscall.Handle = (1 << intSize) - 2
	_HWND_TOP                                     winsyscall.Handle = 0
	_HWND_TOPMOST                                 winsyscall.Handle = (1 << intSize) - 1
	_ICON_BIG                                                       = 1
	_ICON_SMALL                                                     = 0
	_IDC_ARROW                                                      = 32512
	_IDI_APPLICATION                                                = 32512
	_IMAGE_CURSOR                                                   = 2
	_IMAGE_ICON                                                     = 1
	_KF_ALTDOWN                                                     = 0x2000
	_KF_DLGMODE                                                     = 0x0800
	_KF_EXTENDED                                                    = 0x0100
	_KF_MENUMODE                                                    = 0x1000
	_KF_REPEAT                                                      = 0x4000
	_KF_UP                                                          = 0x8000
	_LOGPIXELSX                                                     = 88
	_LOGPIXELSY                                                     = 90
	_LR_DEFAULTSIZE                                                 = 0x0040
	_LR_SHARED                                                      = 0x8000
	_LWA_ALPHA                                                      = 0x00000002
	_MAPVK_VK_TO_VSC                                                = 0
	_MAPVK_VSC_TO_VK                                                = 1
	_MONITOR_DEFAULTTONEAREST                                       = 0x00000002
	_MOUSE_MOVE_ABSOLUTE                                            = 0x01
	_MSGFLT_ALLOW                                                   = 1
	_OCR_CROSS                                                      = 32515
	_OCR_HAND                                                       = 32649
	_OCR_IBEAM                                                      = 32513
	_OCR_NO                                                         = 32648
	_OCR_NORMAL                                                     = 32512
	_OCR_SIZEALL                                                    = 32646
	_OCR_SIZENESW                                                   = 32643
	_OCR_SIZENS                                                     = 32645
	_OCR_SIZENWSE                                                   = 32642
	_OCR_SIZEWE                                                     = 32644
	_PM_NOREMOVE                                                    = 0x0000
	_PM_REMOVE                                                      = 0x0001
	_PFD_DRAW_TO_WINDOW                                             = 0x00000004
	_PFD_DOUBLEBUFFER                                               = 0x00000001
	_PFD_GENERIC_ACCELERATED                                        = 0x00001000
	_PFD_GENERIC_FORMAT                                             = 0x00000040
	_PFD_STEREO                                                     = 0x00000002
	_PFD_SUPPORT_OPENGL                                             = 0x00000020
	_PFD_TYPE_RGBA                                                  = 0
	_QS_ALLEVENTS                                                   = _QS_INPUT | _QS_POSTMESSAGE | _QS_TIMER | _QS_PAINT | _QS_HOTKEY
	_QS_HOTKEY                                                      = 0x0080
	_QS_INPUT                                                       = _QS_MOUSE | _QS_KEY | _QS_RAWINPUT
	_QS_KEY                                                         = 0x0001
	_QS_MOUSE                                                       = _QS_MOUSEMOVE | _QS_MOUSEBUTTON
	_QS_MOUSEBUTTON                                                 = 0x0004
	_QS_MOUSEMOVE                                                   = 0x0002
	_QS_PAINT                                                       = 0x0020
	_QS_POSTMESSAGE                                                 = 0x0008
	_QS_RAWINPUT                                                    = 0x0400
	_QS_TIMER                                                       = 0x0010
	_RID_INPUT                                                      = 0x10000003
	_RIDEV_REMOVE                                                   = 0x00000001
	_SC_KEYMENU                                                     = 0xf100
	_SC_MONITORPOWER                                                = 0xf170
	_SC_SCREENSAVE                                                  = 0xf140
	_SIZE_MAXIMIZED                                                 = 2
	_SIZE_MINIMIZED                                                 = 1
	_SIZE_RESTORED                                                  = 0
	_SM_CXICON                                                      = 11
	_SM_CXSMICON                                                    = 49
	_SM_CYCAPTION                                                   = 4
	_SM_CYICON                                                      = 12
	_SM_CYSMICON                                                    = 50
	_SPI_GETFOREGROUNDLOCKTIMEOUT                                   = 0x2000
	_SPI_GETMOUSETRAILS                                             = 94
	_SPI_SETFOREGROUNDLOCKTIMEOUT                                   = 0x2001
	_SPI_SETMOUSETRAILS                                             = 93
	_SPIF_SENDCHANGE                                                = _SPIF_SENDWININICHANGE
	_SPIF_SENDWININICHANGE                                          = 2
	_SW_HIDE                                                        = 0
	_SW_MAXIMIZE                                                    = _SW_SHOWMAXIMIZED
	_SW_MINIMIZE                                                    = 6
	_SW_RESTORE                                                     = 9
	_SW_SHOWNA                                                      = 8
	_SW_SHOWMAXIMIZED                                               = 3
	_SWP_FRAMECHANGED                                               = 0x0020
	_SWP_NOACTIVATE                                                 = 0x0010
	_SWP_NOCOPYBITS                                                 = 0x0100
	_SWP_NOMOVE                                                     = 0x0002
	_SWP_NOOWNERZORDER                                              = 0x0200
	_SWP_NOSIZE                                                     = 0x0001
	_SWP_NOZORDER                                                   = 0x0004
	_SWP_SHOWWINDOW                                                 = 0x0040
	_TLS_OUT_OF_INDEXES                           uint32            = 0xffffffff
	_TME_LEAVE                                                      = 0x00000002
	_UNICODE_NOCHAR                                                 = 0xffff
	_USER_DEFAULT_SCREEN_DPI                                        = 96
	_VERTSIZE                                                       = 6
	_VK_ADD                                                         = 0x6B
	_VK_CAPITAL                                                     = 0x14
	_VK_CONTROL                                                     = 0x11
	_VK_DECIMAL                                                     = 0x6E
	_VK_DIVIDE                                                      = 0x6F
	_VK_LSHIFT                                                      = 0xA0
	_VK_LWIN                                                        = 0x5B
	_VK_MENU                                                        = 0x12
	_VK_MULTIPLY                                                    = 0x6A
	_VK_NUMLOCK                                                     = 0x90
	_VK_NUMPAD0                                                     = 0x60
	_VK_NUMPAD1                                                     = 0x61
	_VK_NUMPAD2                                                     = 0x62
	_VK_NUMPAD3                                                     = 0x63
	_VK_NUMPAD4                                                     = 0x64
	_VK_NUMPAD5                                                     = 0x65
	_VK_NUMPAD6                                                     = 0x66
	_VK_NUMPAD7                                                     = 0x67
	_VK_NUMPAD8                                                     = 0x68
	_VK_NUMPAD9                                                     = 0x69
	_VK_PROCESSKEY                                                  = 0xE5
	_VK_RSHIFT                                                      = 0xA1
	_VK_RWIN                                                        = 0x5C
	_VK_SHIFT                                                       = 0x10
	_VK_SNAPSHOT                                                    = 0x2C
	_VK_SUBTRACT                                                    = 0x6D
	_WAIT_FAILED                                                    = 0xffffffff
	_WHEEL_DELTA                                                    = 120
	_WGL_ACCUM_BITS_ARB                                             = 0x201D
	_WGL_ACCELERATION_ARB                                           = 0x2003
	_WGL_ACCUM_ALPHA_BITS_ARB                                       = 0x2021
	_WGL_ACCUM_BLUE_BITS_ARB                                        = 0x2020
	_WGL_ACCUM_GREEN_BITS_ARB                                       = 0x201F
	_WGL_ACCUM_RED_BITS_ARB                                         = 0x201E
	_WGL_AUX_BUFFERS_ARB                                            = 0x2024
	_WGL_ALPHA_BITS_ARB                                             = 0x201B
	_WGL_ALPHA_SHIFT_ARB                                            = 0x201C
	_WGL_BLUE_BITS_ARB                                              = 0x2019
	_WGL_BLUE_SHIFT_ARB                                             = 0x201A
	_WGL_COLOR_BITS_ARB                                             = 0x2014
	_WGL_COLORSPACE_EXT                                             = 0x309D
	_WGL_COLORSPACE_SRGB_EXT                                        = 0x3089
	_WGL_CONTEXT_COMPATIBILITY_PROFILE_BIT_ARB                      = 0x00000002
	_WGL_CONTEXT_CORE_PROFILE_BIT_ARB                               = 0x00000001
	_WGL_CONTEXT_DEBUG_BIT_ARB                                      = 0x0001
	_WGL_CONTEXT_ES2_PROFILE_BIT_EXT                                = 0x00000004
	_WGL_CONTEXT_FLAGS_ARB                                          = 0x2094
	_WGL_CONTEXT_FORWARD_COMPATIBLE_BIT_ARB                         = 0x0002
	_WGL_CONTEXT_MAJOR_VERSION_ARB                                  = 0x2091
	_WGL_CONTEXT_MINOR_VERSION_ARB                                  = 0x2092
	_WGL_CONTEXT_OPENGL_NO_ERROR_ARB                                = 0x31B3
	_WGL_CONTEXT_PROFILE_MASK_ARB                                   = 0x9126
	_WGL_CONTEXT_RELEASE_BEHAVIOR_ARB                               = 0x2097
	_WGL_CONTEXT_RELEASE_BEHAVIOR_NONE_ARB                          = 0x0000
	_WGL_CONTEXT_RELEASE_BEHAVIOR_FLUSH_ARB                         = 0x2098
	_WGL_CONTEXT_RESET_NOTIFICATION_STRATEGY_ARB                    = 0x8256
	_WGL_CONTEXT_ROBUST_ACCESS_BIT_ARB                              = 0x00000004
	_WGL_DEPTH_BITS_ARB                                             = 0x2022
	_WGL_DRAW_TO_BITMAP_ARB                                         = 0x2002
	_WGL_DRAW_TO_WINDOW_ARB                                         = 0x2001
	_WGL_DOUBLE_BUFFER_ARB                                          = 0x2011
	_WGL_FRAMEBUFFER_SRGB_CAPABLE_ARB                               = 0x20A9
	_WGL_GREEN_BITS_ARB                                             = 0x2017
	_WGL_GREEN_SHIFT_ARB                                            = 0x2018
	_WGL_LOSE_CONTEXT_ON_RESET_ARB                                  = 0x8252
	_WGL_NEED_PALETTE_ARB                                           = 0x2004
	_WGL_NEED_SYSTEM_PALETTE_ARB                                    = 0x2005
	_WGL_NO_ACCELERATION_ARB                                        = 0x2025
	_WGL_NO_RESET_NOTIFICATION_ARB                                  = 0x8261
	_WGL_NUMBER_OVERLAYS_ARB                                        = 0x2008
	_WGL_NUMBER_PIXEL_FORMATS_ARB                                   = 0x2000
	_WGL_NUMBER_UNDERLAYS_ARB                                       = 0x2009
	_WGL_PIXEL_TYPE_ARB                                             = 0x2013
	_WGL_RED_BITS_ARB                                               = 0x2015
	_WGL_RED_SHIFT_ARB                                              = 0x2016
	_WGL_SAMPLES_ARB                                                = 0x2042
	_WGL_SHARE_ACCUM_ARB                                            = 0x200E
	_WGL_SHARE_DEPTH_ARB                                            = 0x200C
	_WGL_SHARE_STENCIL_ARB                                          = 0x200D
	_WGL_STENCIL_BITS_ARB                                           = 0x2023
	_WGL_STEREO_ARB                                                 = 0x2012
	_WGL_SUPPORT_GDI_ARB                                            = 0x200F
	_WGL_SUPPORT_OPENGL_ARB                                         = 0x2010
	_WGL_SWAP_LAYER_BUFFERS_ARB                                     = 0x2006
	_WGL_SWAP_METHOD_ARB                                            = 0x2007
	_WGL_TRANSPARENT_ARB                                            = 0x200A
	_WGL_TRANSPARENT_ALPHA_VALUE_ARB                                = 0x203A
	_WGL_TRANSPARENT_BLUE_VALUE_ARB                                 = 0x2039
	_WGL_TRANSPARENT_GREEN_VALUE_ARB                                = 0x2038
	_WGL_TRANSPARENT_INDEX_VALUE_ARB                                = 0x203B
	_WGL_TRANSPARENT_RED_VALUE_ARB                                  = 0x2037
	_WGL_TYPE_RGBA_ARB                                              = 0x202B

	EmptyHandle = winsyscall.Handle(0)
)

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

func (e ErrorCode) Error() string {
	switch e {
	case NotInitialized:
		return "the GLFW library is not initialized"
	case NoCurrentContext:
		return "there is no current context"
	case InvalidEnum:
		return "invalid argument for enum parameter"
	case InvalidValue:
		return "invalid value for parameter"
	case OutOfMemory:
		return "out of memory"
	case APIUnavailable:
		return "the requested API is unavailable"
	case VersionUnavailable:
		return "the requested API version is unavailable"
	case PlatformError:
		return "a platform-specific error occurred"
	case FormatUnavailable:
		return "the requested format is unavailable"
	case NoWindowContext:
		return "the specified window has no context"
	default:
		return fmt.Sprintf("GLFW error (%d)", e)
	}
}

var _ mado.Context = (*glContext)(nil)

type glContext struct {
	win          *window
	hglrc        winsyscall.Handle
	prevHdc      winsyscall.Handle
	prevHglrc    winsyscall.Handle
	context      mado.GlfwContext
	doublebuffer bool
}

func init() {
	mado.GlfwConfigInit()
	drivers = append(drivers, gpuAPI{
		priority: 2,
		name:     "opengl",
		initializer: func(w *window) (mado.Context, error) {
			ctx := &glContext{win: w}
			err := ctx.createContext()
			return ctx, err
		},
	})
}

func (c *glContext) Release() {
}

func (c *glContext) Refresh() error {
	return nil
}

func (c *glContext) Lock() error {
	return gl.WglMakeCurrent(c.win.hdc, c.hglrc)
}

func (c *glContext) Unlock() {
	gl.WglMakeCurrent(winsyscall.Handle(0), winsyscall.Handle(0))
}

func (c *glContext) createContext() error {
	// OpenGL contexts are implicit and thread-local. Lock the OS thread.
	runtime.LockOSThread()

	if err := c.initWGL(); err != nil {
		panic(err)
	}
	if hglrc, err := c.createContextWGL(&mado.GlfwConfig.Hints.Context, &mado.GlfwConfig.Hints.Framebuffer); err != nil {
		panic(err)
	} else {
		c.hglrc = hglrc
	}
	if err := c.refreshContextAttribs(&mado.GlfwConfig.Hints.Context); err != nil {
		return err
	}

	return nil
}

func (c *glContext) SwapBuffers() error {
	return windows.SwapBuffers(c.win.hdc)
}

func (c *glContext) SwapInterval(interval int) {
	if mado.GlfwConfig.PlatformContext.EXT_swap_control {
		if err := gl.WglSwapIntervalEXT(int32(interval)); err != nil {
			panic(err)
		}
	}
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

func (c *glContext) initWGL() error {
	if mado.GlfwConfig.PlatformContext.Inited {
		return nil
	}
	if err := checkValidContextConfig(&mado.GlfwConfig.Hints.Context); err != nil {
		return err
	}
	createHelperWindow()
	pfd := windows.PIXELFORMATDESCRIPTOR{
		Version:   1,
		Flags:     _PFD_DRAW_TO_WINDOW | _PFD_SUPPORT_OPENGL | _PFD_DOUBLEBUFFER,
		PixelType: _PFD_TYPE_RGBA,
		ColorBits: 24,
	}
	pfd.Size = uint16(unsafe.Sizeof(pfd))
	dc, err := windows.GetDC(resources.helperHwnd)
	if err != nil {
		return err
	}

	format, err := windows.ChoosePixelFormat(dc, &pfd)
	if err != nil {
		return err
	}
	if err := windows.SetPixelFormat(dc, format, &pfd); err != nil {
		return err
	}

	rc, err := gl.WglCreateContext(dc)
	if err != nil {
		return err
	}

	pdc := gl.WglGetCurrentDC()
	prc := gl.WglGetCurrentContext()

	if err := gl.WglMakeCurrent(dc, rc); err != nil {
		_ = gl.WglMakeCurrent(pdc, prc)
		_ = gl.WglDeleteContext(rc)
		return err
	}

	// NOTE: Functions must be loaded first as they're needed to retrieve the
	//       extension string that tells us whether the functions are supported
	//
	// Interestingly, gl.WglGetProcAddress might return 0 afterc.ExtensionSupportedWGL is called.
	gl.InitWGLExtensionFunctions()

	// NOTE: gl.Wgl_ARB_extensions_string and gl.Wgl_EXT_extensions_string are not
	//       checked below as we are already using them
	mado.GlfwConfig.PlatformContext.ARB_multisample = ExtensionSupportedWGL("WGL_ARB_multisample")
	mado.GlfwConfig.PlatformContext.ARB_framebuffer_sRGB = ExtensionSupportedWGL("WGL_ARB_framebuffer_sRGB")
	mado.GlfwConfig.PlatformContext.EXT_framebuffer_sRGB = ExtensionSupportedWGL("WGL_EXT_framebuffer_sRGB")
	mado.GlfwConfig.PlatformContext.ARB_create_context = ExtensionSupportedWGL("WGL_ARB_create_context")
	mado.GlfwConfig.PlatformContext.ARB_create_context_profile = ExtensionSupportedWGL("WGL_ARB_create_context_profile")
	mado.GlfwConfig.PlatformContext.EXT_create_context_es2_profile = ExtensionSupportedWGL("WGL_EXT_create_context_es2_profile")
	mado.GlfwConfig.PlatformContext.ARB_create_context_robustness = ExtensionSupportedWGL("WGL_ARB_create_context_robustness")
	mado.GlfwConfig.PlatformContext.ARB_create_context_no_error = ExtensionSupportedWGL("WGL_ARB_create_context_no_error")
	mado.GlfwConfig.PlatformContext.EXT_swap_control = ExtensionSupportedWGL("WGL_EXT_swap_control")
	mado.GlfwConfig.PlatformContext.EXT_colorspace = ExtensionSupportedWGL("WGL_EXT_colorspace")
	mado.GlfwConfig.PlatformContext.ARB_pixel_format = ExtensionSupportedWGL("WGL_ARB_pixel_format")
	mado.GlfwConfig.PlatformContext.ARB_context_flush_control = ExtensionSupportedWGL("WGL_ARB_context_flush_control")

	if err := gl.WglMakeCurrent(pdc, prc); err != nil {
		return err
	}
	if err := gl.WglDeleteContext(rc); err != nil {
		return err
	}
	//windows.ReleaseDC(dc)
	mado.GlfwConfig.PlatformContext.Inited = true
	return nil
}

func (c *glContext) createContextWGL(ctxconfig *mado.CtxConfig, fbconfig *mado.FbConfig) (winsyscall.Handle, error) {
	share := EmptyHandle
	// var share _HGLRC
	// if ctxconfig.share != nil {
	// 	share = ctxconfig.share.context.platform.handle
	// }

	pixelFormat, err := c.win.choosePixelFormat(ctxconfig, fbconfig)
	if err != nil {
		return EmptyHandle, err
	}

	var pfd windows.PIXELFORMATDESCRIPTOR
	if _, err := windows.DescribePixelFormat(c.win.hdc, int32(pixelFormat), uint32(unsafe.Sizeof(pfd)), &pfd); err != nil {
		return EmptyHandle, err
	}

	if err := windows.SetPixelFormat(c.win.hdc, int32(pixelFormat), &pfd); err != nil {
		return EmptyHandle, err
	}

	if ctxconfig.Client == mado.OpenGLAPI {
		if ctxconfig.Forward && !mado.GlfwConfig.PlatformContext.ARB_create_context {
			return EmptyHandle, fmt.Errorf("glfw: a forward compatible OpenGL context requested but WGL_ARB_create_context is unavailable: %w", VersionUnavailable)
		}

		if ctxconfig.Profile != 0 && !mado.GlfwConfig.PlatformContext.ARB_create_context_profile {
			return EmptyHandle, fmt.Errorf("glfw: OpenGL profile requested but WGL_ARB_create_context_profile is unavailable: %w", VersionUnavailable)
		}
	} else {
		if !mado.GlfwConfig.PlatformContext.ARB_create_context || !mado.GlfwConfig.PlatformContext.ARB_create_context_profile || !mado.GlfwConfig.PlatformContext.EXT_create_context_es2_profile {
			return EmptyHandle, fmt.Errorf("glfw: OpenGL ES requested but WGL_ARB_create_context_es2_profile is unavailable: %w", APIUnavailable)
		}
	}

	if mado.GlfwConfig.PlatformContext.ARB_create_context {
		var flags int32
		var mask int32
		if ctxconfig.Client == mado.OpenGLAPI {
			if ctxconfig.Forward {
				flags |= _WGL_CONTEXT_FORWARD_COMPATIBLE_BIT_ARB
			}

			if ctxconfig.Profile == mado.OpenGLCoreProfile {
				mask |= _WGL_CONTEXT_CORE_PROFILE_BIT_ARB
			} else if ctxconfig.Profile == mado.OpenGLCompatProfile {
				mask |= _WGL_CONTEXT_COMPATIBILITY_PROFILE_BIT_ARB
			}
		} else {
			mask |= _WGL_CONTEXT_ES2_PROFILE_BIT_EXT
		}

		if ctxconfig.Debug {
			flags |= _WGL_CONTEXT_DEBUG_BIT_ARB
		}

		var attribs []int32
		if ctxconfig.Robustness != 0 {
			if mado.GlfwConfig.PlatformContext.ARB_create_context_robustness {
				if ctxconfig.Robustness == mado.NoResetNotification {
					attribs = append(attribs, _WGL_CONTEXT_RESET_NOTIFICATION_STRATEGY_ARB, _WGL_NO_RESET_NOTIFICATION_ARB)
				} else if ctxconfig.Robustness == mado.LoseContextOnReset {
					attribs = append(attribs, _WGL_CONTEXT_RESET_NOTIFICATION_STRATEGY_ARB, _WGL_LOSE_CONTEXT_ON_RESET_ARB)
				}
				flags |= _WGL_CONTEXT_ROBUST_ACCESS_BIT_ARB
			}
		}

		if ctxconfig.Release != 0 {
			if mado.GlfwConfig.PlatformContext.ARB_context_flush_control {
				if ctxconfig.Release == mado.ReleaseBehaviorNone {
					attribs = append(attribs, _WGL_CONTEXT_RELEASE_BEHAVIOR_ARB, _WGL_CONTEXT_RELEASE_BEHAVIOR_NONE_ARB)
				} else if ctxconfig.Release == mado.ReleaseBehaviorFlush {
					attribs = append(attribs, _WGL_CONTEXT_RELEASE_BEHAVIOR_ARB, _WGL_CONTEXT_RELEASE_BEHAVIOR_FLUSH_ARB)
				}
			}
		}

		if ctxconfig.Noerror {
			if mado.GlfwConfig.PlatformContext.ARB_create_context_no_error {
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

		c.context.ExtensionSupported = extensionSupportedWGL
		return gl.WglCreateContextAttribsARB(c.win.hdc, share, &attribs[0])
	} else {
		return gl.WglCreateContext(c.win.hdc)
	}
}

func (c *glContext) refreshContextAttribs(ctxconfig *mado.CtxConfig) (ferr error) {
	const (
		GL_COLOR_BUFFER_BIT                    = 0x00004000
		GL_CONTEXT_COMPATIBILITY_PROFILE_BIT   = 0x00000002
		GL_CONTEXT_CORE_PROFILE_BIT            = 0x00000001
		GL_CONTEXT_FLAG_DEBUG_BIT              = 0x00000002
		GL_CONTEXT_FLAG_FORWARD_COMPATIBLE_BIT = 0x00000001
		GL_CONTEXT_FLAG_NO_ERROR_BIT_KHR       = 0x00000008
		GL_CONTEXT_FLAGS                       = 0x821E
		GL_CONTEXT_PROFILE_MASK                = 0x9126
		GL_CONTEXT_RELEASE_BEHAVIOR            = 0x82FB
		GL_CONTEXT_RELEASE_BEHAVIOR_FLUSH      = 0x82FC
		GL_LOSE_CONTEXT_ON_RESET_ARB           = 0x8252
		GL_NO_RESET_NOTIFICATION_ARB           = 0x8261
		GL_NONE                                = 0
		GL_RESET_NOTIFICATION_STRATEGY_ARB     = 0x8256
		GL_VERSION                             = 0x1F02
	)

	c.context.Source = ctxconfig.Source
	c.context.Client = mado.OpenGLAPI

	if err := c.Lock(); err != nil {
		return err
	}

	getIntegerv := gl.GetProcAddressWGL("glGetIntegerv")
	getString := gl.GetProcAddressWGL("glGetString")
	if getIntegerv == 0 || getString == 0 {
		return fmt.Errorf("glfw: entry point retrieval is broken: %w", PlatformError)
	}

	r, _, _ := syscall.Syscall(getString, 1, uintptr(GL_VERSION), 0, 0)
	version := bytePtrToString((*byte)(unsafe.Pointer(r)))
	if version == "" {
		if ctxconfig.Client == mado.OpenGLAPI {
			return fmt.Errorf("glfw: OpenGL version string retrieval is broken: %w", PlatformError)
		} else {
			return fmt.Errorf("glfw: OpenGL ES version string retrieval is broken: %w", PlatformError)
		}
	}

	for _, prefix := range []string{
		"OpenGL ES-CM ",
		"OpenGL ES-CL ",
		"OpenGL ES "} {
		if strings.HasPrefix(version, prefix) {
			version = version[len(prefix):]
			c.context.Client = mado.OpenGLESAPI
			break
		}
	}

	m := regexp.MustCompile(`^(\d+)(\.(\d+)(\.(\d+))?)?`).FindStringSubmatch(version)
	if m == nil {
		if c.context.Client == mado.OpenGLAPI {
			return fmt.Errorf("glfw: no version found in OpenGL version string: %w", PlatformError)
		} else {
			return fmt.Errorf("glfw: no version found in OpenGL ES version string: %w", PlatformError)
		}
	}
	c.context.Major, _ = strconv.Atoi(m[1])
	c.context.Minor, _ = strconv.Atoi(m[3])
	c.context.Revision, _ = strconv.Atoi(m[5])

	if c.context.Major < ctxconfig.Major || (c.context.Major == ctxconfig.Major && c.context.Minor < ctxconfig.Minor) {
		// The desired OpenGL version is greater than the actual version
		// This only happens if the machine lacks {GLX|WGL}_ARB_create_context
		// /and/ the user has requested an OpenGL version greater than 1.0

		// For API consistency, we emulate the behavior of the
		// {GLX|WGL}_ARB_create_context extension and fail here

		if c.context.Client == mado.OpenGLAPI {
			return fmt.Errorf("glfw: requested OpenGL version %d.%d, got version %d.%d: %w", ctxconfig.Major, ctxconfig.Minor, c.context.Major, c.context.Minor, VersionUnavailable)
		} else {
			return fmt.Errorf("glfw: requested OpenGL ES version %d.%d, got version %d.%d: %w", ctxconfig.Major, ctxconfig.Minor, c.context.Major, c.context.Minor, VersionUnavailable)
		}
	}

	if c.context.Major >= 3 {
		// OpenGL 3.0+ uses a different function for extension string retrieval
		// We cache it here instead of in glfwExtensionSupported mostly to alert
		// users as early as possible that their build may be broken

		glGetStringi := gl.GetProcAddressWGL("glGetStringi")
		if glGetStringi == 0 {
			return fmt.Errorf("glfw: entry point retrieval is broken: %w", PlatformError)
		}
	}

	if c.context.Client == mado.OpenGLAPI {
		// Read back context flags (OpenGL 3.0 and above)
		if c.context.Major >= 3 {
			var flags int32
			_, _, _ = syscall.Syscall(getIntegerv, GL_CONTEXT_FLAGS, uintptr(unsafe.Pointer(&flags)), 0, 0)

			if flags&GL_CONTEXT_FLAG_FORWARD_COMPATIBLE_BIT != 0 {
				c.context.Forward = true
			}

			if flags&GL_CONTEXT_FLAG_DEBUG_BIT != 0 {
				c.context.Debug = true
			} else {
				ok, err := c.ExtensionSupported("GL_ARB_debug_output")
				if err != nil {
					return err
				}
				if ok && ctxconfig.Debug {
					// HACK: This is a workaround for older drivers (pre KHR_debug)
					//       not setting the debug bit in the context flags for
					//       debug contexts
					c.context.Debug = true
				}
			}

			if flags&GL_CONTEXT_FLAG_NO_ERROR_BIT_KHR != 0 {
				c.context.Noerror = true
			}
		}

		// Read back OpenGL context profile (OpenGL 3.2 and above)
		if c.context.Major >= 4 || (c.context.Major == 3 && c.context.Minor >= 2) {
			var mask int32
			_, _, _ = syscall.Syscall(getIntegerv, GL_CONTEXT_PROFILE_MASK, uintptr(unsafe.Pointer(&mask)), 0, 0)

			if mask&GL_CONTEXT_COMPATIBILITY_PROFILE_BIT != 0 {
				c.context.Profile = mado.OpenGLCompatProfile
			} else if mask&GL_CONTEXT_CORE_PROFILE_BIT != 0 {
				c.context.Profile = mado.OpenGLCoreProfile
			} else {
				ok, err := c.ExtensionSupported("GL_ARB_compatibility")
				if err != nil {
					return err
				}
				if ok {
					// HACK: This is a workaround for the compatibility profile bit
					//       not being set in the context flags if an OpenGL 3.2+
					//       context was created without having requested a specific
					//       version
					c.context.Profile = mado.OpenGLCompatProfile
				}
			}
		}

		// Read back robustness strategy
		ok, err := c.ExtensionSupported("GL_ARB_robustness")
		if err != nil {
			return err
		}
		if ok {
			// NOTE: We avoid using the context flags for detection, as they are
			//       only present from 3.0 while the extension applies from 1.1

			var strategy int32
			_, _, _ = syscall.Syscall(getIntegerv, GL_RESET_NOTIFICATION_STRATEGY_ARB, uintptr(unsafe.Pointer(&strategy)), 0, 0)

			if strategy == GL_LOSE_CONTEXT_ON_RESET_ARB {
				c.context.Robustness = mado.LoseContextOnReset
			} else if strategy == GL_NO_RESET_NOTIFICATION_ARB {
				c.context.Robustness = mado.NoResetNotification
			}
		}
	} else {
		// Read back robustness strategy
		ok, err := c.ExtensionSupported("GL_EXT_robustness")
		if err != nil {
			return err
		}
		if ok {
			// NOTE: The values of these constants match those of the OpenGL ARB
			//       one, so we can reuse them here

			var strategy int32
			_, _, _ = syscall.Syscall(getIntegerv, GL_RESET_NOTIFICATION_STRATEGY_ARB, uintptr(unsafe.Pointer(&strategy)), 0, 0)

			if strategy == GL_LOSE_CONTEXT_ON_RESET_ARB {
				c.context.Robustness = mado.LoseContextOnReset
			} else if strategy == GL_NO_RESET_NOTIFICATION_ARB {
				c.context.Robustness = mado.NoResetNotification
			}
		}
	}

	ok, err := c.ExtensionSupported("GL_KHR_context_flush_control")
	if err != nil {
		return err
	}
	if ok {
		var behavior int32
		_, _, _ = syscall.Syscall(getIntegerv, GL_CONTEXT_RELEASE_BEHAVIOR, uintptr(unsafe.Pointer(&behavior)), 0, 0)

		if behavior == GL_NONE {
			c.context.Release = mado.ReleaseBehaviorNone
		} else if behavior == GL_CONTEXT_RELEASE_BEHAVIOR_FLUSH {
			c.context.Release = mado.ReleaseBehaviorFlush
		}
	}

	// Clearing the front buffer to black to avoid garbage pixels left over from
	// previous uses of our bit of VRAM
	glClear := gl.GetProcAddressWGL("glClear")
	_, _, _ = syscall.Syscall(glClear, GL_COLOR_BUFFER_BIT, 0, 0, 0)

	c.Unlock()
	return nil
}

func (w *window) choosePixelFormat(ctxconfig *mado.CtxConfig, fbconfig_ *mado.FbConfig) (int, error) {
	var nativeCount int32
	var attribs []int32

	if mado.GlfwConfig.PlatformContext.ARB_pixel_format {
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

		if mado.GlfwConfig.PlatformContext.ARB_multisample {
			attribs = append(attribs, _WGL_SAMPLES_ARB)
		}

		if ctxconfig.Client == mado.OpenGLAPI {
			if mado.GlfwConfig.PlatformContext.ARB_framebuffer_sRGB || mado.GlfwConfig.PlatformContext.EXT_framebuffer_sRGB {
				attribs = append(attribs, _WGL_FRAMEBUFFER_SRGB_CAPABLE_ARB)
			}
		} else {
			if mado.GlfwConfig.PlatformContext.EXT_colorspace {
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

	usableConfigs := make([]*mado.FbConfig, 0, nativeCount)
	for i := int32(0); i < nativeCount; i++ {
		var u mado.FbConfig
		pixelFormat := uintptr(i) + 1

		if mado.GlfwConfig.PlatformContext.ARB_pixel_format {
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

			if (findAttribValue(_WGL_DOUBLE_BUFFER_ARB) != 0) != fbconfig_.Doublebuffer {
				continue
			}

			u.RedBits = int(findAttribValue(_WGL_RED_BITS_ARB))
			u.GreenBits = int(findAttribValue(_WGL_GREEN_BITS_ARB))
			u.BlueBits = int(findAttribValue(_WGL_BLUE_BITS_ARB))
			u.AlphaBits = int(findAttribValue(_WGL_ALPHA_BITS_ARB))

			u.DepthBits = int(findAttribValue(_WGL_DEPTH_BITS_ARB))
			u.StencilBits = int(findAttribValue(_WGL_STENCIL_BITS_ARB))

			u.AccumRedBits = int(findAttribValue(_WGL_ACCUM_RED_BITS_ARB))
			u.AccumGreenBits = int(findAttribValue(_WGL_ACCUM_GREEN_BITS_ARB))
			u.AccumBlueBits = int(findAttribValue(_WGL_ACCUM_BLUE_BITS_ARB))
			u.AccumAlphaBits = int(findAttribValue(_WGL_ACCUM_ALPHA_BITS_ARB))

			u.AuxBuffers = int(findAttribValue(_WGL_AUX_BUFFERS_ARB))

			if findAttribValue(_WGL_STEREO_ARB) != 0 {
				u.Stereo = true
			}

			if mado.GlfwConfig.PlatformContext.ARB_multisample {
				u.Samples = int(findAttribValue(_WGL_SAMPLES_ARB))
			}

			if ctxconfig.Client == mado.OpenGLAPI {
				if mado.GlfwConfig.PlatformContext.ARB_framebuffer_sRGB || mado.GlfwConfig.PlatformContext.EXT_framebuffer_sRGB {
					if findAttribValue(_WGL_FRAMEBUFFER_SRGB_CAPABLE_ARB) != 0 {
						u.SRGB = true
					}
				}
			} else {
				if mado.GlfwConfig.PlatformContext.EXT_colorspace {
					if findAttribValue(_WGL_COLORSPACE_EXT) == _WGL_COLORSPACE_SRGB_EXT {
						u.SRGB = true
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

			if (pfd.Flags&_PFD_DOUBLEBUFFER != 0) != fbconfig_.Doublebuffer {
				continue
			}

			u.RedBits = int(pfd.RedBits)
			u.GreenBits = int(pfd.GreenBits)
			u.BlueBits = int(pfd.BlueBits)
			u.AlphaBits = int(pfd.AlphaBits)

			u.DepthBits = int(pfd.DepthBits)
			u.StencilBits = int(pfd.StencilBits)

			u.AccumRedBits = int(pfd.AccumRedBits)
			u.AccumGreenBits = int(pfd.AccumGreenBits)
			u.AccumBlueBits = int(pfd.AccumBlueBits)
			u.AccumAlphaBits = int(pfd.AccumAlphaBits)

			u.AuxBuffers = int(pfd.AuxBuffers)

			if pfd.Flags&_PFD_STEREO != 0 {
				u.Stereo = true
			}
		}

		u.Handle = pixelFormat
		usableConfigs = append(usableConfigs, &u)
	}

	if len(usableConfigs) == 0 {
		return 0, fmt.Errorf("glfw: the driver does not appear to support OpenGL")
	}

	closest := chooseFBConfig(fbconfig_, usableConfigs)
	if closest == nil {
		return 0, fmt.Errorf("glfw: failed to find a suitable pixel format")
	}

	return int(closest.Handle), nil
}

func createHelperWindow() error {
	if resources.helperHwnd != 0 {
		return nil
	}
	h, err := windows.CreateWindowEx(
		windows.WS_EX_OVERLAPPEDWINDOW,
		resources.class,
		"Mado message window",
		windows.WS_CLIPSIBLINGS|windows.WS_CLIPCHILDREN,
		0, 0, 1, 1, 0, 0,
		resources.handle, 0)
	if err != nil {
		return err
	}

	resources.helperHwnd = h

	// HACK: The command to the first ShowWindow call is ignored if the parent
	//       process passed along a STARTUPINFO, so clear that with a no-op call
	windows.ShowWindow(resources.helperHwnd, windows.SW_HIDE)

	// // Register for HID device notifications
	// if !microsoftgdk.IsXbox() {
	// 	_GUID_DEVINTERFACE_HID := windows.GUID{
	// 		Data1: 0x4d1e55b2,
	// 		Data2: 0xf16f,
	// 		Data3: 0x11cf,
	// 		Data4: [...]byte{0x88, 0xcb, 0x00, 0x11, 0x11, 0x00, 0x00, 0x30},
	// 	}

	// 	var dbi _DEV_BROADCAST_DEVICEINTERFACE_W
	// 	dbi.dbcc_size = uint32(unsafe.Sizeof(dbi))
	// 	dbi.dbcc_devicetype = _DBT_DEVTYP_DEVICEINTERFACE
	// 	dbi.dbcc_classguid = _GUID_DEVINTERFACE_HID
	// 	notify, err := _RegisterDeviceNotificationW(windows.Handle(_glfw.platformWindow.helperWindowHandle), unsafe.Pointer(&dbi), _DEVICE_NOTIFY_WINDOW_HANDLE)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	_glfw.platformWindow.deviceNotificationHandle = notify
	// }

	var msg windows.Msg
	for windows.PeekMessage(&msg, resources.helperHwnd, 0, 0, _PM_REMOVE) {
		windows.TranslateMessage(&msg)
		windows.DispatchMessage(&msg)
	}

	return nil
}

func ExtensionSupportedWGL(extension string) bool {
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

func checkValidContextConfig(ctxconfig *mado.CtxConfig) error {
	// if ctxconfig.share != nil {
	// 	if ctxconfig.Client == NoAPI || ctxconfig.share.context.client == NoAPI {
	// 		return NoWindowContext
	// 	}
	// }

	if ctxconfig.Source != mado.NativeContextAPI &&
		ctxconfig.Source != mado.EGLContextAPI &&
		ctxconfig.Source != mado.OSMesaContextAPI {
		return fmt.Errorf("glfw: invalid context creation API 0x%08X: %w", ctxconfig.Source, InvalidEnum)
	}

	if ctxconfig.Client != mado.NoAPI &&
		ctxconfig.Client != mado.OpenGLAPI &&
		ctxconfig.Client != mado.OpenGLESAPI {
		return fmt.Errorf("glfw: invalid client API 0x%08X: %w", ctxconfig.Client, InvalidEnum)
	}

	if ctxconfig.Client == mado.OpenGLAPI {
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

		if ctxconfig.Profile != 0 {
			if ctxconfig.Profile != mado.OpenGLCoreProfile && ctxconfig.Profile != mado.OpenGLCompatProfile {
				return fmt.Errorf("glfw: invalid OpenGL profile 0x%08X: %w", ctxconfig.Profile, InvalidEnum)
			}

			if ctxconfig.Major <= 2 || (ctxconfig.Major == 3 && ctxconfig.Minor < 2) {
				// Desktop OpenGL context profiles are only defined for version 3.2
				// and above

				return fmt.Errorf("glfw: context profiles are only defined for OpenGL version 3.2 and above: %w", InvalidValue)
			}
		}

		if ctxconfig.Forward && ctxconfig.Major <= 2 {
			// Forward-compatible contexts are only defined for OpenGL version 3.0 and above
			return fmt.Errorf("glfw: forward-compatibility is only defined for OpenGL version 3.0 and above: %w", InvalidValue)
		}
	} else if ctxconfig.Client == mado.OpenGLESAPI {
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

	if ctxconfig.Robustness != 0 {
		if ctxconfig.Robustness != mado.NoResetNotification && ctxconfig.Robustness != mado.LoseContextOnReset {
			return fmt.Errorf("glfw: invalid context robustness mode 0x%08X: %w", ctxconfig.Robustness, InvalidEnum)
		}
	}

	if ctxconfig.Release != 0 {
		if ctxconfig.Release != mado.ReleaseBehaviorNone && ctxconfig.Release != mado.ReleaseBehaviorFlush {
			return fmt.Errorf("glfw: invalid context release behavior 0x%08X: %w", ctxconfig.Release, InvalidEnum)
		}
	}

	return nil
}

func chooseFBConfig(desired *mado.FbConfig, alternatives []*mado.FbConfig) *mado.FbConfig {
	leastMissing := math.MaxInt32
	leastColorDiff := math.MaxInt32
	leastExtraDiff := math.MaxInt32

	var closest *mado.FbConfig
	for _, current := range alternatives {
		if desired.Stereo && !current.Stereo {
			// Stereo is a hard constraint
			continue
		}

		// Count number of missing buffers
		missing := 0

		if desired.AlphaBits > 0 && current.AlphaBits == 0 {
			missing++
		}

		if desired.DepthBits > 0 && current.DepthBits == 0 {
			missing++
		}

		if desired.StencilBits > 0 && current.StencilBits == 0 {
			missing++
		}

		if desired.AuxBuffers > 0 &&
			current.AuxBuffers < desired.AuxBuffers {
			missing += desired.AuxBuffers - current.AuxBuffers
		}

		if desired.Samples > 0 && current.Samples == 0 {
			// Technically, several multisampling buffers could be
			// involved, but that's a lower level implementation detail and
			// not important to us here, so we count them as one
			missing++
		}

		if desired.Transparent != current.Transparent {
			missing++
		}

		// These polynomials make many small channel size differences matter
		// less than one large channel size difference

		// Calculate color channel size difference value
		colorDiff := 0

		if desired.RedBits != mado.DontCare {
			colorDiff += (desired.RedBits - current.RedBits) *
				(desired.RedBits - current.RedBits)
		}

		if desired.GreenBits != mado.DontCare {
			colorDiff += (desired.GreenBits - current.GreenBits) *
				(desired.GreenBits - current.GreenBits)
		}

		if desired.BlueBits != mado.DontCare {
			colorDiff += (desired.BlueBits - current.BlueBits) *
				(desired.BlueBits - current.BlueBits)
		}

		// Calculate non-color channel size difference value
		extraDiff := 0

		if desired.AlphaBits != mado.DontCare {
			extraDiff += (desired.AlphaBits - current.AlphaBits) *
				(desired.AlphaBits - current.AlphaBits)
		}

		if desired.DepthBits != mado.DontCare {
			extraDiff += (desired.DepthBits - current.DepthBits) *
				(desired.DepthBits - current.DepthBits)
		}

		if desired.StencilBits != mado.DontCare {
			extraDiff += (desired.StencilBits - current.StencilBits) *
				(desired.StencilBits - current.StencilBits)
		}

		if desired.AccumRedBits != mado.DontCare {
			extraDiff += (desired.AccumRedBits - current.AccumRedBits) *
				(desired.AccumRedBits - current.AccumRedBits)
		}

		if desired.AccumGreenBits != mado.DontCare {
			extraDiff += (desired.AccumGreenBits - current.AccumGreenBits) *
				(desired.AccumGreenBits - current.AccumGreenBits)
		}

		if desired.AccumBlueBits != mado.DontCare {
			extraDiff += (desired.AccumBlueBits - current.AccumBlueBits) *
				(desired.AccumBlueBits - current.AccumBlueBits)
		}

		if desired.AccumAlphaBits != mado.DontCare {
			extraDiff += (desired.AccumAlphaBits - current.AccumAlphaBits) *
				(desired.AccumAlphaBits - current.AccumAlphaBits)
		}

		if desired.Samples != mado.DontCare {
			extraDiff += (desired.Samples - current.Samples) *
				(desired.Samples - current.Samples)
		}

		if desired.SRGB && !current.SRGB {
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

func (c *glContext) ExtensionSupported(extension string) (bool, error) {
	const (
		GL_EXTENSIONS     = 0x1F03
		GL_NUM_EXTENSIONS = 0x821D
	)

	if !mado.GlfwConfig.Initialized {
		return false, NotInitialized
	}

	// ptr, err := _glfc.contextSlot.get()
	// if err != nil {
	// 	return false, err
	// }
	// window := (*Window)(unsafe.Pointer(ptr))
	// if window == nil {
	// 	return false, fmt.Errorf("glfw: cannot query extension without a current OpenGL or OpenGL ES context %w", NoCurrentContext)
	// }

	if c.context.Major >= 3 {
		// Check if extension is in the modern OpenGL extensions string list

		glGetIntegerv := gl.GetProcAddressWGL("glGetIntegerv")
		var count int32
		_, _, _ = syscall.Syscall(glGetIntegerv, GL_NUM_EXTENSIONS, uintptr(unsafe.Pointer(&count)), 0, 0)

		glGetStringi := gl.GetProcAddressWGL("glGetStringi")
		for i := 0; i < int(count); i++ {
			r, _, _ := syscall.Syscall(glGetStringi, GL_EXTENSIONS, uintptr(i), 0, 0)
			if r == 0 {
				return false, fmt.Errorf("glfw: extension string retrieval is broken: %w", PlatformError)
			}

			en := bytePtrToString((*byte)(unsafe.Pointer(r)))
			if en == extension {
				return true, nil
			}
		}
	} else {
		// Check if extension is in the old style OpenGL extensions string

		glGetString := gl.GetProcAddressWGL("glGetString")
		r, _, _ := syscall.Syscall(glGetString, GL_EXTENSIONS, 0, 0, 0)
		if r == 0 {
			return false, fmt.Errorf("glfw: extension string retrieval is broken: %w", PlatformError)
		}

		extensions := bytePtrToString((*byte)(unsafe.Pointer(r)))
		for _, str := range strings.Split(extensions, " ") {
			if str == extension {
				return true, nil
			}
		}
	}

	// Check if extension is in the platform-specific string
	return c.context.ExtensionSupported(extension), nil
}

// bytePtrToString takes a pointer to a sequence of text and returns the corresponding string.
// If the pointer is nil, it returns the empty string. It assumes that the text sequence is
// terminated at a zero byte; if the zero byte is not present, the program may crash.
// It is copied from golang.org/x/sys/windows/winsyscall.go for use on macOS, Linux and Windows
func bytePtrToString(p *byte) string {
	if p == nil {
		return ""
	}
	if *p == 0 {
		return ""
	}

	// Find NUL terminator.
	n := 0
	for ptr := unsafe.Pointer(p); *(*byte)(ptr) != 0; n++ {
		ptr = unsafe.Add(ptr, 1)
	}

	// unsafe.String(p, n) is available as of Go 1.20.
	return string(unsafe.Slice(p, n))
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
