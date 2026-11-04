package mado

// SurfaceTarget is the backend-neutral information required to create a WGPU
// surface. A WGPU adapter should consume this value and keep the native window
// alive for at least as long as the surface.
type SurfaceTarget struct {
	Window Window
	Native NativeHandle
	Size   Size
}

func SurfaceTargetFor(window Window) (SurfaceTarget, error) {
	if window == nil {
		return SurfaceTarget{}, ErrClosed
	}
	size := window.Size()
	if size.Width <= 0 || size.Height <= 0 {
		return SurfaceTarget{}, ErrInvalidSize
	}
	return SurfaceTarget{Window: window, Native: window.NativeHandle(), Size: size}, nil
}

// WGPUBridge keeps the core package independent of a particular Go WGPU
// binding. The concrete gogpu/wgpu or go-webgpu/webgpu integration implements
// CreateSurface using SurfaceTarget.Native and calls Resize when Resized events
// arrive.
type WGPUBridge interface {
	CreateSurface(SurfaceTarget) error
	Resize(Size) error
	Present() error
	Destroy() error
}
