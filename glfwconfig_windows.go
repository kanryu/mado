//go:build windows
// +build windows

package mado

import (
	winsyscall "golang.org/x/sys/windows"
)

type PlatformContextState struct {
	Dc       winsyscall.Handle
	Handle   winsyscall.Handle
	Interval int
}

type PlatformLibraryContextState struct {
	Inited bool

	EXT_swap_control               bool
	EXT_colorspace                 bool
	ARB_multisample                bool
	ARB_framebuffer_sRGB           bool
	EXT_framebuffer_sRGB           bool
	ARB_pixel_format               bool
	ARB_create_context             bool
	ARB_create_context_profile     bool
	EXT_create_context_es2_profile bool
	ARB_create_context_robustness  bool
	ARB_create_context_no_error    bool
	ARB_context_flush_control      bool
}
