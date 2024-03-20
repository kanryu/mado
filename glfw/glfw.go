package glfw

import (
	"fmt"
	"sync"

	"github.com/kanryu/mado"
	"github.com/kanryu/mado/app"
	"github.com/kanryu/mado/io/event"
	"github.com/kanryu/mado/io/key"
	"github.com/kanryu/mado/io/pointer"
	"github.com/kanryu/mado/op"
)

// Version constants.
const (
	VersionMajor    = 3 // This is incremented when the API is changed in non-compatible ways.
	VersionMinor    = 3 // This is incremented when features are added to the API but it remains backward-compatible.
	VersionRevision = 9 // This is incremented when a bug fix release is made that does not contain any API changes.
)

var theApp *Application

// Application keeps track of all the windows and global state.
type Application struct {
	// Context is used to broadcast application shutdown.
	// Context    context.Context
	// Stop       context.CancelFunc
	Ctx        mado.Context
	MainWindow *Window
	// // Shutdown shuts down all windows.
	// Shutdown func()
	// active keeps track the open windows, such that application
	// can shut down, when all of them are closed.
	Active sync.WaitGroup

	fJoystickHolder func(joy Joystick, event PeripheralEvent)
}

// func NewApplication(ctx context.Context, stop context.CancelFunc) *Application {
func NewApplication() *Application {
	// ctx, cancel := context.WithCancel(ctx)
	return &Application{
		// Context:         ctx,
		// Stop:            stop,
		// Shutdown:        cancel,
		fJoystickHolder: func(joy Joystick, event PeripheralEvent) {},
	}
}

// appendWindow add a window to your application. Count Active and watch until window is destroyed
func (a *Application) appendWindow(w *Window) {
	windows.put(w)
	a.Active.Add(1)
	a.MainWindow = w

	go func() {
		defer a.Active.Done()
		a.run(w)
	}()
	PollEvents()
}

func (a *Application) run(w *Window) {
	tag := new(int)
	var ops op.Ops
	for {
		e := w.data.NextEvent()
		switch e2 := e.(type) {
		case mado.DestroyEvent:
			w.shouldClose = true
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
					}
				}
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
}

// Wait waits for all windows to close.
func (a *Application) Wait() {
	a.Active.Wait()
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
	mado.GlfwConfig.Enable = true
	mado.EnablePollEvents()
	app.Main()

	// ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	// theApp = NewApplication(ctx, stop)
	theApp = NewApplication()
	mado.GlfwConfig.Initialized = true
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
	//theApp.Stop()
	flushErrors()
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
	mado.PollEvents()
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

// GetTime returns the value of the GLFW timer. Unless the timer has been set
// using SetTime, the timer measures time elapsed since GLFW was initialized.
//
// The resolution of the timer is system dependent, but is usually on the order
// of a few micro- or nanoseconds. It uses the highest-resolution monotonic time
// source on each supported platform.
func GetTime() float64 {
	tm := mado.GetTimerValue()
	freq := mado.GetTimerFrequency()
	ret := float64(tm) / float64(freq)
	panicError()
	return ret
}
