package mado

import (
	"errors"
	"image"
	"image/color"
	"unicode"
	"unicode/utf16"

	"github.com/kanryu/mado/f32"
	"github.com/kanryu/mado/gpu"
	"github.com/kanryu/mado/io/event"
	"github.com/kanryu/mado/io/input"
	"github.com/kanryu/mado/io/key"
	"github.com/kanryu/mado/io/pointer"
	"github.com/kanryu/mado/io/system"
	"github.com/kanryu/mado/unit"
)

type Context interface {
	API() gpu.API
	RenderTarget() (gpu.RenderTarget, error)
	Present() error
	Refresh() error
	Release()
	Lock() error
	Unlock()
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
}

// Option configures a window.
type Option func(unit.Metric, *Config)

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

type EditorState struct {
	input.EditorState
	Compose key.Range
}

func (e *EditorState) Replace(r key.Range, text string) {
	if r.Start > r.End {
		r.Start, r.End = r.End, r.Start
	}
	runes := []rune(text)
	newEnd := r.Start + len(runes)
	adjust := func(pos int) int {
		switch {
		case newEnd < pos && pos <= r.End:
			return newEnd
		case r.End < pos:
			diff := newEnd - r.End
			return pos + diff
		}
		return pos
	}
	e.Selection.Start = adjust(e.Selection.Start)
	e.Selection.End = adjust(e.Selection.End)
	if e.Compose.Start != -1 {
		e.Compose.Start = adjust(e.Compose.Start)
		e.Compose.End = adjust(e.Compose.End)
	}
	s := e.Snippet
	if r.End < s.Start || r.Start > s.End {
		// Discard snippet if it doesn't overlap with replacement.
		s = key.Snippet{
			Range: key.Range{
				Start: r.Start,
				End:   r.Start,
			},
		}
	}
	var newSnippet []rune
	snippet := []rune(s.Text)
	// Append first part of existing snippet.
	if end := r.Start - s.Start; end > 0 {
		newSnippet = append(newSnippet, snippet[:end]...)
	}
	// Append replacement.
	newSnippet = append(newSnippet, runes...)
	// Append last part of existing snippet.
	if start := r.End; start < s.End {
		newSnippet = append(newSnippet, snippet[start-s.Start:]...)
	}
	// Adjust snippet range to include replacement.
	if r.Start < s.Start {
		s.Start = r.Start
	}
	s.End = s.Start + len(newSnippet)
	s.Text = string(newSnippet)
	e.Snippet = s
}

// UTF16Index converts the given index in runes into an index in utf16 characters.
func (e *EditorState) UTF16Index(runes int) int {
	if runes == -1 {
		return -1
	}
	if runes < e.Snippet.Start {
		// Assume runes before sippet are one UTF-16 character each.
		return runes
	}
	chars := e.Snippet.Start
	runes -= e.Snippet.Start
	for _, r := range e.Snippet.Text {
		if runes == 0 {
			break
		}
		runes--
		chars++
		if r1, _ := utf16.EncodeRune(r); r1 != unicode.ReplacementChar {
			chars++
		}
	}
	// Assume runes after snippets are one UTF-16 character each.
	return chars + runes
}

// RunesIndex converts the given index in utf16 characters to an index in runes.
func (e *EditorState) RunesIndex(chars int) int {
	if chars == -1 {
		return -1
	}
	if chars < e.Snippet.Start {
		// Assume runes before offset are one UTF-16 character each.
		return chars
	}
	runes := e.Snippet.Start
	chars -= e.Snippet.Start
	for _, r := range e.Snippet.Text {
		if chars == 0 {
			break
		}
		chars--
		runes++
		if r1, _ := utf16.EncodeRune(r); r1 != unicode.ReplacementChar {
			chars--
		}
	}
	// Assume runes after snippets are one UTF-16 character each.
	return runes + chars
}

type Callbacks interface {
	SetWindow(w *Window)
	SetDriver(d Driver)
	Event(e event.Event) bool
	EditorInsert(text string)
	EditorState() EditorState
	EditorReplace(r key.Range, text string)
	SetComposingRegion(r key.Range)
	SetEditorSelection(r key.Range)
	SetEditorSnippet(r key.Range)
	ActionAt(p f32.Point) (system.Action, bool)
}

// type Callbacks struct {
// 	W          *Window
// 	D          Driver
// 	Busy       bool
// 	WaitEvents []event.Event
// }

// WindowMode is the window mode (WindowMode.Option sets it).
// Note that mode can be changed programatically as well as by the user
// clicking on the minimize/maximize buttons on the window's title bar.
type WindowMode uint8

const (
	// Windowed is the normal window mode with OS specific window decorations.
	Windowed WindowMode = iota
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
func (ConfigEvent) ImplementsEvent() {}

// ErrOutOfDate is reported when the GPU surface dimensions or properties no
// longer match the window.
var ErrOutOfDate = errors.New("app: GPU surface out of date")
