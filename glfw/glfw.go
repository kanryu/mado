package glfw

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"reflect"
	"sync"

	"github.com/kanryu/mado"
	"github.com/kanryu/mado/app"
	"github.com/kanryu/mado/font/gofont"
	"github.com/kanryu/mado/io/event"
	"github.com/kanryu/mado/io/key"
	"github.com/kanryu/mado/io/pointer"
	"github.com/kanryu/mado/io/system"
	"github.com/kanryu/mado/layout"
	"github.com/kanryu/mado/op"
	"github.com/kanryu/mado/text"
	"github.com/kanryu/mado/widget/material"
)

// Version constants.
const (
	VersionMajor    = 3 // This is incremented when the API is changed in non-compatible ways.
	VersionMinor    = 3 // This is incremented when features are added to the API but it remains backward-compatible.
	VersionRevision = 9 // This is incremented when a bug fix release is made that does not contain any API changes.
)

type WindowEvent struct {
	w *Window
	e event.Event
}

var theApp *Application

// Application keeps track of all the windows and global state.
type Application struct {
	// Context is used to broadcast application shutdown.
	Context context.Context
	Stop    context.CancelFunc
	// Shutdown shuts down all windows.
	Shutdown func()
	// active keeps track the open windows, such that application
	// can shut down, when all of them are closed.
	active sync.WaitGroup

	chans      []chan WindowEvent
	eventCases []reflect.SelectCase
	windowList []*Window
}

func NewApplication(ctx context.Context, stop context.CancelFunc) *Application {
	ctx, cancel := context.WithCancel(ctx)
	return &Application{
		Context:  ctx,
		Stop:     stop,
		Shutdown: cancel,
	}
}

func (a *Application) appendWindow(w *Window) {
	windows.put(w)

	ch := make(chan WindowEvent)
	a.chans = append(a.chans, ch)
	a.eventCases = append(a.eventCases, reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(ch)})
	a.windowList = append(a.windowList, w)
	a.active.Add(1)

	go func() {
		tag := new(int)
		var ops op.Ops
		for {
			e := w.data.NextEvent()
			ch <- WindowEvent{w: w, e: e}
			switch e2 := e.(type) {
			case mado.DestroyEvent:
				w.shouldClose = true
				close(ch)
				return
			case mado.FrameEvent:
				gtx := app.NewContext(&ops, e2)
				for {
					ev, ok := gtx.Source.Event(pointer.Filter{
						Target: tag,
						Kinds:  pointer.Release,
					})
					if !ok {
						break
					}
					switch ev := ev.(type) {
					case pointer.Event:
						if ev.Kind == pointer.Release {
							gtx.Execute(key.FocusCmd{Tag: tag})
							fmt.Println("triggered focus command")
						}
					}
					fmt.Printf("%#+v\n", ev)
				}
				for {
					ev, ok := gtx.Source.Event(key.Filter{
						Focus: tag,
					})
					if !ok {
						break
					}
					fmt.Printf("%#+v\n", ev)
				}
				event.Op(gtx.Ops, tag)
				e2.Frame(gtx.Ops)
			}
		}
	}()

	// go func() {
	// 	defer a.active.Done()
	// 	a.Run(w)
	// }()
}

// Wait waits for all windows to close.
func (a *Application) Wait() {
	a.active.Wait()
}

// View describes .
type View interface {
	// Run handles the window event loop.
	Run(w *Window) error
}

// WidgetView allows to use layout.Widget as a view.
type WidgetView func(gtx layout.Context, th *material.Theme) layout.Dimensions

// Run displays the widget with default handling.
func (a *Application) Run(w *Window) error {
	var ops op.Ops

	th := material.NewTheme()
	th.Shaper = text.NewShaper(text.WithCollection(gofont.Collection()))

	go func() {
		<-w.App.Context.Done()
		w.data.Perform(system.ActionClose)
	}()
	for {
		switch e := w.data.NextEvent().(type) {
		case mado.DestroyEvent:
			return e.Err
		case mado.FrameEvent:
			gtx := app.NewContext(&ops, e)
			e.Frame(gtx.Ops)
		}
	}
}

// Init initializes the GLFW library. Before most GLFW functions can be used,
// GLFW must be initialized, and before a program terminates GLFW should be
// terminated in order to free any resources allocated during or after
// initialization.
//
// If this function fails, it calls Terminate before returning. If it succeeds,
// you should call Terminate before the program exits.
//
// Additional calls to this function after successful initialization but before
// termination will succeed but will do nothing.
//
// This function may take several seconds to complete on some systems, while on
// other systems it may take only a fraction of a second to complete.
//
// On Mac OS X, this function will change the current directory of the
// application to the Contents/Resources subdirectory of the application's
// bundle, if present.
//
// This function may only be called from the main thread.
func Init() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	theApp = NewApplication(ctx, stop)
	return acceptError(APIUnavailable)
}

// Terminate destroys all remaining windows, frees any allocated resources and
// sets the library to an uninitialized state. Once this is called, you must
// again call Init successfully before you will be able to use most GLFW
// functions.
//
// If GLFW has been successfully initialized, this function should be called
// before the program exits. If initialization fails, there is no need to call
// this function, as it is called by Init before it returns failure.
//
// This function may only be called from the main thread.
func Terminate() {
	theApp.Stop()
	flushErrors()
	fmt.Println("not implemented")
}

// PollEvents processes only those events that have already been received and
// then returns immediately. Processing events will cause the window and input
// callbacks associated with those events to be called.
//
// This function is not required for joystick input to work.
//
// This function may not be called from a callback.
//
// This function may only be called from the main thread.
func PollEvents() {
	remaining := len(theApp.eventCases)
	if remaining <= 0 {
		return
	}
	chosen, _, ok := reflect.Select(theApp.eventCases)
	if !ok {
		// remove evantCase
		theApp.chans = append(theApp.chans[:chosen], theApp.chans[chosen+1:]...)
		theApp.eventCases = append(theApp.eventCases[:chosen], theApp.eventCases[chosen+1:]...)
		theApp.windowList = append(theApp.windowList[:chosen], theApp.windowList[chosen+1:]...)
		remaining -= 1
	}
	panicError()
}

// WaitEvents puts the calling thread to sleep until at least one event has been
// received. Once one or more events have been recevied, it behaves as if
// PollEvents was called, i.e. the events are processed and the function then
// returns immediately. Processing events will cause the window and input
// callbacks associated with those events to be called.
//
// Since not all events are associated with callbacks, this function may return
// without a callback having been called even if you are monitoring all
// callbacks.
//
// This function may not be called from a callback.
//
// This function may only be called from the main thread.
func WaitEvents() {
	fmt.Println("not implemented")
	panicError()
}

// WaitEventsTimeout puts the calling thread to sleep until at least one event is available in the
// event queue, or until the specified timeout is reached. If one or more events are available,
// it behaves exactly like PollEvents, i.e. the events in the queue are processed and the function
// then returns immediately. Processing events will cause the window and input callbacks associated
// with those events to be called.
//
// The timeout value must be a positive finite number.
//
// Since not all events are associated with callbacks, this function may return without a callback
// having been called even if you are monitoring all callbacks.
//
// On some platforms, a window move, resize or menu operation will cause event processing to block.
// This is due to how event processing is designed on those platforms. You can use the window
// refresh callback to redraw the contents of your window when necessary during such operations.
//
// On some platforms, certain callbacks may be called outside of a call to one of the event
// processing functions.
//
// If no windows exist, this function returns immediately. For synchronization of threads in
// applications that do not create windows, use native Go primitives.
//
// Event processing is not required for joystick input to work.
func WaitEventsTimeout(timeout float64) {
	fmt.Println("not implemented")
	panicError()
}

// PostEmptyEvent posts an empty event from the current thread to the main
// thread event queue, causing WaitEvents to return.
//
// If no windows exist, this function returns immediately. For synchronization of threads in
// applications that do not create windows, use native Go primitives.
//
// This function may be called from secondary threads.
func PostEmptyEvent() {
	fmt.Println("not implemented")
	panicError()
}

// InitHint function sets hints for the next initialization of GLFW.
//
// The values you set hints to are never reset by GLFW, but they only take
// effect during initialization. Once GLFW has been initialized, any values you
// set will be ignored until the library is terminated and initialized again.
//
// Some hints are platform specific. These may be set on any platform but they
// will only affect their specific platform. Other platforms will ignore them.
// Setting these hints requires no platform specific headers or functions.
//
// This function must only be called from the main thread.
func InitHint(hint Hint, value int) {
	fmt.Println("not implemented")
}

// GetVersion retrieves the major, minor and revision numbers of the GLFW
// library. It is intended for when you are using GLFW as a shared library and
// want to ensure that you are using the minimum required version.
//
// This function may be called before Init.
func GetVersion() (major, minor, revision int) {
	return int(VersionMajor), int(VersionMinor), int(VersionRevision)
}

// GetVersionString returns a static string generated at compile-time according
// to which configuration macros were defined. This is intended for use when
// submitting bug reports, to allow developers to see which code paths are
// enabled in a binary.
//
// This function may be called before Init.
func GetVersionString() string {
	return fmt.Sprintf("%d.%d.%d Win32 WGL EGL OSMesa", VersionMajor, VersionMinor, VersionRevision)
}

// GetClipboardString returns the contents of the system clipboard, if it
// contains or is convertible to a UTF-8 encoded string.
//
// This function may only be called from the main thread.
func GetClipboardString() string {
	fmt.Println("not implemented")
	return ""
}

// SetClipboardString sets the system clipboard to the specified UTF-8 encoded
// string.
//
// This function may only be called from the main thread.
func SetClipboardString(str string) {
	fmt.Println("not implemented")
	panicError()
}
