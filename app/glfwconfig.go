// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2002-2006 Marcus Geelnard
// SPDX-FileCopyrightText: 2006-2019 Camilla Löwy <elmindreda@glfw.org>
// SPDX-FileCopyrightText: 2022 The Ebitengine Authors

package app

const (
	_GLFW_INSERT_FIRST = 0
	_GLFW_INSERT_LAST  = 1
)
const (
	True     int = 1 // GL_TRUE
	False    int = 0 // GL_FALSE
	DontCare int = -1
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
}
