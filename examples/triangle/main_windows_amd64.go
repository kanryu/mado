//go:build windows && amd64

// triangle combines Mado's Winit-style window driver with the local WGPU
// binding. The application never creates a native Win32 window directly:
// Mado owns the window and only exposes its native handle to the renderer.
package main

import (
	"fmt"
	"runtime"

	"github.com/gogpu/gputypes"
	"github.com/kanryu/mado"
	windowsdriver "github.com/kanryu/mado/backend/winit"
	"github.com/kanryu/wgpu-static/wgpu"
)

const (
	windowWidth  = 800
	windowHeight = 600
)

type application struct {
	factory mado.Backend
	window  mado.Window
	render  *triangleRenderer
}

func (a *application) Resumed(eventLoop *mado.ActiveEventLoop) {
	window, err := mado.NewWindow(a.factory, mado.WindowAttributes{
		Title:     "Mado + WGPU triangle",
		InnerSize: mado.Size{Width: windowWidth, Height: windowHeight},
		Resizable: true,
		HighDPI:   true,
		Visible:   true,
	})
	if err != nil {
		panic(err)
	}
	a.window = window

	renderer, err := newTriangleRenderer(window)
	if err != nil {
		panic(err)
	}
	a.render = renderer
	if err := renderer.Draw(); err != nil {
		panic(err)
	}
	fmt.Println("triangle rendered")
}

func (*application) Suspended(*mado.ActiveEventLoop) {}

func (a *application) WindowEvent(eventLoop *mado.ActiveEventLoop, _ mado.WindowID, event mado.WindowEvent) {
	switch e := event.(type) {
	case mado.WindowCloseRequested:
		if a.render != nil {
			a.render.Close()
		}
		if a.window != nil {
			_ = a.window.Close()
		}
		eventLoop.Exit()
	case mado.WindowResized:
		if a.render != nil {
			if err := a.render.Resize(e.Size); err != nil {
				fmt.Printf("resize failed: %v\n", err)
				return
			}
			if err := a.render.Draw(); err != nil {
				fmt.Printf("draw after resize failed: %v\n", err)
			}
		}
	}
}

func (*application) AboutToWait(*mado.ActiveEventLoop) {}

type triangleRenderer struct {
	instance *wgpu.Instance
	surface  *wgpu.Surface
	adapter  *wgpu.Adapter
	device   *wgpu.Device
	queue    *wgpu.Queue
	pipeline *wgpu.RenderPipeline
	format   gputypes.TextureFormat
}

func newTriangleRenderer(window mado.Window) (*triangleRenderer, error) {
	target, err := mado.SurfaceTargetFor(window)
	if err != nil {
		return nil, err
	}
	if target.Native.Kind != mado.HandleWin32HWND || target.Native.Value == 0 {
		return nil, fmt.Errorf("triangle: expected a Win32 HWND, got %q", target.Native.Kind)
	}

	if err := wgpu.Init(); err != nil {
		return nil, fmt.Errorf("wgpu init: %w", err)
	}
	wgpu.InitializeCallbacks()
	instance, err := wgpu.CreateInstance(nil)
	if err != nil {
		return nil, fmt.Errorf("create instance: %w", err)
	}
	surface, err := instance.CreateSurfaceFromWindowsHWND(target.Native.Aux, target.Native.Value)
	if err != nil {
		instance.Release()
		return nil, fmt.Errorf("create surface: %w", err)
	}
	adapter, err := instance.RequestAdapter(&wgpu.RequestAdapterOptions{CompatibleSurface: surface})
	if err != nil {
		surface.Release()
		instance.Release()
		return nil, fmt.Errorf("request adapter: %w", err)
	}
	device, err := adapter.RequestDevice(nil)
	if err != nil {
		adapter.Release()
		surface.Release()
		instance.Release()
		return nil, fmt.Errorf("request device: %w", err)
	}

	r := &triangleRenderer{instance: instance, surface: surface, adapter: adapter, device: device, queue: device.Queue()}
	caps, err := surface.GetCapabilities(adapter)
	if err != nil || len(caps.Formats) == 0 || len(caps.AlphaModes) == 0 {
		r.Close()
		if err == nil {
			err = fmt.Errorf("no supported surface formats")
		}
		return nil, fmt.Errorf("surface capabilities: %w", err)
	}
	r.format = caps.Formats[0]
	if err := r.configure(target.Size); err != nil {
		r.Close()
		return nil, err
	}

	shader, err := device.CreateShaderModuleWGSL(shaderSource)
	if err != nil {
		r.Close()
		return nil, fmt.Errorf("create shader: %w", err)
	}
	defer shader.Release()
	layout, err := device.CreatePipelineLayoutSimple(nil)
	if err != nil {
		r.Close()
		return nil, fmt.Errorf("create pipeline layout: %w", err)
	}
	defer layout.Release()
	r.pipeline, err = device.CreateRenderPipelineSimple(layout, shader, "vs_main", shader, "fs_main", r.format)
	if err != nil {
		r.Close()
		return nil, fmt.Errorf("create render pipeline: %w", err)
	}
	return r, nil
}

func (r *triangleRenderer) configure(size mado.Size) error {
	if size.Width <= 0 || size.Height <= 0 {
		return mado.ErrInvalidSize
	}
	return r.surface.Configure(r.device, &wgpu.SurfaceConfiguration{
		Format: r.format, Usage: gputypes.TextureUsageRenderAttachment,
		Width: uint32(size.Width), Height: uint32(size.Height),
		AlphaMode: gputypes.CompositeAlphaModeAuto, PresentMode: gputypes.PresentModeFifo,
	})
}

func (r *triangleRenderer) Resize(size mado.Size) error { return r.configure(size) }

func (r *triangleRenderer) Draw() error {
	frame, _, err := r.surface.GetCurrentTexture()
	if err != nil {
		return err
	}
	defer frame.Texture.Release()
	view, err := frame.Texture.CreateView(nil)
	if err != nil {
		return err
	}
	defer view.Release()
	encoder, err := r.device.CreateCommandEncoder(nil)
	if err != nil {
		return err
	}
	defer encoder.Release()
	pass, err := encoder.BeginRenderPass(&wgpu.RenderPassDescriptor{ColorAttachments: []wgpu.RenderPassColorAttachment{{
		View: view, LoadOp: gputypes.LoadOpClear, StoreOp: gputypes.StoreOpStore,
		ClearValue: wgpu.Color{R: 0.04, G: 0.08, B: 0.16, A: 1.0},
	}}})
	if err != nil {
		return err
	}
	pass.SetPipeline(r.pipeline)
	pass.Draw(3, 1, 0, 0)
	pass.End()
	commands, err := encoder.Finish()
	if err != nil {
		return err
	}
	defer commands.Release()
	if _, err := r.queue.Submit(commands); err != nil {
		return err
	}
	return r.surface.Present(frame)
}

func (r *triangleRenderer) Close() {
	if r == nil {
		return
	}
	if r.pipeline != nil {
		r.pipeline.Release()
	}
	if r.queue != nil {
		r.queue.Release()
	}
	if r.device != nil {
		r.device.Release()
	}
	if r.adapter != nil {
		r.adapter.Release()
	}
	if r.surface != nil {
		r.surface.Release()
	}
	if r.instance != nil {
		r.instance.Release()
	}
}

func main() {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	factory, err := windowsdriver.NewFactory()
	if err != nil {
		panic(err)
	}
	eventLoop, err := mado.NewEventLoop(factory)
	if err != nil {
		panic(err)
	}
	if err := eventLoop.RunApp(&application{factory: factory}); err != nil {
		panic(err)
	}
}

const shaderSource = `@vertex
fn vs_main(@builtin(vertex_index) index: u32) -> @builtin(position) vec4<f32> {
    let x = f32(i32(index) - 1);
    let y = f32(i32(index & 1u) * 2 - 1);
    return vec4<f32>(x, y, 0.0, 1.0);
}
@fragment
fn fs_main() -> @location(0) vec4<f32> {
    return vec4<f32>(1.0, 0.15, 0.05, 1.0);
}`
