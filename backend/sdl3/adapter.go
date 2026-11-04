package sdl3

import (
	"errors"
	"time"

	"github.com/kanryu/mado"
)

// SDL event identifiers. They intentionally mirror SDL3's public event IDs
// without importing a particular SDL3 Go binding.
const (
	EventQuit            uint32 = 0x0100
	EventWindowResized   uint32 = 0x0206
	EventWindowMinimized uint32 = 0x0209
	EventWindowMaximized uint32 = 0x020A
	EventWindowRestored  uint32 = 0x020B
	EventWindowFocusIn   uint32 = 0x020E
	EventWindowFocusOut  uint32 = 0x020F
	EventWindowClose     uint32 = 0x0210
	EventKeyDown         uint32 = 0x0300
	EventKeyUp           uint32 = 0x0301
	EventTextEditing     uint32 = 0x0302
	EventTextInput       uint32 = 0x0303
	EventMouseMotion     uint32 = 0x0400
	EventMouseButtonDown uint32 = 0x0401
	EventMouseButtonUp   uint32 = 0x0402
)

type Event struct {
	Type     uint32
	WindowID uint64
	Width    int
	Height   int
	X        float64
	Y        float64
	Button   uint8
	Down     bool
	Key      string
	Repeat   bool
	Text     string
	Occluded bool
}

// HostWindow is the small surface required from a concrete SDL3 binding.
// A go-sdl3 adapter can implement this using its SDL_Window handle.
type HostWindow interface {
	Destroy()
	SetSize(width, height int) error
	Size() (width, height int)
	Title() string
	SetTitle(string) error
	NativeHandle() mado.NativeHandle
	StartTextInput() error
	StopTextInput() error
}

// Host is implemented by the SDL3 binding adapter. PollEvent must be called
// only from the thread on which the SDL video subsystem was initialized.
type Host interface {
	InitVideo() error
	CreateWindow(title string, width, height int, resizable bool, visible bool) (HostWindow, error)
	PollEvent() (Event, bool)
	Delay(time.Duration)
	Quit()
}

type Factory struct {
	host        Host
	win         *sdlWindow
	session     *mado.Session
	quit        bool
	initialized bool
}

type sdlWindow struct {
	native  HostWindow
	id      mado.WindowID
	session *mado.Session
}

// NewFactory creates an SDL-backed factory. Native SDL initialization is
// deferred until mado.NewWindow invokes CreateWindow.
// NewFactory creates the native SDL3 window driver. Applications normally use
// this constructor and do not need to know about the Host abstraction. The
// optional argument preserves the convenient injection path for existing
// tests and custom hosts.
func NewFactory(hosts ...Host) (*Factory, error) {
	if len(hosts) > 1 {
		return nil, errors.New("mado/sdl3: at most one host is allowed")
	}
	if len(hosts) == 1 {
		return NewFactoryWithHost(hosts[0])
	}
	return NewFactoryWithHost(NewGoSDL3Host())
}

// NewFactoryWithHost creates a driver with an injected host. It is intended
// for tests and applications that provide their own SDL event source.
func NewFactoryWithHost(host Host) (*Factory, error) {
	if host == nil {
		return nil, errors.New("mado/sdl3: host is nil")
	}
	return &Factory{host: host}, nil
}

func (f *Factory) Name() mado.BackendKind { return mado.BackendSDL3 }

func (f *Factory) CreateWindow(o mado.WindowOptions) (mado.Window, error) {
	if !f.initialized {
		if err := f.host.InitVideo(); err != nil {
			return nil, err
		}
		f.initialized = true
	}
	if o.Size.Width <= 0 || o.Size.Height <= 0 {
		return nil, mado.ErrInvalidSize
	}
	w, err := f.host.CreateWindow(o.Title, o.Size.Width, o.Size.Height, o.Resizable, o.Visible)
	if err != nil {
		return nil, err
	}
	if err := w.StartTextInput(); err != nil {
		return nil, err
	}
	if f.session == nil {
		f.session = mado.NewSession(mado.BackendSDL3)
	}
	f.win = &sdlWindow{native: w, id: 1, session: f.session}
	return f.win, nil
}

func (f *Factory) Run(dispatch func(mado.Event) error) error {
	timeout := 10 * time.Millisecond
	timeSinceEvent := func(hadEvent bool) {
		if !hadEvent {
			f.host.Delay(timeout)
		}
	}
	timeSinceEvent(false)
	for !f.quit {
		hadEvent := false
		var pendingResize *mado.Resized
		for {
			event, ok := f.host.PollEvent()
			if !ok {
				break
			}
			hadEvent = true
			if event.Type == EventQuit || event.Type == EventWindowClose {
				if err := dispatch(mado.CloseRequested{}); err != nil {
					return err
				}
				f.Quit()
				return nil
			}
			if event.Type == EventTextInput {
				for _, character := range []rune(event.Text) {
					if err := dispatch(mado.ReceivedCharacter{Character: character}); err != nil {
						return err
					}
				}
				continue
			}
			if event.Type == EventWindowResized {
				converted := convertEvent(event).(mado.Resized)
				pendingResize = &converted
				continue
			}
			if err := dispatch(convertEvent(event)); err != nil {
				return err
			}
		}
		if pendingResize != nil {
			if err := dispatch(*pendingResize); err != nil {
				return err
			}
		}
		timeSinceEvent(hadEvent)
	}
	return nil
}

func (f *Factory) Quit() { f.quit = true; f.host.Quit() }

func convertEvent(e Event) mado.Event {
	switch e.Type {
	case EventWindowResized:
		return mado.Resized{Size: mado.Size{Width: e.Width, Height: e.Height}}
	case EventWindowMinimized, EventWindowMaximized:
		if e.Type == EventWindowMinimized {
			return mado.WindowMinimized{}
		}
		return mado.WindowMaximized{}
	case EventWindowRestored:
		return mado.WindowRestored{}
	case EventWindowFocusIn:
		return mado.FocusChanged{Focused: true}
	case EventWindowFocusOut:
		return mado.FocusChanged{Focused: false}
	case EventKeyDown, EventKeyUp:
		return mado.KeyEvent{Key: e.Key, Down: e.Type == EventKeyDown, Repeat: e.Repeat}
	case EventMouseMotion:
		return mado.MouseMoved{Position: mado.Point{X: int(e.X), Y: int(e.Y)}}
	case EventMouseButtonDown, EventMouseButtonUp:
		return mado.MouseButtonEvent{Button: e.Button, Down: e.Type == EventMouseButtonDown, Position: mado.Point{X: int(e.X), Y: int(e.Y)}}
	}
	return nil
}

func (w *sdlWindow) ID() mado.WindowID               { return w.id }
func (w *sdlWindow) Title() string                   { return w.native.Title() }
func (w *sdlWindow) Size() mado.Size                 { x, y := w.native.Size(); return mado.Size{Width: x, Height: y} }
func (w *sdlWindow) SetSize(s mado.Size) error       { return w.native.SetSize(s.Width, s.Height) }
func (w *sdlWindow) SetTitle(title string) error     { return w.native.SetTitle(title) }
func (w *sdlWindow) NativeHandle() mado.NativeHandle { return w.native.NativeHandle() }
func (w *sdlWindow) Close() error                    { _ = w.native.StopTextInput(); w.native.Destroy(); return nil }
func (w *sdlWindow) Input() *mado.Input              { return w.session.Input() }
