package mado

import (
	"image"
	"time"

	"github.com/kanryu/mado/io/input"
	"github.com/kanryu/mado/op"
	"github.com/kanryu/mado/unit"
)

// ViewEvent provides handles to the underlying window objects for the
// current display protocol.
type ViewEvent interface {
	ImplementsViewEvent()
	ImplementsEvent()
}

// A FrameEvent requests a new frame in the form of a list of
// operations that describes the window content.
type FrameEvent struct {
	// Now is the current animation. Use Now instead of time.Now to
	// synchronize animation and to avoid the time.Now call overhead.
	Now time.Time
	// Metric converts device independent dp and sp to device pixels.
	Metric unit.Metric
	// Size is the dimensions of the window.
	Size image.Point
	// Insets represent the space occupied by system decorations and controls.
	Insets Insets
	// Frame completes the FrameEvent by drawing the graphical operations
	// from ops into the window.
	Frame func(frame *op.Ops)
	// Source is the interface between the window and widgets.
	Source input.Source

	Sync bool
}

// Insets is the space taken up by
// system decoration such as translucent
// system bars and software keyboards.
type Insets struct {
	// Values are in pixels.
	Top, Bottom, Left, Right unit.Dp
}

// DestroyEvent is the last event sent through
// a window event channel.
type DestroyEvent struct {
	// Err is nil for normal window closures. If a
	// window is prematurely closed, Err is the cause.
	Err error
}

func (FrameEvent) ImplementsEvent()   {}
func (DestroyEvent) ImplementsEvent() {}
