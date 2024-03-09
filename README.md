# About

[Mado](https://github.com/kanryu/mado) is an window/framebuffer client for Windows/Mac/Linux.

Based a lot on [Gio UI](https://github.com/gioui/gio).

Currently, WIP.

## Purpose

- Intended to be a base layer for GUI toolkits and game engines.
- Transparent window control for various GUI desktops
  - Support for MS-Windows, Mac OS, X11 and Wayland
- Selectable hardware renderers
  - Support for OpenGL(ES), DirectX 11, Vulkan, Metal, CPU
- Transparent IME Text input support
- Provide glfw compatible API (a valid alternative from go-gl/glfw)
- Reduce cgo programs as much as possible and enable development in Go language

## Means

The word 'Mado' is the Japanese word for '窓(まど)' and the meaning is Window.

## Motivation

- I would like to see more GUI toolkits that support CJK Languages, e.g. Japanese.
- Solve the problem that IME text input is not supported in go-gl/glfw, which is widely used in Go language.

## Aims for

- Achieve the purpose items
- Development by communities rather than individuals
- Provision of parts useful for GUI toolkit or game engines development 
  - Events, OS Thread Managements, Common Drawing API

## Not aims for

- Completed GUI toolkit and game engine

## FAQ

- OpenGL examples can'nt run on MacOS
  - currently, metal context is selected as default. so run it
  - `go main -tags=nometal main.go`

## License

MIT / The UNLICENSE

