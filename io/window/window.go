package window

import (
	"image"

	"github.com/kanryu/mado/f32"
)

type MoveEvent struct {
	Pos image.Point
}
type SizeEvent struct {
	Size image.Point
}
type FramebufferSizeEvent struct {
	Size image.Point
}
type FrameScaleEvent struct {
	Scaling f32.Point
}

type CloseEvent struct {
}

func (MoveEvent) ImplementsEvent()            {}
func (SizeEvent) ImplementsEvent()            {}
func (FramebufferSizeEvent) ImplementsEvent() {}
func (FrameScaleEvent) ImplementsEvent()      {}
func (CloseEvent) ImplementsEvent()           {}
