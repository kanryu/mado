package sdl3

import (
	"testing"
	"time"

	"github.com/kanryu/mado"
)

type fakeHost struct {
	events []Event
	closed bool
}

func (*fakeHost) InitVideo() error                                              { return nil }
func (*fakeHost) CreateWindow(string, int, int, bool, bool) (HostWindow, error) { return nil, nil }
func (h *fakeHost) PollEvent() (Event, bool) {
	if len(h.events) == 0 {
		return Event{}, false
	}
	e := h.events[0]
	h.events = h.events[1:]
	return e, true
}
func (*fakeHost) Delay(time.Duration) {}
func (h *fakeHost) Quit()             { h.closed = true }

func TestConvertSDL3Events(t *testing.T) {
	event := convertEvent(Event{Type: EventMouseMotion, X: 12, Y: 34})
	moved, ok := event.(mado.MouseMoved)
	if !ok || moved.Position != (mado.Point{X: 12, Y: 34}) {
		t.Fatalf("unexpected event: %#v", event)
	}
}

func TestConvertWindowStateEventsSeparately(t *testing.T) {
	tests := []struct {
		typ  uint32
		want any
	}{
		{EventWindowMinimized, mado.WindowMinimized{}},
		{EventWindowMaximized, mado.WindowMaximized{}},
		{EventWindowRestored, mado.WindowRestored{}},
	}
	for _, test := range tests {
		got := convertEvent(Event{Type: test.typ})
		if got != test.want {
			t.Errorf("event %x: got %#v, want %#v", test.typ, got, test.want)
		}
	}
}
