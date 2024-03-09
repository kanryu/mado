// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2002-2006 Marcus Geelnard
// SPDX-FileCopyrightText: 2006-2019 Camilla Löwy <elmindreda@glfw.org>
// SPDX-FileCopyrightText: 2022 The Ebitengine Authors

package app

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
	Fajor      int
	Minor      int
	Revision   int
	Forward    bool
	Debug      bool
	Noerror    bool
	Profile    int
	Robustness int
	Release    int

	// TODO: Put these functions in an interface type.
	makeCurrent        func(*Window) error
	swapBuffers        func(*Window) error
	swapInterval       func(int) error
	extensionSupported func(string) bool
	getProcAddress     func(string) uintptr
	destroy            func(*Window) error

	Platform PlatformContextState
}

type Library struct {
	Initialized bool
	Enable      bool

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

func glfwconfiginit() {
	// The default is OpenGL with minimum version 1.0
	GlfwConfig.Hints.Context.Client = GLFW_OPENGL_API
	GlfwConfig.Hints.Context.Source = NativeContextAPI
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
