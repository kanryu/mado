package mado

import (
	"errors"
	"image"
	"image/color"

	"github.com/kanryu/mado/gpu"
	"github.com/kanryu/mado/io/key"
	"github.com/kanryu/mado/io/pointer"
	"github.com/kanryu/mado/io/system"
	"github.com/kanryu/mado/unit"
)

// Config describes a Window configuration.
type Config struct {
	// Size is the window dimensions (Width, Height).
	Size image.Point
	// MaxSize is the window maximum allowed dimensions.
	MaxSize image.Point
	// MinSize is the window minimum allowed dimensions.
	MinSize image.Point
	// Title is the window title displayed in its decoration bar.
	Title string
	// WindowMode is the window mode.
	Mode WindowMode
	// StatusColor is the color of the Android status bar.
	StatusColor color.NRGBA
	// NavigationColor is the color of the navigation bar
	// on Android, or the address bar in browsers.
	NavigationColor color.NRGBA
	// Orientation is the current window orientation.
	Orientation Orientation
	// CustomRenderer is true when the window content is rendered by the
	// client.
	CustomRenderer bool
	// Decorated reports whether window decorations are provided automatically.
	Decorated bool
	// decoHeight is the height of the fallback decoration for platforms such
	// as Wayland that may need fallback client-side decorations.
	DecoHeight unit.Dp
}

// ConfigEvent is sent whenever the configuration of a Window changes.
type ConfigEvent struct {
	Config Config
}

func (c *Config) Apply(m unit.Metric, options []Option) {
	for _, o := range options {
		o(m, c)
	}
}

type WakeupEvent struct{}

// A StageEvent is generated whenever the stage of a
// Window changes.
type StageEvent struct {
	Stage      Stage
	WindowMode WindowMode
}

// Stage of a Window.
type Stage uint8

const (
	// StagePaused is the stage for windows that have no on-screen representation.
	// Paused windows don't receive frames.
	StagePaused Stage = iota
	// StageInactive is the stage for windows that are visible, but not active.
	// Inactive windows receive frames.
	StageInactive
	// StageRunning is for active and visible Windows.
	// Running windows receive frames.
	StageRunning
)

// String implements fmt.Stringer.
func (l Stage) String() string {
	switch l {
	case StagePaused:
		return "StagePaused"
	case StageInactive:
		return "StageInactive"
	case StageRunning:
		return "StageRunning"
	default:
		panic("unexpected Stage value")
	}
}

// WindowMode is the window mode (WindowMode.Option sets it).
// Note that mode can be changed programatically as well as by the user
// clicking on the minimize/maximize buttons on the window's title bar.
type WindowMode uint8

const (
	Noop WindowMode = iota
	// Windowed is the normal window mode with OS specific window decorations.
	Windowed
	// Fullscreen is the full screen window mode.
	Fullscreen
	// Minimized is for systems where the window can be minimized to an icon.
	Minimized
	// Maximized is for systems where the window can be made to fill the available monitor area.
	Maximized
)

// Option changes the mode of a Window.
func (m WindowMode) Option() Option {
	return func(_ unit.Metric, cnf *Config) {
		cnf.Mode = m
	}
}

// String returns the mode name.
func (m WindowMode) String() string {
	switch m {
	case Windowed:
		return "windowed"
	case Fullscreen:
		return "fullscreen"
	case Minimized:
		return "minimized"
	case Maximized:
		return "maximized"
	}
	return ""
}

// Orientation is the orientation of the app (Orientation.Option sets it).
//
// Supported platforms are Android and JS.
type Orientation uint8

const (
	// AnyOrientation allows the window to be freely orientated.
	AnyOrientation Orientation = iota
	// LandscapeOrientation constrains the window to landscape orientations.
	LandscapeOrientation
	// PortraitOrientation constrains the window to portrait orientations.
	PortraitOrientation
)

func (o Orientation) Option() Option {
	return func(_ unit.Metric, cnf *Config) {
		cnf.Orientation = o
	}
}

func (o Orientation) String() string {
	switch o {
	case AnyOrientation:
		return "any"
	case LandscapeOrientation:
		return "landscape"
	case PortraitOrientation:
		return "portrait"
	}
	return ""
}

// type frameEvent struct {
// 	mado.FrameEvent

// 	Sync bool
// }

type Context interface {
	API() gpu.API
	RenderTarget() (gpu.RenderTarget, error)
	Present() error
	Refresh() error
	Release()
	Lock() error
	Unlock()
	SwapBuffers()
}

// Driver is the interface for the platform implementation
// of a window.
type Driver interface {
	// SetAnimating sets the animation flag. When the window is animating,
	// FrameEvents are delivered as fast as the display can handle them.
	SetAnimating(anim bool)
	// ShowTextInput updates the virtual keyboard state.
	ShowTextInput(show bool)
	SetInputHint(mode key.InputHint)
	NewContext() (Context, error)
	// ReadClipboard requests the clipboard content.
	ReadClipboard()
	// WriteClipboard requests a clipboard write.
	WriteClipboard(mime string, s []byte)
	// Configure the window.
	Configure([]Option)
	// SetCursor updates the current cursor to name.
	SetCursor(cursor pointer.Cursor)
	// Wakeup wakes up the event loop and sends a WakeupEvent.
	Wakeup()
	// Perform actions on the window.
	Perform(system.Action)
	// EditorStateChanged notifies the driver that the editor state changed.
	EditorStateChanged(old, new EditorState)
	GetFrameBufferSize() image.Point
}

// Make it possible to update the options into Callbacks
// in a pseudo-windowless OS such as Android or iOS
type WindowRendezvous struct {
	in   chan WindowAndConfig
	out  chan WindowAndConfig
	errs chan error
}

type WindowAndConfig struct {
	Window  *Callbacks
	Options []Option
}

func NewWindowRendezvous() *WindowRendezvous {
	wr := &WindowRendezvous{
		in:   make(chan WindowAndConfig),
		out:  make(chan WindowAndConfig),
		errs: make(chan error),
	}
	go func() {
		var main WindowAndConfig
		var out chan WindowAndConfig
		for {
			select {
			case w := <-wr.in:
				var err error
				if main.Window != nil {
					err = errors.New("multiple windows are not supported")
				}
				wr.errs <- err
				main = w
				out = wr.out
			case out <- main:
			}
		}
	}()
	return wr
}

func WalkActions(actions system.Action, do func(system.Action)) {
	for a := system.Action(1); actions != 0; a <<= 1 {
		if actions&a != 0 {
			actions &^= a
			do(a)
		}
	}
}

func (ConfigEvent) ImplementsEvent() {}
func (WakeupEvent) ImplementsEvent() {}
func (StageEvent) ImplementsEvent()  {}
