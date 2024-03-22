// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2002-2006 Marcus Geelnard
// SPDX-FileCopyrightText: 2006-2019 Camilla Löwy <elmindreda@glfw.org>
// SPDX-FileCopyrightText: 2022 The Ebitengine Authors

package mado

import (
	"fmt"
	"math"
	"runtime"
	"slices"
	"strings"

	"github.com/kanryu/mado/internal/gl"
)

const (
	_GLFW_INSERT_FIRST = 0
	_GLFW_INSERT_LAST  = 1

	GLFW_TRUE  = true
	GLFW_FALSE = false

	GLFW_DONT_CARE              = -1
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
)
const (
	True     int = 1 // GL_TRUE
	False    int = 0 // GL_FALSE
	DontCare int = -1

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

	WindowTypeWindows = 0x00040001
	WindowTypeMac     = 0x00040002
	WindowTypeX11     = 0x00040003
	WindowTypeWayland = 0x00040004
)

var GlfwConfig Library

type initconfig struct {
	hatButtons bool
}

type WndConfig struct {
	Width            int
	Height           int
	Title            string
	Resizable        bool
	Visible          bool
	Decorated        bool
	Focused          bool
	AutoIconify      bool
	Floating         bool
	Maximized        bool
	CenterCursor     bool
	FocusOnShow      bool
	MousePassthrough bool
	ScaleToMonitor   bool
	Ns               struct {
		Retina    bool
		FrameName string
	}
	X11 struct {
		ClassName    string
		InstanceName string
	}
}

type CtxConfig struct {
	Client     int
	Source     int
	Major      int
	Minor      int
	Forward    bool
	Debug      bool
	Noerror    bool
	Profile    int
	Robustness int
	Release    int
	share      *Window
	Nsgl       struct {
		Offline bool
	}
}

type FbConfig struct {
	RedBits        int
	GreenBits      int
	BlueBits       int
	AlphaBits      int
	DepthBits      int
	StencilBits    int
	AccumRedBits   int
	AccumGreenBits int
	AccumBlueBits  int
	AccumAlphaBits int
	AuxBuffers     int
	Stereo         bool
	Samples        int
	SRGB           bool
	Doublebuffer   bool
	Transparent    bool
	Handle         uintptr
}

type GlfwContext struct {
	Client     int
	Source     int
	Major      int
	Minor      int
	Revision   int
	Forward    bool
	Debug      bool
	Noerror    bool
	Profile    int
	Robustness int
	Release    int

	// TODO: Put these functions in an interface type.
	MakeCurrent        func(*Window) error
	SwapBuffers        func(*Window) error
	SwapInterval       func(int) error
	ExtensionSupported func(string) bool
	GetProcAddress     func(string) uintptr
	Destroy            func(*Window) error
	GetString          func(gl.Enum) string
	GetInteger         func(gl.Enum) int

	Platform PlatformContextState
}

func (c *GlfwContext) RefreshContextAttribs(ctxconfig *CtxConfig) error {
	exts := strings.Split(c.GetString(gl.EXTENSIONS), " ")
	glVer := c.GetString(gl.VERSION)

	prefixes := []string{
		"OpenGL ES-CM ",
		"OpenGL ES-CL ",
		"OpenGL ES ",
	}
	for _, pref := range prefixes {
		if glVer[:len(pref)] == pref {
			c.Client = GLFW_OPENGL_ES_API
		}
	}
	ver, _, err := gl.ParseGLVersion(glVer)
	if err != nil {
		return err
	}
	c.Major, c.Minor = ver[0], ver[1]
	if c.Major < ctxconfig.Major ||
		(c.Major == ctxconfig.Major &&
			c.Minor < ctxconfig.Minor) {
		// The desired OpenGL version is greater than the actual version
		// This only happens if the machine lacks {GLX|WGL}_ARB_create_context
		// /and/ the user has requested an OpenGL version greater than 1.0

		// For API consistency, we emulate the behavior of the
		// {GLX|WGL}_ARB_create_context extension and fail here

		if c.Client == GLFW_OPENGL_API {
			return fmt.Errorf("Requested OpenGL version %i.%i, got version %i.%i",
				ctxconfig.Major, ctxconfig.Minor,
				c.Major, c.Minor,
			)
		} else {
			return fmt.Errorf("Requested OpenGL ES version %i.%i, got version %i.%i",
				ctxconfig.Major, ctxconfig.Minor,
				c.Major, c.Minor,
			)
		}
	}
	// Read back context flags (OpenGL 3.0 and above)
	if c.Major >= 3 {
		flags := c.GetInteger(gl.CONTEXT_FLAGS)
		if flags&gl.CONTEXT_FLAG_FORWARD_COMPATIBLE_BIT != 0 {
			c.Forward = true
		}
		if flags&gl.CONTEXT_FLAG_DEBUG_BIT != 0 {
			c.Debug = true
		} else if slices.Contains(exts, "GL_ARB_debug_output") && ctxconfig.Debug {
			// HACK: This is a workaround for older drivers (pre KHR_debug)
			//       not setting the debug bit in the context flags for
			//       debug contexts
			c.Debug = true
		}
		if flags&gl.CONTEXT_FLAG_NO_ERROR_BIT_KHR != 0 {
			c.Noerror = true
		}
	}
	// Read back OpenGL context profile (OpenGL 3.2 and above)
	if c.Major >= 4 ||
		(c.Major == 3 && c.Minor >= 2) {
		mask := c.GetInteger(gl.CONTEXT_PROFILE_MASK)
		if mask&gl.CONTEXT_COMPATIBILITY_PROFILE_BIT != 0 {
			c.Profile = GLFW_OPENGL_COMPAT_PROFILE
		} else if mask&gl.CONTEXT_CORE_PROFILE_BIT != 0 {
			c.Profile = GLFW_OPENGL_CORE_PROFILE
		} else if slices.Contains(exts, "GL_ARB_compatibility") {
			// HACK: This is a workaround for the compatibility profile bit
			//       not being set in the context flags if an OpenGL 3.2+
			//       context was created without having requested a specific
			//       version
			c.Profile = GLFW_OPENGL_COMPAT_PROFILE
		}
	}

	// Read back robustness strategy
	if slices.Contains(exts, "GL_ARB_robustness") {
		// NOTE: We avoid using the context flags for detection, as they are
		//       only present from 3.0 while the extension applies from 1.1
		strategy := c.GetInteger(gl.RESET_NOTIFICATION_STRATEGY_ARB)
		if strategy == gl.LOSE_CONTEXT_ON_RESET_ARB {
			c.Robustness = GLFW_LOSE_CONTEXT_ON_RESET
		} else if strategy == gl.NO_RESET_NOTIFICATION_ARB {
			c.Robustness = GLFW_NO_RESET_NOTIFICATION
		}
	}
	if slices.Contains(exts, "GL_KHR_context_flush_control") {
		behavior := c.GetInteger(gl.CONTEXT_RELEASE_BEHAVIOR)
		if behavior == gl.ZERO {
			c.Release = GLFW_RELEASE_BEHAVIOR_NONE
		} else if behavior == gl.CONTEXT_RELEASE_BEHAVIOR_FLUSH {
			c.Release = GLFW_RELEASE_BEHAVIOR_FLUSH
		}
	}
	return nil
}

type Library struct {
	Initialized bool
	Enable      bool
	WindowType  int

	Hints struct {
		init        initconfig
		Framebuffer FbConfig
		Window      WndConfig
		Context     CtxConfig
		RefreshRate int
	}

	errors []error // TODO: Check the error at polling?
	// cursors []*Cursor
	// windows []*Window

	// monitors []*Monitor

	// contextSlot tls

	// platformWindow  platformLibraryWindowState
	PlatformContext PlatformLibraryContextState
}

func boolToInt(x bool) int {
	if x {
		return 1
	}
	return 0
}

func intToBool(x int) bool {
	return x != 0
}

func GlfwConfigInit() {
	// The default is OpenGL with minimum version 1.0
	GlfwConfig.Hints.Context.Client = GLFW_OPENGL_API
	switch {
	case runtime.GOOS == "darwin", runtime.GOOS == "windows":
		GlfwConfig.Hints.Context.Source = NativeContextAPI
	default:
		GlfwConfig.Hints.Context.Source = EGLContextAPI
	}
	GlfwConfig.Hints.Context.Major = 2
	GlfwConfig.Hints.Context.Minor = 0

	// The default is a focused, visible, resizable window with decorations
	GlfwConfig.Hints.Window.Resizable = true
	GlfwConfig.Hints.Window.Decorated = true
	GlfwConfig.Hints.Window.Focused = true
	GlfwConfig.Hints.Window.AutoIconify = true
	GlfwConfig.Hints.Window.Floating = true
	GlfwConfig.Hints.Window.Maximized = true
	GlfwConfig.Hints.Window.Visible = true

	// The default is 24 bits of color, 24 bits of depth and 8 bits of stencil,
	GlfwConfig.Hints.Framebuffer.RedBits = 8
	GlfwConfig.Hints.Framebuffer.GreenBits = 8
	GlfwConfig.Hints.Framebuffer.BlueBits = 8
	GlfwConfig.Hints.Framebuffer.AlphaBits = 8
	GlfwConfig.Hints.Framebuffer.DepthBits = 24
	GlfwConfig.Hints.Framebuffer.StencilBits = 8
	GlfwConfig.Hints.Framebuffer.Doublebuffer = GLFW_TRUE

	// The default is to select the highest available refresh rate
	GlfwConfig.Hints.RefreshRate = GLFW_DONT_CARE

	// The default is to use full Retina resolution framebuffers
	GlfwConfig.Hints.Window.Ns.Retina = GLFW_TRUE

}

func ChooseFBConfig(desired *FbConfig, alternatives []*FbConfig) *FbConfig {
	leastMissing := math.MaxInt32
	leastColorDiff := math.MaxInt32
	leastExtraDiff := math.MaxInt32

	var closest *FbConfig
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

		if desired.RedBits != DontCare {
			colorDiff += (desired.RedBits - current.RedBits) *
				(desired.RedBits - current.RedBits)
		}

		if desired.GreenBits != DontCare {
			colorDiff += (desired.GreenBits - current.GreenBits) *
				(desired.GreenBits - current.GreenBits)
		}

		if desired.BlueBits != DontCare {
			colorDiff += (desired.BlueBits - current.BlueBits) *
				(desired.BlueBits - current.BlueBits)
		}

		// Calculate non-color channel size difference value
		extraDiff := 0

		if desired.AlphaBits != DontCare {
			extraDiff += (desired.AlphaBits - current.AlphaBits) *
				(desired.AlphaBits - current.AlphaBits)
		}

		if desired.DepthBits != DontCare {
			extraDiff += (desired.DepthBits - current.DepthBits) *
				(desired.DepthBits - current.DepthBits)
		}

		if desired.StencilBits != DontCare {
			extraDiff += (desired.StencilBits - current.StencilBits) *
				(desired.StencilBits - current.StencilBits)
		}

		if desired.AccumRedBits != DontCare {
			extraDiff += (desired.AccumRedBits - current.AccumRedBits) *
				(desired.AccumRedBits - current.AccumRedBits)
		}

		if desired.AccumGreenBits != DontCare {
			extraDiff += (desired.AccumGreenBits - current.AccumGreenBits) *
				(desired.AccumGreenBits - current.AccumGreenBits)
		}

		if desired.AccumBlueBits != DontCare {
			extraDiff += (desired.AccumBlueBits - current.AccumBlueBits) *
				(desired.AccumBlueBits - current.AccumBlueBits)
		}

		if desired.AccumAlphaBits != DontCare {
			extraDiff += (desired.AccumAlphaBits - current.AccumAlphaBits) *
				(desired.AccumAlphaBits - current.AccumAlphaBits)
		}

		if desired.Samples != DontCare {
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
