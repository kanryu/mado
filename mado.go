// Package mado is a small, backend-independent Windows application framework.
//
// Window creation and event delivery are deliberately separated from rendering.
// This lets an application use SDL3 or Winit for window management while using
// the same mado API and a WGPU surface for rendering.
package mado

import "errors"

var (
	ErrClosed      = errors.New("mado: window is closed")
	ErrUnsupported = errors.New("mado: operation is unsupported by this backend")
	ErrInvalidSize = errors.New("mado: window size must be positive")
)

type Size struct{ Width, Height int }

type Point struct{ X, Y int }

type WindowID uint64

type WindowOptions struct {
	Title     string
	Size      Size
	Resizable bool
	HighDPI   bool
	Visible   bool
	Backend   BackendKind
}

type BackendKind string

const (
	BackendAuto  BackendKind = "auto"
	BackendSDL3  BackendKind = "sdl3"
	BackendWinit BackendKind = "winit"
)

// NativeHandle is intentionally opaque to the core package. Renderer adapters
// can type-switch on its Kind and use the platform-specific value.
type NativeHandle struct {
	Kind  NativeHandleKind
	Value uintptr
	Aux   uintptr
}

type NativeHandleKind string

const (
	HandleWin32HWND NativeHandleKind = "win32-hwnd"
	HandleSDL3      NativeHandleKind = "sdl3-window"
	HandleWinit     NativeHandleKind = "winit-window"
)

type Event interface{ isEvent() }

type CloseRequested struct{}

func (CloseRequested) isEvent() {}

type Resized struct{ Size Size }

func (Resized) isEvent() {}

type ScaleChanged struct{ Scale float64 }

func (ScaleChanged) isEvent() {}

type KeyEvent struct {
	Key          string
	Down, Repeat bool
}

func (KeyEvent) isEvent() {}

type MouseMoved struct{ Position Point }

func (MouseMoved) isEvent() {}

type MouseButtonEvent struct {
	Button   uint8
	Down     bool
	Position Point
}

func (MouseButtonEvent) isEvent() {}

type ReceivedCharacter struct{ Character rune }

func (ReceivedCharacter) isEvent() {}

type FocusChanged struct{ Focused bool }

func (FocusChanged) isEvent() {}

type OccludedChanged struct{ Occluded bool }

func (OccludedChanged) isEvent() {}

type WindowMinimized struct{}

func (WindowMinimized) isEvent() {}

type WindowMaximized struct{}

func (WindowMaximized) isEvent() {}

type WindowRestored struct{}

func (WindowRestored) isEvent() {}

type Backend interface {
	Name() BackendKind
	CreateWindow(WindowOptions) (Window, error)
	Run(func(Event) error) error
	Quit()
}

type Window interface {
	ID() WindowID
	Title() string
	Size() Size
	SetSize(Size) error
	SetTitle(string) error
	NativeHandle() NativeHandle
	Close() error
	Input() *Input
}

// App is the common application lifecycle used by all backends.
type App struct{ backend Backend }

func New(backend Backend) (*App, error) {
	if backend == nil {
		return nil, errors.New("mado: backend is nil")
	}
	return &App{backend: backend}, nil
}

func (a *App) Backend() Backend { return a.backend }

// CreateWindow creates a window using the backend selected when the App was
// constructed. Applications should normally use this instead of calling the
// backend directly.
func (a *App) CreateWindow(options WindowOptions) (Window, error) {
	return a.backend.CreateWindow(options)
}

func (a *App) Run(window Window, render func(RenderContext) error) error {
	if window == nil {
		return errors.New("mado: window is nil")
	}
	if render == nil {
		return errors.New("mado: render callback is nil")
	}
	if err := render(RenderContext{Window: window, Size: window.Size()}); err != nil {
		return err
	}
	return a.backend.Run(func(event Event) error {
		switch e := event.(type) {
		case CloseRequested:
			return ErrClosed
		case Resized:
			return render(RenderContext{Window: window, Size: e.Size})
		default:
			return nil
		}
	})
}

type RenderContext struct {
	Window Window
	Size   Size
}
