package window

import "image"

type MoveEvent struct {
	Pos image.Point
}
type SizeEvent struct {
	Size image.Point
}

func (MoveEvent) ImplementsEvent() {}
func (SizeEvent) ImplementsEvent() {}
