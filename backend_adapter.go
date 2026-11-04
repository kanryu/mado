package mado

// BackendAdapter is useful when a backend package wants to expose a thin
// package-specific constructor while delegating the actual native operations
// to a binding such as go-sdl3 or go-winit.
type BackendAdapter struct {
	Kind    BackendKind
	Factory interface {
		CreateWindow(WindowOptions) (Window, error)
		Run(func(Event) error) error
		Quit()
	}
}

func (b BackendAdapter) Name() BackendKind { return b.Kind }
func (b BackendAdapter) CreateWindow(o WindowOptions) (Window, error) {
	if b.Factory == nil {
		return nil, ErrUnsupported
	}
	return b.Factory.CreateWindow(o)
}
func (b BackendAdapter) Run(handle func(Event) error) error {
	if b.Factory == nil {
		return ErrUnsupported
	}
	return b.Factory.Run(handle)
}
func (b BackendAdapter) Quit() {
	if b.Factory != nil {
		b.Factory.Quit()
	}
}
