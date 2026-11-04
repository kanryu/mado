# Mado

Mado is a Go meta-framework for Windows applications. It provides a common,
Winit-inspired application and window API while keeping native window drivers
and graphics APIs separate.

```text
Application
    │
    ├── Mado window API
    │       ├── SDL3 window driver
    │       └── Winit window driver
    │
    └── Graphics integration
            ├── WGPU
            └── SDL_GPU
```

Mado does not try to make WGPU and SDL_GPU expose the same command-buffer,
shader, pipeline, or resource APIs. Those APIs remain native to each graphics
library.

## Architecture

### Core package

The root `github.com/kanryu/mado` package provides the backend-independent API:

- `Backend` and `Window`
- `EventLoop`, `ApplicationHandler`, and `ActiveEventLoop`
- `WindowAttributes` and `WindowEvent`
- `NativeHandle` and `SurfaceTarget`
- `Session` and `Input`

The public application API follows the Winit model. A driver factory is passed
to Mado explicitly; there is no global backend registration or hidden `init`
state.

### Window drivers

Window drivers own native window creation, native event polling, and conversion
of platform events into Mado events.

Available drivers are:

- `github.com/kanryu/mado/backend/sdl3`
- `github.com/kanryu/mado/backend/winit`

Both drivers expose the same factory entry point:

```go
factory, err := windowsdriver.NewFactory()
```

An application can use an import alias so switching drivers only requires one
import change:

```go
import windowsdriver "github.com/kanryu/mado/backend/sdl3"

// Or:
// import windowsdriver "github.com/kanryu/mado/backend/winit"
```

The rest of the application continues to use the Mado API.

### Window lifecycle

The portable event-loop setup is:

```go
factory, err := windowsdriver.NewFactory()
if err != nil {
    panic(err)
}

eventLoop, err := mado.NewEventLoop(factory)
if err != nil {
    panic(err)
}

eventLoop.RunApp(&application{})
```

For Winit-style lifecycle code, create the Window from `Resumed` through the
active event loop:

```go
func (a *application) Resumed(loop *mado.ActiveEventLoop) {
    window, err := loop.CreateWindow(mado.WindowAttributes{
        Title:     "Mado application",
        InnerSize: mado.Size{Width: 800, Height: 600},
        Resizable: true,
        Visible:   true,
    })
    if err != nil {
        panic(err)
    }
    a.window = window
}
```

`mado.NewWindow(factory, attributes)` is also available when an application
needs to create the Window explicitly before starting the event loop.

### Events

Applications receive Winit-shaped events through `WindowEvent`, including close
requests, resizing, redraw requests, scale-factor changes, keyboard input,
mouse input, text input, focus, occlusion, and minimized/maximized/restored
state changes.

Window drivers may coalesce noisy native resize notifications. During a resize
drag, applications should treat a resize notification as the latest known size
instead of performing expensive work for every native message.

### Input sessions

Input objects are created through the Window's shared input facade:

```go
input := window.Input()
mouse := input.Mouse()
keyboard := input.Keyboard()
joystick := input.Joystick()
```

These objects share the same Mado session created when the Window is created.
Device capabilities remain driver-specific where necessary.

## Graphics integration

Graphics APIs are kept outside the core package. A renderer can obtain the
native information required to create a graphics Surface with:

```go
target, err := mado.SurfaceTargetFor(window)
if err != nil {
    panic(err)
}
```

`SurfaceTarget.Native` contains an opaque platform handle. On Windows, the
Window drivers provide the HWND; the Winit driver also provides the associated
HINSTANCE through `NativeHandle.Aux` when required by a graphics binding.

### WGPU

The WGPU triangle sample uses this path:

```text
Mado Winit window driver
    → Win32 HWND/HINSTANCE
    → WGPU Surface
    → Adapter / Device / Queue
    → Render Pipeline
    → Triangle
```

The sample reconfigures the WGPU Surface when the Window is resized.

### SDL_GPU

The SDL triangle sample uses SDL3 for both Window and graphics operations:

```text
Mado SDL3 window driver
    → SDL_Window
    → SDL_GPUDevice
    → SDL_GPUShader
    → SDL_GPUGraphicsPipeline
    → SDL_GPURenderPass
    → Triangle
```

This demonstrates that Mado can still provide the application and Window
lifecycle layer when native Window and GPU operations are both supplied by
SDL3.

## Examples

### Simple application

`examples/simple` demonstrates the common Mado Window and event API. Its import
block contains both Window driver choices and explains how to switch between
them by commenting and uncommenting one import.

```powershell
go run ./examples/simple
```

### WGPU triangle

`examples/triangle` uses the Winit Window driver and WGPU to create a Win32
Surface and render a triangle.

```powershell
go run ./examples/triangle
```

### SDL_GPU triangle

`examples/triangle_sdl` uses SDL3 for the Window driver and SDL_GPU for device,
shader, pipeline, command-buffer, and render-pass operations. It also keeps a
small line triangle in the upper-right corner as a line-list example.

```powershell
go run ./examples/triangle_sdl
```

The graphics examples currently target Windows amd64 because their native
bindings and shader assets are Windows-specific.

## Related projects

- [go-sdl3](https://github.com/kanryu/go-sdl3) — Go bindings for SDL3
- [go-winit](https://github.com/kanryu/go-winit) — Go bindings for Winit
- [wgpu-static](https://github.com/kanryu/wgpu-static) — WGPU binding used by the WGPU sample

## Development status

The Window driver boundary, common event model, input session model,
native-handle handoff, WGPU integration, and SDL_GPU integration are
established. Higher-level rendering helpers, continuous redraw scheduling, and
additional input-device implementations can be added without changing the
fundamental Window driver architecture.
