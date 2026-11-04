package mado

import "testing"

func TestSurfaceTargetForwardsNativeHandle(t *testing.T) {
	w := &StubWindow{SizeValue: Size{Width: 1280, Height: 720}, Handle: NativeHandle{Kind: HandleWin32HWND, Value: 42}}
	target, err := SurfaceTargetFor(w)
	if err != nil {
		t.Fatal(err)
	}
	if target.Native.Value != 42 || target.Size.Width != 1280 {
		t.Fatalf("unexpected target: %+v", target)
	}
}

func TestSurfaceTargetRejectsInvalidSize(t *testing.T) {
	_, err := SurfaceTargetFor(&StubWindow{SizeValue: Size{Width: 0, Height: 720}})
	if err != ErrInvalidSize {
		t.Fatalf("expected ErrInvalidSize, got %v", err)
	}
}

func TestQueueBackendCreatesWindow(t *testing.T) {
	b := NewQueueBackend(BackendSDL3)
	w, err := b.CreateWindow(WindowOptions{Title: "test", Size: Size{Width: 640, Height: 480}})
	if err != nil {
		t.Fatal(err)
	}
	if w.ID() != 1 || w.Title() != "test" {
		t.Fatalf("unexpected window: id=%d title=%q", w.ID(), w.Title())
	}
}

func TestWindowSharesInputSession(t *testing.T) {
	b := NewQueueBackend(BackendSDL3)
	w, err := b.CreateWindow(WindowOptions{Title: "test", Size: Size{Width: 640, Height: 480}})
	if err != nil {
		t.Fatal(err)
	}
	if w.Input() == nil || w.Input().Backend() != BackendSDL3 {
		t.Fatal("window session was not initialized")
	}
	if w.Input() != w.Input() || w.Input().Mouse() != w.Input().Mouse() || w.Input().Keyboard() != w.Input().Keyboard() || w.Input().Joystick() != w.Input().Joystick() {
		t.Fatal("input devices are not shared by the session")
	}
}
