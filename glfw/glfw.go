package glfw

import (
	"fmt"
)

// Version constants.
const (
	VersionMajor    = 3 // This is incremented when the API is changed in non-compatible ways.
	VersionMinor    = 3 // This is incremented when features are added to the API but it remains backward-compatible.
	VersionRevision = 9 // This is incremented when a bug fix release is made that does not contain any API changes.
)

// Init initializes the GLFW library. Before most GLFW functions can be used,
// GLFW must be initialized, and before a program terminates GLFW should be
// terminated in order to free any resources allocated during or after
// initialization.
//
// If this function fails, it calls Terminate before returning. If it succeeds,
// you should call Terminate before the program exits.
//
// Additional calls to this function after successful initialization but before
// termination will succeed but will do nothing.
//
// This function may take several seconds to complete on some systems, while on
// other systems it may take only a fraction of a second to complete.
//
// On Mac OS X, this function will change the current directory of the
// application to the Contents/Resources subdirectory of the application's
// bundle, if present.
//
// This function may only be called from the main thread.
func Init() error {
	fmt.Println("not implemented")
	return acceptError(APIUnavailable)
}

// Terminate destroys all remaining windows, frees any allocated resources and
// sets the library to an uninitialized state. Once this is called, you must
// again call Init successfully before you will be able to use most GLFW
// functions.
//
// If GLFW has been successfully initialized, this function should be called
// before the program exits. If initialization fails, there is no need to call
// this function, as it is called by Init before it returns failure.
//
// This function may only be called from the main thread.
func Terminate() {
	flushErrors()
	fmt.Println("not implemented")
}

// InitHint function sets hints for the next initialization of GLFW.
//
// The values you set hints to are never reset by GLFW, but they only take
// effect during initialization. Once GLFW has been initialized, any values you
// set will be ignored until the library is terminated and initialized again.
//
// Some hints are platform specific. These may be set on any platform but they
// will only affect their specific platform. Other platforms will ignore them.
// Setting these hints requires no platform specific headers or functions.
//
// This function must only be called from the main thread.
func InitHint(hint Hint, value int) {
	fmt.Println("not implemented")
}

// GetVersion retrieves the major, minor and revision numbers of the GLFW
// library. It is intended for when you are using GLFW as a shared library and
// want to ensure that you are using the minimum required version.
//
// This function may be called before Init.
func GetVersion() (major, minor, revision int) {
	return int(VersionMajor), int(VersionMinor), int(VersionRevision)
}

// GetVersionString returns a static string generated at compile-time according
// to which configuration macros were defined. This is intended for use when
// submitting bug reports, to allow developers to see which code paths are
// enabled in a binary.
//
// This function may be called before Init.
func GetVersionString() string {
	return fmt.Sprintf("%d.%d.%d Win32 WGL EGL OSMesa", VersionMajor, VersionMinor, VersionRevision)
}

// GetClipboardString returns the contents of the system clipboard, if it
// contains or is convertible to a UTF-8 encoded string.
//
// This function may only be called from the main thread.
func GetClipboardString() string {
	fmt.Println("not implemented")
	return ""
}

// SetClipboardString sets the system clipboard to the specified UTF-8 encoded
// string.
//
// This function may only be called from the main thread.
func SetClipboardString(str string) {
	fmt.Println("not implemented")
	panicError()
}
