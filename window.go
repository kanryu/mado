package mado

// NewWindow creates a native window through the supplied backend factory.
func NewWindow(factory Backend, attributes WindowAttributes) (Window, error) {
	if factory == nil {
		return nil, ErrUnsupported
	}
	return factory.CreateWindow(attributes.withDefaults().windowOptions())
}
