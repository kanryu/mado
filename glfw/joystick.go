package glfw

// JoystickCallback is the joystick configuration callback.
type JoystickCallback func(joy Joystick, event PeripheralEvent)

// SetJoystickCallback sets the joystick configuration callback, or removes the
// currently set callback. This is called when a joystick is connected to or
// disconnected from the system.
func SetJoystickCallback(cbfun JoystickCallback) (previous JoystickCallback) {
	previous = theApp.fJoystickHolder
	theApp.fJoystickHolder = cbfun
	panicError()
	return previous
}
