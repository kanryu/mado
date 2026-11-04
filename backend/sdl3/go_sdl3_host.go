package sdl3

import (
	"time"

	goSDL "github.com/kanryu/go-sdl3/sdl"
	"github.com/kanryu/mado"
)

// GoSDL3Host is the ready-to-use Host for github.com/kanryu/go-sdl3.
type GoSDL3Host struct{}

func NewGoSDL3Host() *GoSDL3Host { return &GoSDL3Host{} }
func (*GoSDL3Host) InitVideo() error {
	if !goSDL.Init(goSDL.InitVideo) {
		return sdlError(goSDL.GetError())
	}
	return nil
}
func (*GoSDL3Host) CreateWindow(title string, width, height int, resizable, visible bool) (HostWindow, error) {
	var flags uint64
	if resizable {
		flags |= goSDL.WindowResizable
	}
	w := goSDL.CreateWindow(title, width, height, int(flags))
	if w == nil {
		return nil, sdlError(goSDL.GetError())
	}
	return &goSDLWindow{window: w}, nil
}
func (*GoSDL3Host) PollEvent() (Event, bool) {
	event, ok := goSDL.PollEventTyped()
	if !ok {
		return Event{}, false
	}
	switch e := event.(type) {
	case goSDL.QuitEvent:
		return Event{Type: EventQuit}, true
	case goSDL.WindowEvent:
		return Event{Type: e.Type, WindowID: uint64(e.WindowID), Width: int(e.Data1), Height: int(e.Data2)}, true
	case goSDL.KeyboardEvent:
		return Event{Type: e.Type, WindowID: uint64(e.WindowID), Key: e.KeyName, Down: e.Down, Repeat: e.Repeat}, true
	case goSDL.TextInputEvent:
		return Event{Type: EventTextInput, WindowID: uint64(e.WindowID), Text: e.Text}, true
	case goSDL.TextEditingEvent:
		return Event{Type: EventTextEditing, WindowID: uint64(e.WindowID), Text: e.Text}, true
	case goSDL.MouseMotionEvent:
		return Event{Type: EventMouseMotion, WindowID: uint64(e.WindowID), X: float64(e.X), Y: float64(e.Y)}, true
	case goSDL.MouseButtonEvent:
		return Event{Type: e.Type, WindowID: uint64(e.WindowID), Button: e.Button, Down: e.Down, X: float64(e.X), Y: float64(e.Y)}, true
	default:
		return Event{}, true
	}
}
func (*GoSDL3Host) Delay(duration time.Duration) { time.Sleep(duration) }
func (*GoSDL3Host) Quit()                        { goSDL.Quit() }

type goSDLWindow struct{ window *goSDL.Window }

func (w *goSDLWindow) Destroy()                        { w.window.Destroy() }
func (w *goSDLWindow) SetSize(width, height int) error { return w.window.SetSize(width, height) }
func (w *goSDLWindow) Size() (int, int)                { return w.window.Size() }
func (w *goSDLWindow) Title() string                   { return w.window.Title() }
func (w *goSDLWindow) SetTitle(title string) error     { return w.window.SetTitle(title) }
func (w *goSDLWindow) NativeHandle() mado.NativeHandle {
	return mado.NativeHandle{Kind: mado.HandleSDL3, Value: w.window.NativeHandle()}
}
func (w *goSDLWindow) StartTextInput() error { return w.window.StartTextInput() }
func (w *goSDLWindow) StopTextInput() error  { return w.window.StopTextInput() }

type sdlError string

func (e sdlError) Error() string { return "sdl3: " + string(e) }
