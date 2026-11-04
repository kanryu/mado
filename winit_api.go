package mado

// ApplicationHandler is the Winit-inspired application lifecycle used by
// Mado. Backends translate their native event loop into these callbacks.
type ApplicationHandler interface {
	Resumed(*ActiveEventLoop)
	Suspended(*ActiveEventLoop)
	WindowEvent(*ActiveEventLoop, WindowID, WindowEvent)
	AboutToWait(*ActiveEventLoop)
}

// WindowAttributes describes a window before it is created.
type WindowAttributes struct {
	Title     string
	InnerSize Size
	Resizable bool
	HighDPI   bool
	Visible   bool
}

func DefaultWindowAttributes() WindowAttributes {
	return WindowAttributes{
		Title:     "Mado Window",
		InnerSize: Size{Width: 800, Height: 600},
		Resizable: true,
		HighDPI:   true,
		Visible:   true,
	}
}

func (a WindowAttributes) withDefaults() WindowAttributes {
	defaults := DefaultWindowAttributes()
	if a.Title == "" {
		a.Title = defaults.Title
	}
	if a.InnerSize.Width == 0 {
		a.InnerSize.Width = defaults.InnerSize.Width
	}
	if a.InnerSize.Height == 0 {
		a.InnerSize.Height = defaults.InnerSize.Height
	}
	return a
}

// WindowEvent is the Winit-style event set exposed by Mado.
type WindowEvent interface{ isWindowEvent() }

type WindowCloseRequested struct{}

func (WindowCloseRequested) isWindowEvent() {}

type WindowResized struct{ Size Size }

func (WindowResized) isWindowEvent() {}

type WindowRedrawRequested struct{}

func (WindowRedrawRequested) isWindowEvent() {}

type WindowScaleFactorChanged struct{ Scale float64 }

func (WindowScaleFactorChanged) isWindowEvent() {}

type WindowKeyboardInput struct {
	Key    string
	Down   bool
	Repeat bool
}

func (WindowKeyboardInput) isWindowEvent() {}

type WindowCursorMoved struct{ Position Point }

func (WindowCursorMoved) isWindowEvent() {}

type WindowMouseInput struct {
	Button   uint8
	Down     bool
	Position Point
}

func (WindowMouseInput) isWindowEvent() {}

type WindowReceivedCharacter struct{ Character rune }

func (WindowReceivedCharacter) isWindowEvent() {}

type WindowFocused struct{ Focused bool }

func (WindowFocused) isWindowEvent() {}

type WindowOccluded struct{ Occluded bool }

func (WindowOccluded) isWindowEvent() {}

type WindowMinimizedEvent struct{}

func (WindowMinimizedEvent) isWindowEvent() {}

type WindowMaximizedEvent struct{}

func (WindowMaximizedEvent) isWindowEvent() {}

type WindowRestoredEvent struct{}

func (WindowRestoredEvent) isWindowEvent() {}

// EventLoop owns the selected backend and translates it to the Winit-shaped
// Mado API. It must be created and run on the thread required by the backend.
type EventLoop struct {
	backend Backend
	exit    bool
}

func NewEventLoop(backend Backend) (*EventLoop, error) {
	if backend == nil {
		return nil, ErrUnsupported
	}
	return &EventLoop{backend: backend}, nil
}

// ActiveEventLoop is passed during callbacks. Windows should be created from
// Resumed, matching Winit's lifecycle and avoiding premature native creation.
type ActiveEventLoop struct{ loop *EventLoop }

func (e *ActiveEventLoop) CreateWindow(attributes WindowAttributes) (Window, error) {
	return NewWindow(e.loop.backend, attributes)
}

func (e *ActiveEventLoop) Exit() {
	e.loop.exit = true
	e.loop.backend.Quit()
}

func (a WindowAttributes) windowOptions() WindowOptions {
	return WindowOptions{
		Title: a.Title, Size: a.InnerSize, Resizable: a.Resizable,
		HighDPI: a.HighDPI, Visible: a.Visible,
	}
}

// RunApp starts the backend event loop and dispatches events to handler.
func (e *EventLoop) RunApp(handler ApplicationHandler) error {
	if handler == nil {
		return ErrUnsupported
	}
	active := &ActiveEventLoop{loop: e}
	handler.Resumed(active)
	if e.exit {
		return nil
	}
	err := e.backend.Run(func(event Event) error {
		if e.exit {
			return nil
		}
		switch v := event.(type) {
		case CloseRequested:
			handler.WindowEvent(active, 0, WindowCloseRequested{})
			// A close request terminates the application loop after the
			// application has had a chance to release its window resources.
			active.Exit()
		case Resized:
			handler.WindowEvent(active, 0, WindowResized{Size: v.Size})
		case ScaleChanged:
			handler.WindowEvent(active, 0, WindowScaleFactorChanged{Scale: v.Scale})
		case KeyEvent:
			handler.WindowEvent(active, 0, WindowKeyboardInput{Key: v.Key, Down: v.Down, Repeat: v.Repeat})
		case MouseMoved:
			handler.WindowEvent(active, 0, WindowCursorMoved{Position: v.Position})
		case MouseButtonEvent:
			handler.WindowEvent(active, 0, WindowMouseInput{Button: v.Button, Down: v.Down, Position: v.Position})
		case ReceivedCharacter:
			handler.WindowEvent(active, 0, WindowReceivedCharacter{Character: v.Character})
		case FocusChanged:
			handler.WindowEvent(active, 0, WindowFocused{Focused: v.Focused})
		case OccludedChanged:
			handler.WindowEvent(active, 0, WindowOccluded{Occluded: v.Occluded})
		case WindowMinimized:
			handler.WindowEvent(active, 0, WindowMinimizedEvent{})
		case WindowMaximized:
			handler.WindowEvent(active, 0, WindowMaximizedEvent{})
		case WindowRestored:
			handler.WindowEvent(active, 0, WindowRestoredEvent{})
		}
		return nil
	})
	if err != nil {
		return err
	}
	if !e.exit {
		handler.AboutToWait(active)
	}
	return nil
}
