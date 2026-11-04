// triangle_sdl combines Mado's SDL3 Window driver with SDL_GPU.
package main

import (
	_ "embed"
	"fmt"
	"unsafe"

	"github.com/kanryu/go-sdl3/sdl"
	"github.com/kanryu/mado"
	windowsdriver "github.com/kanryu/mado/backend/sdl3"
)

//go:embed triangle.vert.dxil
var vertexShader []byte

//go:embed triangle.frag.dxil
var fragmentShader []byte

const shaderEntryPoint = "main"

type application struct {
	window       mado.Window
	sdlWindow    *sdl.Window
	device       sdl.GPUDevice
	shaderVS     sdl.GPUShader
	shaderFS     sdl.GPUShader
	pipeline     sdl.GPUGraphicsPipeline
	linePipeline sdl.GPUGraphicsPipeline
}

func (a *application) Resumed(*mado.ActiveEventLoop) {
	a.sdlWindow = sdl.WrapWindowHandle(a.window.NativeHandle().Value)
	a.device = sdl.CreateGPUDevice(sdl.GPUShaderFormatDXIL, false, "direct3d12")
	if a.device == nil || !sdl.ClaimWindowForGPUDevice(a.device, a.sdlWindow) {
		panic("triangle_sdl: failed to initialize SDL GPU device")
	}
	if !sdl.SetGPUSwapchainParameters(a.device, a.sdlWindow, sdl.GPUSwapchainCompositionSDR, sdl.GPUPresentModeVSync) {
		panic("triangle_sdl: failed to configure SDL GPU swapchain")
	}

	format := sdl.GPUTextureFormat(sdl.GetGPUSwapchainTextureFormat(a.device, a.sdlWindow))
	entry := append([]byte(shaderEntryPoint), 0)
	a.shaderVS = sdl.CreateGPUShader(a.device, &sdl.GPUShaderCreateInfo{CodeSize: uintptr(len(vertexShader)), Code: unsafe.Pointer(&vertexShader[0]), Entrypoint: &entry[0], Format: sdl.GPUShaderFormatDXIL, Stage: sdl.GPUShaderStageVertex})
	a.shaderFS = sdl.CreateGPUShader(a.device, &sdl.GPUShaderCreateInfo{CodeSize: uintptr(len(fragmentShader)), Code: unsafe.Pointer(&fragmentShader[0]), Entrypoint: &entry[0], Format: sdl.GPUShaderFormatDXIL, Stage: sdl.GPUShaderStageFragment})
	if a.shaderVS == nil || a.shaderFS == nil {
		panic("triangle_sdl: failed to create SDL GPU shaders")
	}
	a.pipeline = createPipeline(a.device, a.shaderVS, a.shaderFS, format, sdl.GPUPrimitiveTypeTriangleList)
	a.linePipeline = createPipeline(a.device, a.shaderVS, a.shaderFS, format, sdl.GPUPrimitiveTypeLineList)
	if a.pipeline == nil || a.linePipeline == nil {
		panic("triangle_sdl: failed to create SDL GPU pipelines")
	}
	if err := a.draw(); err != nil {
		panic(err)
	}
	fmt.Printf("SDL GPU triangle rendered with %s\n", sdl.GetGPUDeviceDriver(a.device))
}

func createPipeline(device sdl.GPUDevice, vertex, fragment sdl.GPUShader, format sdl.GPUTextureFormat, primitive sdl.GPUPrimitiveType) sdl.GPUGraphicsPipeline {
	return sdl.CreateGPUGraphicsPipeline(device, &sdl.GPUGraphicsPipelineCreateInfo{
		VertexShader: vertex, FragmentShader: fragment, PrimitiveType: primitive,
		RasterizerState:  sdl.GPURasterizerState{FillMode: sdl.GPUFillModeFill, CullMode: sdl.GPUCullModeNone, FrontFace: sdl.GPUFrontFaceCounterClockwise, EnableDepthClip: true},
		MultisampleState: sdl.GPUMultisampleState{SampleCount: 1},
		TargetInfo:       sdl.GPUGraphicsPipelineTargetInfo{ColorTargetDescriptions: &sdl.GPUColorTargetDescription{Format: format}, NumColorTargets: 1},
	})
}

func (a *application) draw() error {
	commandBuffer := sdl.AcquireGPUCommandBuffer(a.device)
	if commandBuffer == nil {
		return fmt.Errorf("acquire SDL GPU command buffer failed: %s", sdl.GetError())
	}
	texture, _, _, ok := sdl.AcquireGPUSwapchainTexture(commandBuffer, a.sdlWindow)
	if !ok || texture == nil {
		sdl.CancelGPUCommandBuffer(commandBuffer)
		return nil
	}
	pass := sdl.BeginGPURenderPass(commandBuffer, []sdl.GPUColorTargetInfo{{Texture: texture, ClearColor: sdl.FColor{R: 0.04, G: 0.08, B: 0.16, A: 1}, LoadOp: sdl.GPULoadOpClear, StoreOp: sdl.GPUStoreOpStore}})
	if pass == nil {
		sdl.CancelGPUCommandBuffer(commandBuffer)
		return fmt.Errorf("begin SDL GPU render pass failed: %s", sdl.GetError())
	}
	sdl.BindGPUGraphicsPipeline(pass, a.pipeline)
	sdl.DrawGPUPrimitives(pass, 3, 1, 0, 0)
	sdl.BindGPUGraphicsPipeline(pass, a.linePipeline)
	sdl.DrawGPUPrimitives(pass, 6, 1, 3, 0)
	sdl.EndGPURenderPass(pass)
	if !sdl.SubmitGPUCommandBuffer(commandBuffer) {
		return fmt.Errorf("submit SDL GPU command buffer failed: %s", sdl.GetError())
	}
	return nil
}

func (a *application) WindowEvent(eventLoop *mado.ActiveEventLoop, _ mado.WindowID, event mado.WindowEvent) {
	switch event.(type) {
	case mado.WindowCloseRequested:
		if a.device != nil && a.sdlWindow != nil {
			sdl.ReleaseWindowFromGPUDevice(a.device, a.sdlWindow)
		}
		if a.device != nil {
			if a.pipeline != nil {
				sdl.ReleaseGPUGraphicsPipeline(a.device, a.pipeline)
			}
			if a.linePipeline != nil {
				sdl.ReleaseGPUGraphicsPipeline(a.device, a.linePipeline)
			}
			if a.shaderVS != nil {
				sdl.ReleaseGPUShader(a.device, a.shaderVS)
			}
			if a.shaderFS != nil {
				sdl.ReleaseGPUShader(a.device, a.shaderFS)
			}
			sdl.DestroyGPUDevice(a.device)
		}
		_ = a.window.Close()
		eventLoop.Exit()
	case mado.WindowResized:
		if err := a.draw(); err != nil {
			fmt.Printf("SDL GPU draw after resize failed: %v\n", err)
		}
	}
}

func (*application) Suspended(*mado.ActiveEventLoop)   {}
func (*application) AboutToWait(*mado.ActiveEventLoop) {}

func main() {
	factory, err := windowsdriver.NewFactory()
	if err != nil {
		panic(err)
	}
	window, err := mado.NewWindow(factory, mado.WindowAttributes{Title: "Mado + SDL_GPU triangle", InnerSize: mado.Size{Width: 800, Height: 600}, Resizable: true, HighDPI: true, Visible: true})
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
