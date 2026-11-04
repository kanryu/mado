// simple demonstrates the Winit-inspired Mado application lifecycle.
//
// The example uses the SDL3 window driver. The application handler is
// identical to the Winit version; only the driver import differs.
package main

import (
	"fmt"

	"github.com/kanryu/mado"
	// SDL3 driver: owns SDL initialization, native window creation, and SDL
	// event polling. Use this when SDL input/devices or SDL GPU are needed.
	windowsdriver "github.com/kanryu/mado/backend/sdl3"
	// Winit driver: comment out the SDL3 import above and uncomment this line
	// to use Winit instead. The rest of the application remains unchanged
	// because both drivers implement the same Mado APIs.
	// windowsdriver "github.com/kanryu/mado/backend/winit"
)

type application struct {
	window mado.Window
}

func (a *application) Resumed(*mado.ActiveEventLoop) {
	// All input objects come from the same session created for this window.
	input := a.window.Input()
	_ = input.Mouse()
	_ = input.Keyboard()
	_ = input.Joystick()
	fmt.Printf("window ready: %q (%dx%d)\n", a.window.Title(), a.window.Size().Width, a.window.Size().Height)
}

func (*application) Suspended(*mado.ActiveEventLoop) {}

func (a *application) WindowEvent(eventLoop *mado.ActiveEventLoop, id mado.WindowID, event mado.WindowEvent) {
	switch e := event.(type) {
	case mado.WindowCloseRequested:
		fmt.Println("Window close requested")
		_ = a.window.Close()
		eventLoop.Exit()
	case mado.WindowResized:
		fmt.Printf("Window resized to: (%d x %d)\n", e.Size.Width, e.Size.Height)
	case mado.WindowCursorMoved:
		fmt.Printf("Cursor moved to: (%d @ %d)\n", e.Position.X, e.Position.Y)
	case mado.WindowReceivedCharacter:
		fmt.Printf("Received character: %q\n", e.Character)
	case mado.WindowKeyboardInput:
		state := "Released"
		if e.Down {
			state = "Pressed"
		}
		fmt.Printf("Keyboard %s: key=%q repeat=%t\n", state, e.Key, e.Repeat)
	case mado.WindowScaleFactorChanged:
		fmt.Printf("Scale factor changed to: %f\n", e.Scale)
	case mado.WindowOccluded:
		fmt.Printf("Window occluded (minimized/hidden): %t\n", e.Occluded)
	case mado.WindowMinimizedEvent:
		fmt.Println("Window minimized")
	case mado.WindowMaximizedEvent:
		fmt.Println("Window maximized")
	case mado.WindowRestoredEvent:
		fmt.Println("Window restored")
	case mado.WindowFocused:
		fmt.Printf("Window focused: %t\n", e.Focused)
	case mado.WindowMouseInput:
		state := "Released"
		if e.Down {
			state = "Pressed"
		}
		fmt.Printf("Mouse button %d %s at (%d @ %d)\n", e.Button, state, e.Position.X, e.Position.Y)
	}
}

func (*application) AboutToWait(*mado.ActiveEventLoop) {}

func main() {
	// The selected driver owns native initialization and event polling. Only
	// the import above decides whether this application uses SDL3 or Winit.
	factory, err := windowsdriver.NewFactory()
	if err != nil {
		panic(err)
	}
	window, err := mado.NewWindow(factory, mado.WindowAttributes{
		Title:     "Mado simple application",
		InnerSize: mado.Size{Width: 800, Height: 600},
		Resizable: true,
		HighDPI:   true,
		Visible:   true,
	})
	if err != nil {
		panic(err)
	}
	eventLoop, err := mado.NewEventLoop(factory)
	if err != nil {
		panic(err)
	}
	if err := eventLoop.RunApp(&application{window: window}); err != nil {
		panic(err)
	}
}
