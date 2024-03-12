package glfw

import "image"

// Joystick corresponds to a joystick.
type Joystick int

// Joystick IDs.
const (
	Joystick1    Joystick = 0
	Joystick2    Joystick = 1
	Joystick3    Joystick = 2
	Joystick4    Joystick = 3
	Joystick5    Joystick = 4
	Joystick6    Joystick = 5
	Joystick7    Joystick = 6
	Joystick8    Joystick = 7
	Joystick9    Joystick = 8
	Joystick10   Joystick = 9
	Joystick11   Joystick = 10
	Joystick12   Joystick = 11
	Joystick13   Joystick = 12
	Joystick14   Joystick = 13
	Joystick15   Joystick = 14
	Joystick16   Joystick = 15
	JoystickLast Joystick = Joystick16
)

// JoystickHatState corresponds to joystick hat states.
type JoystickHatState int

// Joystick Hat State IDs.
const (
	HatCentered  JoystickHatState = 0
	HatUp        JoystickHatState = 1
	HatRight     JoystickHatState = 2
	HatDown      JoystickHatState = 4
	HatLeft      JoystickHatState = 8
	HatRightUp   JoystickHatState = (HatRight | HatUp)
	HatRightDown JoystickHatState = (HatRight | HatDown)
	HatLeftUp    JoystickHatState = (HatLeft | HatUp)
	HatLeftDown  JoystickHatState = (HatLeft | HatDown)
)

// GamepadAxis corresponds to a gamepad axis.
type GamepadAxis int

// Gamepad axis IDs.
const (
	AxisLeftX        GamepadAxis = 0
	AxisLeftY        GamepadAxis = 1
	AxisRightX       GamepadAxis = 2
	AxisRightY       GamepadAxis = 3
	AxisLeftTrigger  GamepadAxis = 4
	AxisRightTrigger GamepadAxis = 5
	AxisLast         GamepadAxis = AxisRightTrigger
)

// GamepadButton corresponds to a gamepad button.
type GamepadButton int

// Gamepad button IDs.
const (
	ButtonA           GamepadButton = 0
	ButtonB           GamepadButton = 1
	ButtonX           GamepadButton = 2
	ButtonY           GamepadButton = 3
	ButtonLeftBumper  GamepadButton = 4
	ButtonRightBumper GamepadButton = 5
	ButtonBack        GamepadButton = 6
	ButtonStart       GamepadButton = 7
	ButtonGuide       GamepadButton = 8
	ButtonLeftThumb   GamepadButton = 9
	ButtonRightThumb  GamepadButton = 10
	ButtonDpadUp      GamepadButton = 11
	ButtonDpadRight   GamepadButton = 12
	ButtonDpadDown    GamepadButton = 13
	ButtonDpadLeft    GamepadButton = 14
	ButtonLast        GamepadButton = ButtonDpadLeft
	ButtonCross       GamepadButton = ButtonA
	ButtonCircle      GamepadButton = ButtonB
	ButtonSquare      GamepadButton = ButtonX
	ButtonTriangle    GamepadButton = ButtonY
)

// GamepadState describes the input state of a gamepad.
type GamepadState struct {
	Buttons [15]Action
	Axes    [6]float32
}

// Key corresponds to a keyboard key.
type Key int

// These key codes are inspired by the USB HID Usage Tables v1.12 (p. 53-60),
// but re-arranged to map to 7-bit ASCII for printable keys (function keys are
// put in the 256+ range).
const (
	KeyUnknown      Key = -1
	KeySpace        Key = 32
	KeyApostrophe   Key = 39 /* ' */
	KeyComma        Key = 44 /* , */
	KeyMinus        Key = 45 /* - */
	KeyPeriod       Key = 46 /* . */
	KeySlash        Key = 47 /* / */
	Key0            Key = 48
	Key1            Key = 49
	Key2            Key = 50
	Key3            Key = 51
	Key4            Key = 52
	Key5            Key = 53
	Key6            Key = 54
	Key7            Key = 55
	Key8            Key = 56
	Key9            Key = 57
	KeySemicolon    Key = 59 /* ; */
	KeyEqual        Key = 61 /* = */
	KeyA            Key = 65
	KeyB            Key = 66
	KeyC            Key = 67
	KeyD            Key = 68
	KeyE            Key = 69
	KeyF            Key = 70
	KeyG            Key = 71
	KeyH            Key = 72
	KeyI            Key = 73
	KeyJ            Key = 74
	KeyK            Key = 75
	KeyL            Key = 76
	KeyM            Key = 77
	KeyN            Key = 78
	KeyO            Key = 79
	KeyP            Key = 80
	KeyQ            Key = 81
	KeyR            Key = 82
	KeyS            Key = 83
	KeyT            Key = 84
	KeyU            Key = 85
	KeyV            Key = 86
	KeyW            Key = 87
	KeyX            Key = 88
	KeyY            Key = 89
	KeyZ            Key = 90
	KeyLeftBracket  Key = 91  /* [ */
	KeyBackslash    Key = 92  /* \ */
	KeyRightBracket Key = 93  /* ] */
	KeyGraveAccent  Key = 96  /* ` */
	KeyWorld1       Key = 161 /* non-US #1 */
	KeyWorld2       Key = 162 /* non-US #2 */
	KeyEscape       Key = 256
	KeyEnter        Key = 257
	KeyTab          Key = 258
	KeyBackspace    Key = 259
	KeyInsert       Key = 260
	KeyDelete       Key = 261
	KeyRight        Key = 262
	KeyLeft         Key = 263
	KeyDown         Key = 264
	KeyUp           Key = 265
	KeyPageUp       Key = 266
	KeyPageDown     Key = 267
	KeyHome         Key = 268
	KeyEnd          Key = 269
	KeyCapsLock     Key = 280
	KeyScrollLock   Key = 281
	KeyNumLock      Key = 282
	KeyPrintScreen  Key = 283
	KeyPause        Key = 284
	KeyF1           Key = 290
	KeyF2           Key = 291
	KeyF3           Key = 292
	KeyF4           Key = 293
	KeyF5           Key = 294
	KeyF6           Key = 295
	KeyF7           Key = 296
	KeyF8           Key = 297
	KeyF9           Key = 298
	KeyF10          Key = 299
	KeyF11          Key = 300
	KeyF12          Key = 301
	KeyF13          Key = 302
	KeyF14          Key = 303
	KeyF15          Key = 304
	KeyF16          Key = 305
	KeyF17          Key = 306
	KeyF18          Key = 307
	KeyF19          Key = 308
	KeyF20          Key = 309
	KeyF21          Key = 310
	KeyF22          Key = 311
	KeyF23          Key = 312
	KeyF24          Key = 313
	KeyF25          Key = 314
	KeyKP0          Key = 320
	KeyKP1          Key = 321
	KeyKP2          Key = 322
	KeyKP3          Key = 323
	KeyKP4          Key = 324
	KeyKP5          Key = 325
	KeyKP6          Key = 326
	KeyKP7          Key = 327
	KeyKP8          Key = 328
	KeyKP9          Key = 329
	KeyKPDecimal    Key = 330
	KeyKPDivide     Key = 331
	KeyKPMultiply   Key = 332
	KeyKPSubtract   Key = 333
	KeyKPAdd        Key = 334
	KeyKPEnter      Key = 335
	KeyKPEqual      Key = 336
	KeyLeftShift    Key = 340
	KeyLeftControl  Key = 341
	KeyLeftAlt      Key = 342
	KeyLeftSuper    Key = 343
	KeyRightShift   Key = 344
	KeyRightControl Key = 345
	KeyRightAlt     Key = 346
	KeyRightSuper   Key = 347
	KeyMenu         Key = 348
	KeyLast         Key = KeyMenu
)

// ModifierKey corresponds to a modifier key.
type ModifierKey int

// Modifier keys.
const (
	ModShift    ModifierKey = 0x0001
	ModControl  ModifierKey = 0x0002
	ModAlt      ModifierKey = 0x0004
	ModSuper    ModifierKey = 0x0008
	ModCapsLock ModifierKey = 0x0010
	ModNumLock  ModifierKey = 0x0020
)

// MouseButton corresponds to a mouse button.
type MouseButton int

// Mouse buttons.
const (
	MouseButton1      MouseButton = 0
	MouseButton2      MouseButton = 1
	MouseButton3      MouseButton = 2
	MouseButton4      MouseButton = 3
	MouseButton5      MouseButton = 4
	MouseButton6      MouseButton = 5
	MouseButton7      MouseButton = 6
	MouseButton8      MouseButton = 7
	MouseButtonLast   MouseButton = MouseButton8
	MouseButtonLeft   MouseButton = MouseButton1
	MouseButtonRight  MouseButton = MouseButton2
	MouseButtonMiddle MouseButton = MouseButton3
)

// StandardCursor corresponds to a standard cursor icon.
type StandardCursor int

// Standard cursors
const (
	ArrowCursor     StandardCursor = 0x00036001
	IBeamCursor     StandardCursor = 0x00036002
	CrosshairCursor StandardCursor = 0x00036003
	HandCursor      StandardCursor = 0x00036004
	HResizeCursor   StandardCursor = 0x00036005
	VResizeCursor   StandardCursor = 0x00036006
)

// Action corresponds to a key or button action.
type Action int

// Action types.
const (
	Release Action = 0 // The key or button was released.
	Press   Action = 1 // The key or button was pressed.
	Repeat  Action = 2 // The key was held down until it repeated.
)

// InputMode corresponds to an input mode.
type InputMode int

// Input modes.
const (
	CursorMode             InputMode = 0x00033001 // See Cursor mode values
	StickyKeysMode         InputMode = 0x00033002 // Value can be either 1 or 0
	StickyMouseButtonsMode InputMode = 0x00033003 // Value can be either 1 or 0
	LockKeyMods            InputMode = 0x00033004 // Value can be either 1 or 0
	RawMouseMotion         InputMode = 0x00033005 // Value can be either 1 or 0
	Ime                    InputMode = 0x00033006 // Value can be either 1 or 0
	ImeOwnerDraw           InputMode = 0x00033011 // Value can be either 1 or 0
)

// Cursor mode values.
const (
	CursorNormal   int = 0x00034001
	CursorHidden   int = 0x00034002
	CursorDisabled int = 0x00034003
	CursorCaptured int = 0x00034004
)

// TODO: implement these input functions

// GetInputMode returns the value of an input option of the window.
func (w *Window) GetInputMode(mode InputMode) int {
	return 0
}

// SetInputMode sets an input option for the window.
func (w *Window) SetInputMode(mode InputMode, value int) {}

// RawMouseMotionSupported returns whether raw mouse motion is supported on the
// current system. This status does not change after GLFW has been initialized
// so you only need to check this once. If you attempt to enable raw motion on
// a system that does not support it, PlatformError will be emitted.
//
// Raw mouse motion is closer to the actual motion of the mouse across a
// surface. It is not affected by the scaling and acceleration applied to the
// motion of the desktop cursor. That processing is suitable for a cursor while
// raw motion is better for controlling for example a 3D camera. Because of
// this, raw mouse motion is only provided when the cursor is disabled.
//
// This function must only be called from the main thread.
func RawMouseMotionSupported() bool {
	return true
}

// Cursor represents a cursor.
type Cursor struct{}

// GetCursorPos returns the last reported position of the cursor.
//
// If the cursor is disabled (with CursorDisabled) then the cursor position is
// unbounded and limited only by the minimum and maximum values of a double.
//
// The coordinate can be converted to their integer equivalents with the floor
// function. Casting directly to an integer type works for positive coordinates,
// but fails for negative ones.
func (w *Window) GetCursorPos() (x, y float64) {
	return
}

// SetCursorPos sets the position of the cursor. The specified window must
// be focused. If the window does not have focus when this function is called,
// it fails silently.
//
// If the cursor is disabled (with CursorDisabled) then the cursor position is
// unbounded and limited only by the minimum and maximum values of a double.
func (w *Window) SetCursorPos(xpos, ypos float64) {}

// CreateCursor creates a new custom cursor image that can be set for a window with SetCursor.
// The cursor can be destroyed with Destroy. Any remaining cursors are destroyed by Terminate.
//
// The image is ideally provided in the form of *image.NRGBA.
// The pixels are 32-bit, little-endian, non-premultiplied RGBA, i.e. eight
// bits per channel with the red channel first. They are arranged canonically
// as packed sequential rows, starting from the top-left corner. If the image
// type is not *image.NRGBA, it will be converted to it.
//
// The cursor hotspot is specified in pixels, relative to the upper-left corner of the cursor image.
// Like all other coordinate systems in GLFW, the X-axis points to the right and the Y-axis points down.
func CreateCursor(img image.Image, xhot, yhot int) *Cursor {
	return &Cursor{}
}

// CreateStandardCursor returns a cursor with a standard shape,
// that can be set for a window with SetCursor.
func CreateStandardCursor(shape StandardCursor) *Cursor {
	return &Cursor{}
}

// Destroy destroys a cursor previously created with CreateCursor.
// Any remaining cursors will be destroyed by Terminate.
func (c *Cursor) Destroy() {}

// SetCursor sets the cursor image to be used when the cursor is over the client area
// of the specified window. The set cursor will only be visible when the cursor mode of the
// window is CursorNormal.
//
// On some platforms, the set cursor may not be visible unless the window also has input focus.
func (w *Window) SetCursor(c *Cursor) {}
