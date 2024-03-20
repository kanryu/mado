//go:build linux

package mado

type PlatformContextState struct {
}

type PlatformLibraryContextState struct {
	Inited bool
	Major  int
	Minor  int

	EGL_KHR_create_context      bool
	KHR_create_context_no_error bool
	KHR_gl_colorspace           bool
	KHR_get_all_proc_addresses  bool
	KHR_context_flush_control   bool
	EXT_present_opaque          bool
}
