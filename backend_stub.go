package mado

// StubWindow is useful for backend tests and for renderer integration tests.
// Real backends should implement Window directly.
type StubWindow struct {
	IDValue      WindowID
	TitleValue   string
	SizeValue    Size
	Handle       NativeHandle
	Closed       bool
	SessionValue *Session
}

func (w *StubWindow) ID() WindowID  { return w.IDValue }
func (w *StubWindow) Title() string { return w.TitleValue }
func (w *StubWindow) Size() Size    { return w.SizeValue }
func (w *StubWindow) SetSize(s Size) error {
	if s.Width <= 0 || s.Height <= 0 {
		return ErrInvalidSize
	}
	w.SizeValue = s
	return nil
}
func (w *StubWindow) SetTitle(s string) error    { w.TitleValue = s; return nil }
func (w *StubWindow) NativeHandle() NativeHandle { return w.Handle }
func (w *StubWindow) Close() error               { w.Closed = true; return nil }
func (w *StubWindow) Input() *Input {
	if w.SessionValue == nil {
		return nil
	}
	return w.SessionValue.Input()
}
