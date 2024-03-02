package window

import "image"

type MoveEvent struct {
	Pos image.Point
}
type SizeEvent struct {
	Size image.Point
}
type CloseEvent struct {
}

func (MoveEvent) ImplementsEvent()  {}
func (SizeEvent) ImplementsEvent()  {}
func (CloseEvent) ImplementsEvent() {}
