//go:build windows && amd64

package winit

import (
	"errors"
	"syscall"

	"github.com/kanryu/go-winit/winit"
	"github.com/kanryu/mado"
)

var (
	imm32             = syscall.NewLazyDLL("imm32.dll")
	kernel32          = syscall.NewLazyDLL("kernel32.dll")
	immGetContext     = imm32.NewProc("ImmGetContext")
	immGetOpenStatus  = imm32.NewProc("ImmGetOpenStatus")
	immReleaseContext = imm32.NewProc("ImmReleaseContext")
	getModuleHandle   = kernel32.NewProc("GetModuleHandleW")
)

// NativeFactory delegates window creation and the message loop to go-winit.
// Mado only translates go-winit events into its common event types.
type NativeFactory struct {
	loop    *winit.EventLoop
	proxy   *winit.EventLoopProxy
	win     *NativeWindow
	session *mado.Session
}

type NativeWindow struct {
	native  *winit.Window
	id      mado.WindowID
	title   string
	size    mado.Size
	session *mado.Session
}

func (f *NativeFactory) Name() mado.BackendKind { return mado.BackendWinit }

func NewNativeFactory() (*NativeFactory, error) {
	loop := winit.NewEventLoop()
	if loop == nil {
		return nil, errors.New("mado/winit: failed to create go-winit event loop")
	}
	return &NativeFactory{loop: loop, proxy: loop.CreateProxy()}, nil
}

// NewFactory creates the native Winit window driver.
func NewFactory() (*NativeFactory, error) { return NewNativeFactory() }

func (f *NativeFactory) CreateWindow(o mado.WindowOptions) (mado.Window, error) {
	if f == nil || f.loop == nil {
		return nil, errors.New("mado/winit: event loop is closed")
	}
	if o.Size.Width <= 0 || o.Size.Height <= 0 {
		return nil, mado.ErrInvalidSize
	}
	a := winit.NewWindowAttributes().WithTitle(o.Title).WithDimensions(float64(o.Size.Width), float64(o.Size.Height)).WithResizable(o.Resizable).WithVisibility(o.Visible)
	if a == nil {
		return nil, errors.New("mado/winit: failed to create window attributes")
	}
	native := f.loop.CreateWindow(a)
	if native == nil {
		return nil, errors.New("mado/winit: failed to create window")
	}
	native.SetIMEAllowed(true)
	native.Focus()
	high, low := native.ID()
	if f.session == nil {
		f.session = mado.NewSession(mado.BackendWinit)
	}
	w := &NativeWindow{native: native, id: mado.WindowID(low), title: o.Title, size: o.Size, session: f.session}
	f.win = w
	_ = high
	return w, nil
}

func (f *NativeFactory) Run(dispatch func(mado.Event) error) error {
	if f == nil || f.loop == nil {
		return errors.New("mado/winit: event loop is closed")
	}
	var dispatchErr error
	f.loop.RunBatched(func(events []winit.Event) winit.ControlFlow {
		if dispatchErr != nil {
			return winit.Exit
		}
		var pendingResize *mado.Resized
		for _, event := range events {
			if event.Type == winit.WindowEventResized {
				pendingResize = &mado.Resized{Size: mado.Size{Width: int(event.ResizedWidth), Height: int(event.ResizedHeight)}}
				continue
			}
			var translated *mado.Event
			switch event.Type {
			case winit.WindowEventCloseRequested:
				v := mado.Event(mado.CloseRequested{})
				if dispatchErr = dispatch(v); dispatchErr != nil {
					return winit.Exit
				}
				// CloseRequested is terminal for the Mado application. Returning
				// Exit here prevents the native loop from waiting for another event.
				return winit.Exit
			case winit.WindowEventCursorMoved:
				v := mado.Event(mado.MouseMoved{Position: mado.Point{X: int(event.CursorX), Y: int(event.CursorY)}})
				translated = &v
			case winit.WindowEventReceivedCharacter:
				for _, r := range []rune(event.ReceivedCharacterString()) {
					v := mado.Event(mado.ReceivedCharacter{Character: r})
					if dispatchErr = dispatch(v); dispatchErr != nil {
						return winit.Exit
					}
				}
			case winit.WindowEventKeyboardInput:
				// go-winit suppresses keyboard events after Winit emits Ime::Enabled.
				// On Windows, query IMM32 as well because enabling IME support on a
				// window does not guarantee that the IME is currently open.
				if imeIsOpen(f.win) {
					return winit.Wait
				}
				v := mado.Event(mado.KeyEvent{Key: event.LogicalKey, Down: event.KeyboardState == winit.InputElementStatePressed})
				translated = &v
			case winit.WindowEventScaleFactorChanged:
				v := mado.Event(mado.ScaleChanged{Scale: event.ScaleFactor})
				translated = &v
			case winit.WindowEventFocused:
				// The low-level event record does not currently expose the focused value.
				v := mado.Event(mado.FocusChanged{Focused: true})
				translated = &v
			case winit.WindowEventOccluded:
				v := mado.Event(mado.OccludedChanged{Occluded: event.Occluded})
				translated = &v
			}
			if translated != nil {
				dispatchErr = dispatch(*translated)
				if dispatchErr != nil {
					return winit.Exit
				}
			}
		}
		if pendingResize != nil {
			dispatchErr = dispatch(*pendingResize)
			if dispatchErr != nil {
				return winit.Exit
			}
		}
		return winit.Wait
	})
	return dispatchErr
}

func imeIsOpen(window *NativeWindow) bool {
	if window == nil || window.native == nil {
		return false
	}
	hwnd := window.native.HWND()
	if hwnd == nil {
		return false
	}
	context, _, _ := immGetContext.Call(uintptr(hwnd))
	if context == 0 {
		return false
	}
	open, _, _ := immGetOpenStatus.Call(context)
	immReleaseContext.Call(uintptr(hwnd), context)
	return open != 0
}

func (f *NativeFactory) Quit() {
	if f != nil && f.proxy != nil {
		f.proxy.Exit()
	}
}

func (w *NativeWindow) ID() mado.WindowID { return w.id }
func (w *NativeWindow) Title() string     { return w.title }
func (w *NativeWindow) Size() mado.Size {
	width, height := w.native.Size()
	if width != 0 && height != 0 {
		return mado.Size{Width: int(width), Height: int(height)}
	}
	return w.size
}
func (w *NativeWindow) SetSize(s mado.Size) error {
	if s.Width <= 0 || s.Height <= 0 {
		return mado.ErrInvalidSize
	}
	w.native.SetSize(uint32(s.Width), uint32(s.Height))
	w.size = s
	return nil
}
func (w *NativeWindow) SetTitle(title string) error { w.title = title; return nil }
func (w *NativeWindow) NativeHandle() mado.NativeHandle {
	instance, _, _ := getModuleHandle.Call(0)
	return mado.NativeHandle{Kind: mado.HandleWin32HWND, Value: uintptr(w.native.HWND()), Aux: instance}
}
func (w *NativeWindow) Close() error       { w.native.Close(); return nil }
func (w *NativeWindow) Input() *mado.Input { return w.session.Input() }
