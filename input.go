package mado

import "sync"

// Session is shared by a window and all input devices created from it.
type Session struct {
	kind     BackendKind
	mu       sync.Mutex
	mouse    *Mouse
	keyboard *Keyboard
	joystick *Joystick
	input    *Input
}

func NewSession(kind BackendKind) *Session { return &Session{kind: kind} }
func (s *Session) Backend() BackendKind {
	if s == nil {
		return BackendAuto
	}
	return s.kind
}

type Mouse struct{ session *Session }
type Keyboard struct{ session *Session }
type Joystick struct{ session *Session }
type Input struct{ session *Session }

func (s *Session) Input() *Input {
	if s == nil {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.input == nil {
		s.input = &Input{session: s}
	}
	return s.input
}
func (i *Input) Backend() BackendKind {
	if i == nil || i.session == nil {
		return BackendAuto
	}
	return i.session.Backend()
}
func (i *Input) Mouse() *Mouse {
	if i == nil {
		return nil
	}
	return i.session.Mouse()
}
func (i *Input) Keyboard() *Keyboard {
	if i == nil {
		return nil
	}
	return i.session.Keyboard()
}
func (i *Input) Joystick() *Joystick {
	if i == nil {
		return nil
	}
	return i.session.Joystick()
}

func (s *Session) Mouse() *Mouse {
	if s == nil {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.mouse == nil {
		s.mouse = &Mouse{session: s}
	}
	return s.mouse
}
func (s *Session) Keyboard() *Keyboard {
	if s == nil {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.keyboard == nil {
		s.keyboard = &Keyboard{session: s}
	}
	return s.keyboard
}
func (s *Session) Joystick() *Joystick {
	if s == nil {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.joystick == nil {
		s.joystick = &Joystick{session: s}
	}
	return s.joystick
}

type SessionWindow interface {
	Input() *Input
}
