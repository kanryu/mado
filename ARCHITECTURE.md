# What is Mado's architecture?

The question of what Mado's architecture should be is an essential question.

At the initial stage, the functionality provided by the Mado package is almost the same as that of gioui, with the exception of some APIs that were unavoidably added to support glfw compatibility mode.

Although I think that the API of the Mado package should eventually mature and be used by many external products, for the time being we will prioritize support for the glfw layer as a stub.


## Package structure of mado framework

- mado
 - Provides interfaces and constants for basic concepts such as applications, windows, ime context, and renderers defined by the mado framework.

- mado/io or base
  - Basic OS-independent objects such as events and coordinate calculations
  - Should I change the name to something else?

- mado/driver
  - Implementations and specific structs for various OS and Window Systems are defined.
  - Implementation structs such as Window and App, on each OS
  - MS-Windows, Mac OS, Linux X11, Linux Wayland, Android, iOS, Web(WebAssembly on browsers)

- mado/renderer
  - Various hardware renderers are implemented here.
  - OpenGL, OpenGL ES, Vulkan, Metal, Direct3D11
  - You can register a new externally defined renderer with mado by calling the registration function (e.g. AddRenderer())
    - Additional renderers: e.g. skia, cairo, GDI+, MESA, CPU

## Deprecated packages

- mado/app
  - It is a core package in gioui, but this is because gio is a monolithic library, and it is not necessary for mado. The programs in this will eventually be split into multiple separate packages and disappear.
- mado/f32
  - will be moved to mado/io
- mado/gesture
  - will be moved to mado/io
- mado/font
  - will be moved to mado/io
- mado/widget
- mado/layout
- mado/gpu
- mado/op
- mado/text

## What should be discussed 
- mado/font
  - MS-Windows, Mac, and iOS have their own text rendering
  - I think go-text and freetype are usually used, but I think there are other options as well.
- mado/camera, security, privilige
  - How do you handle OS and hardware-specific events and requests, such as push notifications, cameras, security checks, permission requests and revocations, etc.?

