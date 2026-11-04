module github.com/kanryu/mado

go 1.27

require (
	github.com/gogpu/gputypes v0.5.1
	github.com/kanryu/go-sdl3 v0.0.0-20261104063602-d32f4f08042b
	github.com/kanryu/go-winit v0.0.0
	github.com/kanryu/wgpu-static v0.0.0
)

replace github.com/kanryu/go-sdl3 => ../go-sdl3

replace github.com/kanryu/go-winit => ../go-winit

replace github.com/kanryu/wgpu-static => ../wgpu-static
