// Package sdl3 exposes the SDL3 backend entry point for mado.
//
// The binding-specific implementation is injected as a Factory. This keeps
// the root mado package independent from the selected SDL3 Go binding and
// allows kanryu/go-sdl3 to evolve without changing the common API.
package sdl3

import "github.com/kanryu/mado"

type BackendFactory interface {
	CreateWindow(mado.WindowOptions) (mado.Window, error)
	Run(func(mado.Event) error) error
	Quit()
}

func New(factory BackendFactory) mado.Backend {
	return mado.BackendAdapter{Kind: mado.BackendSDL3, Factory: factory}
}
