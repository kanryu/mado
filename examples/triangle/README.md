# Triangle

This example combines Mado's `windowsdriver` window driver with
`github.com/kanryu/wgpu-static`.
Mado creates the window, exposes its Win32 HWND through `SurfaceTargetFor`, and
the WGPU renderer creates the surface and draws a triangle into it.

This sample currently targets Windows amd64 and uses the local `../wgpu-static`
module replacement from the repository's `go.mod`.

```powershell
go run ./examples/triangle
```
