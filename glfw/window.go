package glfw

import (
	"fmt"
	"image"
	"io"
	"runtime"
	"strings"
	"sync"
	"time"
	"unsafe"

	"github.com/kanryu/mado"
	"github.com/kanryu/mado/app"
	"github.com/kanryu/mado/io/clipboard"
	"github.com/kanryu/mado/io/key"
	"github.com/kanryu/mado/io/system"
	"github.com/kanryu/mado/unit"
)

// Internal window list stuff
type windowList struct {
	l sync.Mutex
	m map[*app.Window]*Window
}

var windows = windowList{m: map[*app.Window]*Window{}}

func (w *windowList) put(wnd *Window) {
	w.l.Lock()
	defer w.l.Unlock()
	w.m[wnd.data] = wnd
}

func (w *windowList) remove(wnd *app.Window) {
	w.l.Lock()
	defer w.l.Unlock()
	delete(w.m, wnd)
}

func (w *windowList) get(wnd *app.Window) *Window {
	w.l.Lock()
	defer w.l.Unlock()
	return w.m[wnd]
}

// Hint corresponds to hints that can be set before creating a window.
//
// Hint also corresponds to the attributes of the window that can be get after
// its creation.
type Hint int

// Init related hints. (Use with glfw.InitHint)
const (
	JoystickHatButtons  Hint = 0x00050001 // Specifies whether to also expose joystick hats as buttons, for compatibility with earlier versions of GLFW that did not have glfwGetJoystickHats.
	CocoaChdirResources Hint = 0x00051001 // Specifies whether to set the current directory to the application to the Contents/Resources subdirectory of the application's bundle, if present.
	CocoaMenubar        Hint = 0x00051002 // Specifies whether to create a basic menu bar, either from a nib or manually, when the first window is created, which is when AppKit is initialized.
	WaylandLibdecor     Hint = 0x00053001 // Wayland specific init hint.
)

// Window related hints/attributes.
const (
	Focused                Hint = 0x00020001 // Specifies whether the window will be given input focus when created. This hint is ignored for full screen and initially hidden windows.
	Iconified              Hint = 0x00020002 // Specifies whether the window will be minimized.
	Resizable              Hint = 0x00020003 // Specifies whether the window will be resizable by the user.
	Visible                Hint = 0x00020004 // Specifies whether the window will be initially visible.
	Decorated              Hint = 0x00020005 // Specifies whether the window will have window decorations such as a border, a close widget, etc.
	AutoIconify            Hint = 0x00020006 // Specifies whether fullscreen windows automatically iconify (and restore the previous video mode) on focus loss.
	Floating               Hint = 0x00020007 // Specifies whether the window will be always-on-top.
	Maximized              Hint = 0x00020008 // Specifies whether the window is maximized.
	CenterCursor           Hint = 0x00020009 // Specifies whether the cursor should be centered over newly created full screen windows. This hint is ignored for windowed mode windows.
	TransparentFramebuffer Hint = 0x0002000A // Specifies whether the framebuffer should be transparent.
	Hovered                Hint = 0x0002000B // Specifies whether the cursor is currently directly over the content area of the window, with no other windows between. See Cursor enter/leave events for details.
	FocusOnShow            Hint = 0x0002000C // Specifies whether the window will be given input focus when glfwShowWindow is called.
)

// Context related hints.
const (
	ClientAPI               Hint = 0x00022001 // Specifies which client API to create the context for. Hard constraint.
	ContextVersionMajor     Hint = 0x00022002 // Specifies the client API version that the created context must be compatible with.
	ContextVersionMinor     Hint = 0x00022003 // Specifies the client API version that the created context must be compatible with.
	ContextRevision         Hint = 0x00022004 // Context client API revision number attribute.
	ContextRobustness       Hint = 0x00022005 // Specifies the robustness strategy to be used by the context.
	OpenGLForwardCompatible Hint = 0x00022006 // Specifies whether the OpenGL context should be forward-compatible. Hard constraint.
	OpenGLDebugContext      Hint = 0x00022007 // Specifies whether to create a debug OpenGL context, which may have additional error and performance issue reporting functionality. If OpenGL ES is requested, this hint is ignored.
	OpenGLProfile           Hint = 0x00022008 // Specifies which OpenGL profile to create the context for. Hard constraint.
	ContextReleaseBehavior  Hint = 0x00022009 // Specifies the release behavior to be used by the context.
	ContextNoError          Hint = 0x0002200A // Context error suppression hint and attribute.
	ContextCreationAPI      Hint = 0x0002200B // Specifies which context creation API to use to create the context.
	ScaleToMonitor          Hint = 0x0002200C // Specified whether the window content area should be resized based on the monitor content scale of any monitor it is placed on. This includes the initial placement when the window is created.
)

// Framebuffer related hints.
const (
	RedBits                Hint = 0x00021001 // Specifies the desired bit depth of the default framebuffer.
	GreenBits              Hint = 0x00021002 // Specifies the desired bit depth of the default framebuffer.
	BlueBits               Hint = 0x00021003 // Specifies the desired bit depth of the default framebuffer.
	AlphaBits              Hint = 0x00021004 // Specifies the desired bit depth of the default framebuffer.
	DepthBits              Hint = 0x00021005 // Specifies the desired bit depth of the default framebuffer.
	StencilBits            Hint = 0x00021006 // Specifies the desired bit depth of the default framebuffer.
	AccumRedBits           Hint = 0x00021007 // Specifies the desired bit depth of the accumulation buffer.
	AccumGreenBits         Hint = 0x00021008 // Specifies the desired bit depth of the accumulation buffer.
	AccumBlueBits          Hint = 0x00021009 // Specifies the desired bit depth of the accumulation buffer.
	AccumAlphaBits         Hint = 0x0002100A // Specifies the desired bit depth of the accumulation buffer.
	AuxBuffers             Hint = 0x0002100B // Specifies the desired number of auxiliary buffers.
	Stereo                 Hint = 0x0002100C // Specifies whether to use stereoscopic rendering. Hard constraint.
	Samples                Hint = 0x0002100D // Specifies the desired number of samples to use for multisampling. Zero disables multisampling.
	SRGBCapable            Hint = 0x0002100E // Specifies whether the framebuffer should be sRGB capable.
	RefreshRate            Hint = 0x0002100F // Specifies the desired refresh rate for full screen windows. If set to zero, the highest available refresh rate will be used. This hint is ignored for windowed mode windows.
	DoubleBuffer           Hint = 0x00021010 // Specifies whether the framebuffer should be double buffered. You nearly always want to use double buffering. This is a hard constraint.
	CocoaRetinaFramebuffer Hint = 0x00023001 // Specifies whether to use full resolution framebuffers on Retina displays.
	CocoaGraphicsSwitching Hint = 0x00023003 // Specifies whether to in Automatic Graphics Switching, i.e. to allow the system to choose the integrated GPU for the OpenGL context and move it between GPUs if necessary or whether to force it to always run on the discrete GPU.
)

// Naming related hints. (Use with glfw.WindowHintString)
const (
	CocoaFrameNAME  Hint = 0x00023002 // Specifies the UTF-8 encoded name to use for autosaving the window frame, or if empty disables frame autosaving for the window.
	X11ClassName    Hint = 0x00024001 // Specifies the desired ASCII encoded class parts of the ICCCM WM_CLASS window property.nd instance parts of the ICCCM WM_CLASS window property.
	X11InstanceName Hint = 0x00024002 // Specifies the desired ASCII encoded instance parts of the ICCCM WM_CLASS window property.nd instance parts of the ICCCM WM_CLASS window property.
)

// Values for the ClientAPI hint.
const (
	OpenGLAPI   int = 0x00030001
	OpenGLESAPI int = 0x00030002
	NoAPI       int = 0
)

// Values for ContextCreationAPI hint.
const (
	NativeContextAPI int = 0x00036001
	EGLContextAPI    int = 0x00036002
	OSMesaContextAPI int = 0x00036003
)

// Values for the ContextRobustness hint.
const (
	NoRobustness        int = 0
	NoResetNotification int = 0x00031001
	LoseContextOnReset  int = 0x00031002
)

// Values for ContextReleaseBehavior hint.
const (
	AnyReleaseBehavior   int = 0
	ReleaseBehaviorFlush int = 0x00035001
	ReleaseBehaviorNone  int = 0x00035002
)

// Values for the OpenGLProfile hint.
const (
	OpenGLAnyProfile    int = 0
	OpenGLCoreProfile   int = 0x00032001
	OpenGLCompatProfile int = 0x00032002
)

// Other values.
const (
	True     int = 1 // GL_TRUE
	False    int = 0 // GL_FALSE
	DontCare int = -1
)

// Window represents a window.
type Window struct {
	App       *Application
	data      *app.Window
	callbacks *Callbacks
	pointer   unsafe.Pointer
	ctx       mado.Context

	shouldClose bool

	// Window.
	fPosHolder             func(w *Window, xpos int, ypos int)
	fSizeHolder            func(w *Window, width int, height int)
	fFramebufferSizeHolder func(w *Window, width int, height int)
	fCloseHolder           func(w *Window)
	fMaximizeHolder        func(w *Window, maximized bool)
	fContentScaleHolder    func(w *Window, x float32, y float32)
	fRefreshHolder         func(w *Window)
	fFocusHolder           func(w *Window, focused bool)
	fIconifyHolder         func(w *Window, iconified bool)

	// Input.
	fMouseButtonHolder func(w *Window, button MouseButton, action Action, mod ModifierKey)
	fCursorPosHolder   func(w *Window, xpos float64, ypos float64)
	fCursorEnterHolder func(w *Window, entered bool)
	fScrollHolder      func(w *Window, xoff float64, yoff float64)
	fKeyHolder         func(w *Window, key Key, scancode int, action Action, mods ModifierKey)
	fCharHolder        func(w *Window, char rune)
	fCharModsHolder    func(w *Window, char rune, mods ModifierKey)
	fDropHolder        func(w *Window, names []string)

	// IME Input.
	fPreeditHolder func(
		w *Window,
		preeditCount int,
		preeditString string,
		blockCount int,
		blockSizes string,
		focusedBlock int,
		caret int,
	)
	fImeStatusHolder        func(w *Window)
	fPreeditCandidateHolder func(
		w *Window,
		candidatesCount int,
		selectedIndex int,
		pageStart int,
		pageSize int,
	)
}

// CreateWindow creates a window and its associated context. Most of the options
// controlling how the window and its context should be created are specified
// through Hint.
//
// Successful creation does not change which context is current. Before you can
// use the newly created context, you need to make it current using
// MakeContextCurrent.
//
// Note that the created window and context may differ from what you requested,
// as not all parameters and hints are hard constraints. This includes the size
// of the window, especially for full screen windows. To retrieve the actual
// attributes of the created window and context, use queries like
// Window.GetAttrib and Window.GetSize.
//
// To create the window at a specific position, make it initially invisible using
// the Visible window hint, set its position and then show it.
//
// If a fullscreen window is active, the screensaver is prohibited from starting.
//
// Windows: If the executable has an icon resource named GLFW_ICON, it will be
// set as the icon for the window. If no such icon is present, the IDI_WINLOGO
// icon will be used instead.
//
// Mac OS X: The GLFW window has no icon, as it is not a document window, but the
// dock icon will be the same as the application bundle's icon. Also, the first
// time a window is opened the menu bar is populated with common commands like
// Hide, Quit and About. The (minimal) about dialog uses information from the
// application's bundle. For more information on bundles, see the Bundle
// Programming Guide provided by Apple.
//
// This function may only be called from the main thread.
func CreateWindow(width, height int, title string, monitor *Monitor, share *Window) (*Window, error) {
	options := []mado.Option{
		app.CustomRenderer(true),
		app.Size(unit.Dp(width), unit.Dp(height)),
		app.Title(title),
	}
	c := &Callbacks{
		WindowInitialized: make(chan struct{}),
	}
	w := app.NewWindow(c, options...)
	wnd := &Window{
		App:                    theApp,
		data:                   w,
		callbacks:              c,
		fPosHolder:             func(w *Window, xpos int, ypos int) {},
		fSizeHolder:            func(w *Window, width int, height int) {},
		fFramebufferSizeHolder: func(w *Window, width int, height int) {},
		fCloseHolder:           func(w *Window) {},
		fMaximizeHolder:        func(w *Window, maximized bool) {},
		fContentScaleHolder:    func(w *Window, x float32, y float32) {},
		fRefreshHolder:         func(w *Window) {},
		fFocusHolder:           func(w *Window, focused bool) {},
		fIconifyHolder:         func(w *Window, iconified bool) {},
		fPreeditHolder: func(
			w *Window,
			preeditCount int,
			preeditString string,
			blockCount int,
			blockSizes string,
			focusedBlock int,
			caret int,
		) {
		},
		fImeStatusHolder: func(w *Window) {},
		fPreeditCandidateHolder: func(
			w *Window,
			candidatesCount int,
			selectedIndex int,
			pageStart int,
			pageSize int,
		) {
		},
		fMouseButtonHolder: func(w *Window, button MouseButton, action Action, mod ModifierKey) {},
		fCursorPosHolder:   func(w *Window, xpos float64, ypos float64) {},
		fCursorEnterHolder: func(w *Window, entered bool) {},
		fScrollHolder:      func(w *Window, xoff float64, yoff float64) {},
		fKeyHolder:         func(w *Window, key Key, scancode int, action Action, mods ModifierKey) {},
		fCharHolder:        func(w *Window, char rune) {},
		fCharModsHolder:    func(w *Window, char rune, mods ModifierKey) {},
		fDropHolder:        func(w *Window, names []string) {},
	}
	c.SetGlfwWindow(wnd)
	theApp.appendWindow(wnd)
	for {
		runtime.Gosched()
		t := time.NewTicker(time.Millisecond)
		select {
		case <-t.C:
			PollEvents()
			continue
		case <-wnd.callbacks.WindowInitialized:
			return wnd, nil
		}
	}
}

// Destroy destroys the specified window and its context. On calling this
// function, no further callbacks will be called for that window.
//
// This function may only be called from the main thread.
func (w *Window) Destroy() {
	windows.remove(w.data)
	w.callbacks.Event(mado.DestroyEvent{Err: nil})
	panicError()
}

// ShouldClose reports the value of the close flag of the specified window.
func (w *Window) ShouldClose() bool {
	panicError()
	return w.shouldClose
}

// SetShouldClose sets the value of the close flag of the window. This can be
// used to override the user's attempt to close the window, or to signal that it
// should be closed.
func (w *Window) SetShouldClose(value bool) {
	w.shouldClose = value
	panicError()
}

// SetTitle sets the window title, encoded as UTF-8, of the window.
//
// This function may only be called from the main thread.
func (w *Window) SetTitle(title string) {
	option := app.Title(title)
	w.data.Option(option)
	panicError()
}

// SetIcon sets the icon of the specified window. If passed an array of candidate images,
// those of or closest to the sizes desired by the system are selected. If no images are
// specified, the window reverts to its default icon.
//
// The image is ideally provided in the form of *image.NRGBA.
// The pixels are 32-bit, little-endian, non-premultiplied RGBA, i.e. eight
// bits per channel with the red channel first. They are arranged canonically
// as packed sequential rows, starting from the top-left corner. If the image
// type is not *image.NRGBA, it will be converted to it.
//
// The desired image sizes varies depending on platform and system settings. The selected
// images will be rescaled as needed. Good sizes include 16x16, 32x32 and 48x48.
func (w *Window) SetIcon(images []image.Image) {
	fmt.Println("not implemented")

	panicError()
}

// GetPos returns the position, in screen coordinates, of the upper-left
// corner of the client area of the window.
func (w *Window) GetPos() (x, y int) {
	fmt.Println("not implemented")
	panicError()
	return int(0), int(0)
}

// SetPos sets the position, in screen coordinates, of the upper-left corner
// of the client area of the window.
//
// If it is a full screen window, this function does nothing.
//
// If you wish to set an initial window position you should create a hidden
// window (using Hint and Visible), set its position and then show it.
//
// It is very rarely a good idea to move an already visible window, as it will
// confuse and annoy the user.
//
// The window manager may put limits on what positions are allowed.
//
// This function may only be called from the main thread.
func (w *Window) SetPos(xpos, ypos int) {
	fmt.Println("not implemented")
	w.data.Perform(system.ActionCenter)
	panicError()
}

// GetSize returns the size, in screen coordinates, of the client area of the
// specified window.
func (w *Window) GetSize() (width, height int) {
	fmt.Println("not implemented")
	return 800, 600
}

// SetSize sets the size, in screen coordinates, of the client area of the
// window.
//
// For full screen windows, this function selects and switches to the resolution
// closest to the specified size, without affecting the window's context. As the
// context is unaffected, the bit depths of the framebuffer remain unchanged.
//
// The window manager may put limits on what window sizes are allowed.
//
// This function may only be called from the main thread.
func (w *Window) SetSize(width, height int) {
	fmt.Println("not implemented")
	w.data.Perform(system.ActionCenter)
	panicError()
}

// SetSizeLimits sets the size limits of the client area of the specified window.
// If the window is full screen or not resizable, this function does nothing.
//
// The size limits are applied immediately and may cause the window to be resized.
func (w *Window) SetSizeLimits(minw, minh, maxw, maxh int) {
	fmt.Println("not implemented")
	panicError()
}

// SetAspectRatio sets the required aspect ratio of the client area of the specified window.
// If the window is full screen or not resizable, this function does nothing.
//
// The aspect ratio is specified as a numerator and a denominator and both values must be greater
// than zero. For example, the common 16:9 aspect ratio is specified as 16 and 9, respectively.
//
// If the numerator and denominator is set to glfw.DontCare then the aspect ratio limit is disabled.
//
// The aspect ratio is applied immediately and may cause the window to be resized.
func (w *Window) SetAspectRatio(numer, denom int) {
	fmt.Println("not implemented")
	panicError()
}

// GetFramebufferSize retrieves the size, in pixels, of the framebuffer of the
// specified window.
func (w *Window) GetFramebufferSize() (width, height int) {
	size := w.callbacks.D.GetFrameBufferSize()
	panicError()
	return size.X, size.Y
}

// GetFrameSize retrieves the size, in screen coordinates, of each edge of the frame
// of the specified window. This size includes the title bar, if the window has one.
// The size of the frame may vary depending on the window-related hints used to create it.
//
// Because this function retrieves the size of each window frame edge and not the offset
// along a particular coordinate axis, the retrieved values will always be zero or positive.
func (w *Window) GetFrameSize() (left, top, right, bottom int) {
	fmt.Println("not implemented")
	panicError()
	return 0, 0, 800, 600
}

// GetContentScale function retrieves the content scale for the specified
// window. The content scale is the ratio between the current DPI and the
// platform's default DPI. If you scale all pixel dimensions by this scale then
// your content should appear at an appropriate size. This is especially
// important for text and any UI elements.
//
// This function may only be called from the main thread.
func (w *Window) GetContentScale() (float32, float32) {
	fmt.Println("not implemented")
	return 1.0, 1.0
}

// GetOpacity function returns the opacity of the window, including any
// decorations.
//
// The opacity (or alpha) value is a positive finite number between zero and
// one, where zero is fully transparent and one is fully opaque. If the system
// does not support whole window transparency, this function always returns one.
//
// The initial opacity value for newly created windows is one.
//
// This function may only be called from the main thread.
func (w *Window) GetOpacity() float32 {
	fmt.Println("not implemented")
	return 1.0
}

// SetOpacity function sets the opacity of the window, including any
// decorations. The opacity (or alpha) value is a positive finite number between
// zero and one, where zero is fully transparent and one is fully opaque.
//
// The initial opacity value for newly created windows is one.
//
// A window created with framebuffer transparency may not use whole window
// transparency. The results of doing this are undefined.
//
// This function may only be called from the main thread.
func (w *Window) SetOpacity(opacity float32) {
	fmt.Println("not implemented")
}

// RequestWindowAttention funciton requests user attention to the specified
// window. On platforms where this is not supported, attention is requested to
// the application as a whole.
//
// Once the user has given attention, usually by focusing the window or
// application, the system will end the request automatically.
//
// This function must only be called from the main thread.
func (w *Window) RequestAttention() {
	fmt.Println("not implemented")
}

// Focus brings the specified window to front and sets input focus.
// The window should already be visible and not iconified.
//
// By default, both windowed and full screen mode windows are focused when initially created.
// Set the glfw.Focused to disable this behavior.
//
// Do not use this function to steal focus from other applications unless you are certain that
// is what the user wants. Focus stealing can be extremely disruptive.
func (w *Window) Focus() {
	w.callbacks.Event(key.FocusEvent{Focus: true})
}

// Iconify iconifies/minimizes the window, if it was previously restored. If it
// is a full screen window, the original monitor resolution is restored until the
// window is restored. If the window is already iconified, this function does
// nothing.
//
// This function may only be called from the main thread.
func (w *Window) Iconify() {
	w.data.Perform(system.ActionMinimize)
}

// Maximize maximizes the specified window if it was previously not maximized.
// If the window is already maximized, this function does nothing.
//
// If the specified window is a full screen window, this function does nothing.
func (w *Window) Maximize() {
	w.data.Perform(system.ActionMaximize)
}

// Restore restores the window, if it was previously iconified/minimized. If it
// is a full screen window, the resolution chosen for the window is restored on
// the selected monitor. If the window is already restored, this function does
// nothing.
//
// This function may only be called from the main thread.
func (w *Window) Restore() {
	w.data.Wakeup()
}

// Show makes the window visible, if it was previously hidden. If the window is
// already visible or is in full screen mode, this function does nothing.
//
// This function may only be called from the main thread.
func (w *Window) Show() {
	w.data.Perform(system.ActionRaise)
	panicError()
}

// Hide hides the window, if it was previously visible. If the window is already
// hidden or is in full screen mode, this function does nothing.
//
// This function may only be called from the main thread.
func (w *Window) Hide() {
	fmt.Println("not implemented")
	panicError()
}

// GetMonitor returns the handle of the monitor that the window is in
// fullscreen on.
//
// Returns nil if the window is in windowed mode.
func (w *Window) GetMonitor() *Monitor {
	fmt.Println("not implemented")
	return nil
}

// SetMonitor sets the monitor that the window uses for full screen mode or,
// if the monitor is NULL, makes it windowed mode.
//
// When setting a monitor, this function updates the width, height and refresh
// rate of the desired video mode and switches to the video mode closest to it.
// The window position is ignored when setting a monitor.
//
// When the monitor is NULL, the position, width and height are used to place
// the window client area. The refresh rate is ignored when no monitor is specified.
// If you only wish to update the resolution of a full screen window or the size of
// a windowed mode window, see window.SetSize.
//
// When a window transitions from full screen to windowed mode, this function
// restores any previous window settings such as whether it is decorated, floating,
// resizable, has size or aspect ratio limits, etc..
func (w *Window) SetMonitor(monitor *Monitor, xpos, ypos, width, height, refreshRate int) {
	fmt.Println("not implemented")
	panicError()
}

// GetAttrib returns an attribute of the window. There are many attributes,
// some related to the window and others to its context.
func (w *Window) GetAttrib(attrib Hint) int {
	fmt.Println("not implemented")
	panicError()
	return 0
}

// SetAttrib function sets the value of an attribute of the specified window.
//
// The supported attributes are Decorated, Resizeable, Floating and AutoIconify.
//
// Some of these attributes are ignored for full screen windows. The new value
// will take effect if the window is later made windowed.
//
// Some of these attributes are ignored for windowed mode windows. The new value
// will take effect if the window is later made full screen.
//
// This function may only be called from the main thread.
func (w *Window) SetAttrib(attrib Hint, value int) {
	fmt.Println("not implemented")
}

// SetUserPointer sets the user-defined pointer of the window. The current value
// is retained until the window is destroyed. The initial value is nil.
func (w *Window) SetUserPointer(pointer unsafe.Pointer) {
	w.pointer = pointer
	panicError()
}

// GetUserPointer returns the current value of the user-defined pointer of the
// window. The initial value is nil.
func (w *Window) GetUserPointer() unsafe.Pointer {
	ret := w.pointer
	panicError()
	return ret
}

// PosCallback is the window position callback.
type PosCallback func(w *Window, xpos int, ypos int)

// SetPosCallback sets the position callback of the window, which is called
// when the window is moved. The callback is provided with the screen position
// of the upper-left corner of the client area of the window.
func (w *Window) SetPosCallback(cbfun PosCallback) (previous PosCallback) {
	previous = w.fPosHolder
	w.fPosHolder = cbfun
	panicError()
	return previous
}

// SizeCallback is the window size callback.
type SizeCallback func(w *Window, width int, height int)

// SetSizeCallback sets the size callback of the window, which is called when
// the window is resized. The callback is provided with the size, in screen
// coordinates, of the client area of the window.
func (w *Window) SetSizeCallback(cbfun SizeCallback) (previous SizeCallback) {
	previous = w.fSizeHolder
	w.fSizeHolder = cbfun
	panicError()
	return previous
}

// FramebufferSizeCallback is the framebuffer size callback.
type FramebufferSizeCallback func(w *Window, width int, height int)

// SetFramebufferSizeCallback sets the framebuffer resize callback of the specified
// window, which is called when the framebuffer of the specified window is resized.
func (w *Window) SetFramebufferSizeCallback(cbfun FramebufferSizeCallback) (previous FramebufferSizeCallback) {
	previous = w.fFramebufferSizeHolder
	w.fFramebufferSizeHolder = cbfun
	panicError()
	return previous
}

// CloseCallback is the window close callback.
type CloseCallback func(w *Window)

// SetCloseCallback sets the close callback of the window, which is called when
// the user attempts to close the window, for example by clicking the close
// widget in the title bar.
//
// The close flag is set before this callback is called, but you can modify it at
// any time with SetShouldClose.
//
// Mac OS X: Selecting Quit from the application menu will trigger the close
// callback for all windows.
func (w *Window) SetCloseCallback(cbfun CloseCallback) (previous CloseCallback) {
	previous = w.fCloseHolder
	w.fCloseHolder = cbfun
	panicError()
	return previous
}

// MaximizeCallback is the function signature for window maximize callback
// functions.
type MaximizeCallback func(w *Window, maximized bool)

// SetMaximizeCallback sets the maximization callback of the specified window,
// which is called when the window is maximized or restored.
//
// This function must only be called from the main thread.
func (w *Window) SetMaximizeCallback(cbfun MaximizeCallback) MaximizeCallback {
	previous := w.fMaximizeHolder
	w.fMaximizeHolder = cbfun
	return previous
}

// ContentScaleCallback is the function signature for window content scale
// callback functions.
type ContentScaleCallback func(w *Window, x float32, y float32)

// SetContentScaleCallback function sets the window content scale callback of
// the specified window, which is called when the content scale of the specified
// window changes.
//
// This function must only be called from the main thread.
func (w *Window) SetContentScaleCallback(cbfun ContentScaleCallback) ContentScaleCallback {
	previous := w.fContentScaleHolder
	w.fContentScaleHolder = cbfun
	return previous
}

// RefreshCallback is the window refresh callback.
type RefreshCallback func(w *Window)

// SetRefreshCallback sets the refresh callback of the window, which
// is called when the client area of the window needs to be redrawn, for example
// if the window has been exposed after having been covered by another window.
//
// On compositing window systems such as Aero, Compiz or Aqua, where the window
// contents are saved off-screen, this callback may be called only very
// infrequently or never at all.
func (w *Window) SetRefreshCallback(cbfun RefreshCallback) (previous RefreshCallback) {
	previous = w.fRefreshHolder
	w.fRefreshHolder = cbfun
	panicError()
	return previous
}

// FocusCallback is the window focus callback.
type FocusCallback func(w *Window, focused bool)

// SetFocusCallback sets the focus callback of the window, which is called when
// the window gains or loses focus.
//
// After the focus callback is called for a window that lost focus, synthetic key
// and mouse button release events will be generated for all such that had been
// pressed. For more information, see SetKeyCallback and SetMouseButtonCallback.
func (w *Window) SetFocusCallback(cbfun FocusCallback) (previous FocusCallback) {
	previous = w.fFocusHolder
	w.fFocusHolder = cbfun
	panicError()
	return previous
}

// IconifyCallback is the window iconification callback.
type IconifyCallback func(w *Window, iconified bool)

// SetIconifyCallback sets the iconification callback of the window, which is
// called when the window is iconified or restored.
func (w *Window) SetIconifyCallback(cbfun IconifyCallback) (previous IconifyCallback) {
	previous = w.fIconifyHolder
	w.fIconifyHolder = cbfun
	panicError()
	return previous
}

// MouseButtonCallback is the mouse button callback.
type MouseButtonCallback func(w *Window, button MouseButton, action Action, mods ModifierKey)

// SetMouseButtonCallback sets the mouse button callback which is called when a
// mouse button is pressed or released.
//
// When a window loses focus, it will generate synthetic mouse button release
// events for all pressed mouse buttons. You can tell these events from
// user-generated events by the fact that the synthetic ones are generated after
// the window has lost focus, i.e. Focused will be false and the focus
// callback will have already been called.
func (w *Window) SetMouseButtonCallback(cbfun MouseButtonCallback) (previous MouseButtonCallback) {
	previous = w.fMouseButtonHolder
	w.fMouseButtonHolder = cbfun
	panicError()
	return previous
}

// CursorPosCallback the cursor position callback.
type CursorPosCallback func(w *Window, xpos float64, ypos float64)

// SetCursorPosCallback sets the cursor position callback which is called
// when the cursor is moved. The callback is provided with the position relative
// to the upper-left corner of the client area of the window.
func (w *Window) SetCursorPosCallback(cbfun CursorPosCallback) (previous CursorPosCallback) {
	previous = w.fCursorPosHolder
	w.fCursorPosHolder = cbfun
	panicError()
	return previous
}

// CursorEnterCallback is the cursor boundary crossing callback.
type CursorEnterCallback func(w *Window, entered bool)

// SetCursorEnterCallback the cursor boundary crossing callback which is called
// when the cursor enters or leaves the client area of the window.
func (w *Window) SetCursorEnterCallback(cbfun CursorEnterCallback) (previous CursorEnterCallback) {
	previous = w.fCursorEnterHolder
	w.fCursorEnterHolder = cbfun
	panicError()
	return previous
}

// ScrollCallback is the scroll callback.
type ScrollCallback func(w *Window, xoff float64, yoff float64)

// SetScrollCallback sets the scroll callback which is called when a scrolling
// device is used, such as a mouse wheel or scrolling area of a touchpad.
func (w *Window) SetScrollCallback(cbfun ScrollCallback) (previous ScrollCallback) {
	previous = w.fScrollHolder
	w.fScrollHolder = cbfun
	panicError()
	return previous
}

// KeyCallback is the key callback.
type KeyCallback func(w *Window, key Key, scancode int, action Action, mods ModifierKey)

// SetKeyCallback sets the key callback which is called when a key is pressed,
// repeated or released.
//
// The key functions deal with physical keys, with layout independent key tokens
// named after their values in the standard US keyboard layout. If you want to
// input text, use the SetCharCallback instead.
//
// When a window loses focus, it will generate synthetic key release events for
// all pressed keys. You can tell these events from user-generated events by the
// fact that the synthetic ones are generated after the window has lost focus,
// i.e. Focused will be false and the focus callback will have already been
// called.
func (w *Window) SetKeyCallback(cbfun KeyCallback) (previous KeyCallback) {
	previous = w.fKeyHolder
	w.fKeyHolder = cbfun
	panicError()
	return previous
}

// CharCallback is the character callback.
type CharCallback func(w *Window, char rune)

// SetCharCallback sets the character callback which is called when a
// Unicode character is input.
//
// The character callback is intended for Unicode text input. As it deals with
// characters, it is keyboard layout dependent, whereas the
// key callback is not. Characters do not map 1:1
// to physical keys, as a key may produce zero, one or more characters. If you
// want to know whether a specific physical key was pressed or released, see
// the key callback instead.
//
// The character callback behaves as system text input normally does and will
// not be called if modifier keys are held down that would prevent normal text
// input on that platform, for example a Super (Command) key on OS X or Alt key
// on Windows. There is a character with modifiers callback that receives these events.
func (w *Window) SetCharCallback(cbfun CharCallback) (previous CharCallback) {
	previous = w.fCharHolder
	w.fCharHolder = cbfun
	panicError()
	return previous
}

// CharModsCallback is the character with modifiers callback.
type CharModsCallback func(w *Window, char rune, mods ModifierKey)

// SetCharModsCallback sets the character with modifiers callback which is called when a
// Unicode character is input regardless of what modifier keys are used.
//
// Deprecated: Scheduled for removal in version 4.0.
//
// The character with modifiers callback is intended for implementing custom
// Unicode character input. For regular Unicode text input, see the
// character callback. Like the character callback, the character with modifiers callback
// deals with characters and is keyboard layout dependent. Characters do not
// map 1:1 to physical keys, as a key may produce zero, one or more characters.
// If you want to know whether a specific physical key was pressed or released,
// see the key callback instead.
func (w *Window) SetCharModsCallback(cbfun CharModsCallback) (previous CharModsCallback) {
	previous = w.fCharModsHolder
	w.fCharModsHolder = cbfun
	panicError()
	return previous
}

// DropCallback is the drop callback.
type DropCallback func(w *Window, names []string)

// SetDropCallback sets the drop callback which is called when an object
// is dropped over the window.
func (w *Window) SetDropCallback(cbfun DropCallback) (previous DropCallback) {
	previous = w.fDropHolder
	w.fDropHolder = cbfun
	panicError()
	return previous
}

// PreeditCallback is preedit text input callback.
type PreeditCallback func(
	w *Window,
	preeditCount int,
	preeditString string,
	blockCount int,
	blockSizes string,
	focusedBlock int,
	caret int,
)

// SetPreeditCallback sets the preedit text input callback to the window.
//
// IME Users enter text with the IME turned on. At this time,
// no char input event occurs in the Window, and the window is notified
// of the character string of the undefined input token(called preedit).
// The window must display this token appropriately.
func (w *Window) SetPreeditCallback(cbfun PreeditCallback) (previous PreeditCallback) {
	previous = w.fPreeditHolder
	w.fPreeditHolder = cbfun
	panicError()
	return previous
}

// ImeStatusCallback is change signal of the Ime status to the window callback.
type ImeStatusCallback func(w *Window)

// SetImeStatusCallback sets the callback of change signal of the Ime status to the window.
//
// Users of languages that require text input using an IME should turn on the IME before entering text.
// This callback receives IME ON/OFF events.
func (w *Window) SetImeStatusCallback(cbfun ImeStatusCallback) (previous ImeStatusCallback) {
	previous = w.fImeStatusHolder
	w.fImeStatusHolder = cbfun
	panicError()
	return previous
}

// PreeditCandidateCallback is change signal of the Ime status to the window callback.
type PreeditCandidateCallback func(
	w *Window,
	candidatesCount int,
	selectedIndex int,
	pageStart int,
	pageSize int,
)

// SetPreeditCandidateCallback sets a callback that receives
// the information necessary to display a window of the list of conversion candidate strings for preedit.
//
// When an IME user enters preedit, they typically select from a list of conversion candidate tokens.
// Many OS have a dedicated pull-down display, and you can switch between conversion candidates
// by using the space, tab, up, down cursor keys.
// There may be too many selection candidates to display in the pull-down list.
func (w *Window) SetPreeditCandidateCallback(cbfun PreeditCandidateCallback) (previous PreeditCandidateCallback) {
	previous = w.fPreeditCandidateHolder
	w.fPreeditCandidateHolder = cbfun
	panicError()
	return previous
}

// SetClipboardString sets the system clipboard to the specified UTF-8 encoded
// string.
//
// Ownership to the Window is no longer necessary, see
// glfw.SetClipboardString(string)
//
// This function may only be called from the main thread.
func (w *Window) SetClipboardString(str string) {
	gtx := w.data.Queue.Source()
	gtx.Execute(clipboard.WriteCmd{Type: "application/text", Data: io.NopCloser(strings.NewReader(str))})
	panicError()
}

// GetClipboardString returns the contents of the system clipboard, if it
// contains or is convertible to a UTF-8 encoded string.
//
// Ownership to the Window is no longer necessary, see
// glfw.GetClipboardString()
//
// This function may only be called from the main thread.
func (w *Window) GetClipboardString() string {
	gtx := w.data.Queue.Source()
	gtx.Execute(clipboard.ReadCmd{Tag: w})
	// There is probably no way to retrieve the clipboard synchronously due to the current geoui design.
	fmt.Println("not implemented")
	return ""
}
