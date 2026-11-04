package mado

import "testing"

type testHandler struct {
	resumed bool
	created Window
}

func (h *testHandler) Resumed(loop *ActiveEventLoop) {
	h.resumed = true
	h.created, _ = loop.CreateWindow(WindowAttributes{})
}
func (*testHandler) Suspended(*ActiveEventLoop) {}
func (*testHandler) WindowEvent(loop *ActiveEventLoop, _ WindowID, event WindowEvent) {
	if _, ok := event.(WindowCloseRequested); ok {
		loop.Exit()
	}
}
func (*testHandler) AboutToWait(*ActiveEventLoop) {}

func TestWinitStyleEventLoop(t *testing.T) {
	backend := NewQueueBackend(BackendWinit)
	backend.Push(CloseRequested{})
	loop, err := NewEventLoop(backend)
	if err != nil {
		t.Fatal(err)
	}
	h := &testHandler{}
	if err := loop.RunApp(h); err != nil {
		t.Fatal(err)
	}
	if !h.resumed || h.created == nil {
		t.Fatal("Resumed did not create a window")
	}
	if h.created.Size() != (Size{Width: 800, Height: 600}) {
		t.Fatalf("unexpected default size: %+v", h.created.Size())
	}
}
