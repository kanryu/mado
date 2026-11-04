// Package winit exposes the Winit backend entry point for mado.
//
// The binding-specific implementation is injected as a Factory. This keeps
// Winit event-loop ownership inside the backend package and leaves application
// code using only the mado API.
package winit

import "github.com/kanryu/mado"

type Factory interface {
	CreateWindow(mado.WindowOptions) (mado.Window, error)
	Run(func(mado.Event) error) error
	Quit()
}

func New(factory Factory) mado.Backend {
	return mado.BackendAdapter{Kind: mado.BackendWinit, Factory: factory}
}
