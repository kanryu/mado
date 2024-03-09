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
	width            int
	height           int
	title            string
	resizable        bool
	visible          bool
	decorated        bool
	focused          bool
	autoIconify      bool
	floating         bool
	maximized        bool
	centerCursor     bool
	focusOnShow      bool
	mousePassthrough bool
	scaleToMonitor   bool
}

type CtxConfig struct {
	Client     int
	source     int
	Major      int
	Minor      int
	forward    bool
	debug      bool
	noerror    bool
	profile    int
	robustness int
	release    int
	share      *Window
	nsgl       struct {
		offline bool
	}
}

type FbConfig struct {
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
	handle         uintptr
}

type context struct {
	client     int
	source     int
	major      int
	minor      int
	revision   int
	forward    bool
	debug      bool
	noerror    bool
	profile    int
	robustness int
	release    int

	// TODO: Put these functions in an interface type.
	makeCurrent        func(*Window) error
	swapBuffers        func(*Window) error
	swapInterval       func(int) error
	extensionSupported func(string) bool
	getProcAddress     func(string) uintptr
	destroy            func(*Window) error

	platform PlatformContextState
}

type Library struct {
	Initialized bool
	Enable      bool

	Hints struct {
		init        initconfig
		framebuffer FbConfig
		window      WndConfig
		Context     CtxConfig
		refreshRate int
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
	GlfwConfig.Hints.Context.Major = 2
	GlfwConfig.Hints.Context.Minor = 0
	GlfwConfig.Hints.Context.Client = GLFW_OPENGL_API
	GlfwConfig.Hints.Context.source = NativeContextAPI

	GlfwConfig.Hints.framebuffer.redBits = 8
	GlfwConfig.Hints.framebuffer.greenBits = 8
	GlfwConfig.Hints.framebuffer.blueBits = 8
	GlfwConfig.Hints.framebuffer.alphaBits = 8
	GlfwConfig.Hints.framebuffer.depthBits = 24
	GlfwConfig.Hints.framebuffer.stencilBits = 8
	GlfwConfig.Hints.framebuffer.doublebuffer = GLFW_TRUE
}
