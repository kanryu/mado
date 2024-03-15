// SPDX-License-Identifier: Unlicense OR MIT

//go:build ((linux && !android) || freebsd || openbsd) && !noopengl
// +build linux,!android freebsd openbsd
// +build !noopengl

package app

type GLXint uint32
type EGLSurface uintptr
type EGLDisplay uintptr
type EGLContext uintptr
type EGLNativeDisplayType uintptr

const (
	GLX_VENDOR                                  GLXint = 1
	GLX_RGBA_BIT                                GLXint = 0x00000001
	GLX_WINDOW_BIT                              GLXint = 0x00000001
	GLX_DRAWABLE_TYPE                           GLXint = 0x8010
	GLX_RENDER_TYPE                             GLXint = 0x8011
	GLX_RGBA_TYPE                               GLXint = 0x8014
	GLX_DOUBLEBUFFER                            GLXint = 5
	GLX_STEREO                                  GLXint = 6
	GLX_AUX_BUFFERS                             GLXint = 7
	GLX_RED_SIZE                                GLXint = 8
	GLX_GREEN_SIZE                              GLXint = 9
	GLX_BLUE_SIZE                               GLXint = 10
	GLX_ALPHA_SIZE                              GLXint = 11
	GLX_DEPTH_SIZE                              GLXint = 12
	GLX_STENCIL_SIZE                            GLXint = 13
	GLX_ACCUM_RED_SIZE                          GLXint = 14
	GLX_ACCUM_GREEN_SIZE                        GLXint = 15
	GLX_ACCUM_BLUE_SIZE                         GLXint = 16
	GLX_ACCUM_ALPHA_SIZE                        GLXint = 17
	GLX_SAMPLES                                 GLXint = 0x186a1
	GLX_VISUAL_ID                               GLXint = 0x800b
	GLX_FRAMEBUFFER_SRGB_CAPABLE_ARB            GLXint = 0x20b2
	GLX_CONTEXT_DEBUG_BIT_ARB                   GLXint = 0x00000001
	GLX_CONTEXT_COMPATIBILITY_PROFILE_BIT_ARB   GLXint = 0x00000002
	GLX_CONTEXT_CORE_PROFILE_BIT_ARB            GLXint = 0x00000001
	GLX_CONTEXT_PROFILE_MASK_ARB                GLXint = 0x9126
	GLX_CONTEXT_FORWARD_COMPATIBLE_BIT_ARB      GLXint = 0x00000002
	GLX_CONTEXT_MAJOR_VERSION_ARB               GLXint = 0x2091
	GLX_CONTEXT_MINOR_VERSION_ARB               GLXint = 0x2092
	GLX_CONTEXT_FLAGS_ARB                       GLXint = 0x2094
	GLX_CONTEXT_ES2_PROFILE_BIT_EXT             GLXint = 0x00000004
	GLX_CONTEXT_ROBUST_ACCESS_BIT_ARB           GLXint = 0x00000004
	GLX_LOSE_CONTEXT_ON_RESET_ARB               GLXint = 0x8252
	GLX_CONTEXT_RESET_NOTIFICATION_STRATEGY_ARB GLXint = 0x8256
	GLX_NO_RESET_NOTIFICATION_ARB               GLXint = 0x8261
	GLX_CONTEXT_RELEASE_BEHAVIOR_ARB            GLXint = 0x2097
	GLX_CONTEXT_RELEASE_BEHAVIOR_NONE_ARB       GLXint = 0
	GLX_CONTEXT_RELEASE_BEHAVIOR_FLUSH_ARB      GLXint = 0x2098
	GLX_CONTEXT_OPENGL_NO_ERROR_ARB             GLXint = 0x31b3
)

type XID uintptr
type GLXWindow XID
type GLXDrawable XID
type GLXContext uintptr
type GLXFBConfig uintptr

// Framebuffer configuration
//
// This describes buffers and their sizes.  It also contains
// a platform-specific ID used to map back to the backend API object.
//
// It is used to pass framebuffer parameters from shared code to the platform
// API and also to enumerate and select available framebuffer configs.
type _GLFWfbconfig struct {
	redBits        int
	greenBits      int
	blueBits       int
	alphaBits      int
	depthBits      int
	stencilBits    int
	accumRedBits   int
	accumGreenBits int
	accumBlueBits  int
	accumAlphaBits int
	auxBuffers     int
	stereo         bool
	samples        int
	sRGB           bool
	doublebuffer   bool
	transparent    bool
	handle         uintptr
}

// GLX-specific per-context data
type GLFWcontextGLX struct {
	handle GLXContext
	window GLXWindow
}

// GLX-specific global data
type _GLFWlibraryGLX struct {
	major, minor int
	eventBase    int
	errorBase    int

	handle uintptr

	// // GLX 1.3 functions
	// PFNGLXGETFBCONFIGSPROC              GetFBConfigs;
	// PFNGLXGETFBCONFIGATTRIBPROC         GetFBConfigAttrib;
	// PFNGLXGETCLIENTSTRINGPROC           GetClientString;
	// PFNGLXQUERYEXTENSIONPROC            QueryExtension;
	// PFNGLXQUERYVERSIONPROC              QueryVersion;
	// PFNGLXDESTROYCONTEXTPROC            DestroyContext;
	// PFNGLXMAKECURRENTPROC               MakeCurrent;
	// PFNGLXSWAPBUFFERSPROC               SwapBuffers;
	// PFNGLXQUERYEXTENSIONSSTRINGPROC     QueryExtensionsString;
	// PFNGLXCREATENEWCONTEXTPROC          CreateNewContext;
	// PFNGLXGETVISUALFROMFBCONFIGPROC     GetVisualFromFBConfig;
	// PFNGLXCREATEWINDOWPROC              CreateWindow;
	// PFNGLXDESTROYWINDOWPROC             DestroyWindow;

	// // GLX 1.4 and extension functions
	// PFNGLXGETPROCADDRESSPROC            GetProcAddress;
	// PFNGLXGETPROCADDRESSPROC            GetProcAddressARB;
	// PFNGLXSWAPINTERVALSGIPROC           SwapIntervalSGI;
	// PFNGLXSWAPINTERVALEXTPROC           SwapIntervalEXT;
	// PFNGLXSWAPINTERVALMESAPROC          SwapIntervalMESA;
	// PFNGLXCREATECONTEXTATTRIBSARBPROC   CreateContextAttribsARB;
	SGI_swap_control               bool
	EXT_swap_control               bool
	MESA_swap_control              bool
	ARB_multisample                bool
	ARB_framebuffer_sRGB           bool
	EXT_framebuffer_sRGB           bool
	ARB_create_context             bool
	ARB_create_context_profile     bool
	ARB_create_context_robustness  bool
	EXT_create_context_es2_profile bool
	ARB_create_context_no_error    bool
	ARB_context_flush_control      bool
}

// Return the GLXFBConfig most closely matching the specified hints
func chooseGLXFBConfig(desired *_GLFWfbconfig, result *GLXFBConfig) bool {
	// GLXFBConfig* nativeConfigs;
	// _GLFWfbconfig* usableConfigs;
	// const _GLFWfbconfig* closest;
	// int i, nativeCount, usableCount;
	// const char* vendor;
	// GLFWbool trustWindowBit = GLFW_TRUE;

	// // HACK: This is a (hopefully temporary) workaround for Chromium
	// //       (VirtualBox GL) not setting the window bit on any GLXFBConfigs
	// vendor = glXGetClientString(_glfw.x11.display, GLX_VENDOR);
	// if (vendor && strcmp(vendor, "Chromium") == 0)
	// trustWindowBit = GLFW_FALSE;

	// nativeConfigs =
	// glXGetFBConfigs(_glfw.x11.display, _glfw.x11.screen, &nativeCount);
	// if (!nativeConfigs || !nativeCount)
	// {
	// _glfwInputError(GLFW_API_UNAVAILABLE, "GLX: No GLXFBConfigs returned");
	// return GLFW_FALSE;
	// }

	// usableConfigs = calloc(nativeCount, sizeof(_GLFWfbconfig));
	// usableCount = 0;

	// for (i = 0;  i < nativeCount;  i++)
	// {
	// const GLXFBConfig n = nativeConfigs[i];
	// _GLFWfbconfig* u = usableConfigs + usableCount;

	// // Only consider RGBA GLXFBConfigs
	// if (!(getGLXFBConfigAttrib(n, GLX_RENDER_TYPE) & GLX_RGBA_BIT))
	// continue;

	// // Only consider window GLXFBConfigs
	// if (!(getGLXFBConfigAttrib(n, GLX_DRAWABLE_TYPE) & GLX_WINDOW_BIT))
	// {
	// if (trustWindowBit)
	// continue;
	// }

	// if (getGLXFBConfigAttrib(n, GLX_DOUBLEBUFFER) != desired->doublebuffer)
	// continue;

	// if (desired->transparent)
	// {
	// XVisualInfo* vi = glXGetVisualFromFBConfig(_glfw.x11.display, n);
	// if (vi)
	// {
	// u->transparent = _glfwIsVisualTransparentX11(vi->visual);
	// XFree(vi);
	// }
	// }

	// u->redBits = getGLXFBConfigAttrib(n, GLX_RED_SIZE);
	// u->greenBits = getGLXFBConfigAttrib(n, GLX_GREEN_SIZE);
	// u->blueBits = getGLXFBConfigAttrib(n, GLX_BLUE_SIZE);

	// u->alphaBits = getGLXFBConfigAttrib(n, GLX_ALPHA_SIZE);
	// u->depthBits = getGLXFBConfigAttrib(n, GLX_DEPTH_SIZE);
	// u->stencilBits = getGLXFBConfigAttrib(n, GLX_STENCIL_SIZE);

	// u->accumRedBits = getGLXFBConfigAttrib(n, GLX_ACCUM_RED_SIZE);
	// u->accumGreenBits = getGLXFBConfigAttrib(n, GLX_ACCUM_GREEN_SIZE);
	// u->accumBlueBits = getGLXFBConfigAttrib(n, GLX_ACCUM_BLUE_SIZE);
	// u->accumAlphaBits = getGLXFBConfigAttrib(n, GLX_ACCUM_ALPHA_SIZE);

	// u->auxBuffers = getGLXFBConfigAttrib(n, GLX_AUX_BUFFERS);

	// if (getGLXFBConfigAttrib(n, GLX_STEREO))
	// u->stereo = GLFW_TRUE;

	// if (_glfw.glx.ARB_multisample)
	// u->samples = getGLXFBConfigAttrib(n, GLX_SAMPLES);

	// if (_glfw.glx.ARB_framebuffer_sRGB || _glfw.glx.EXT_framebuffer_sRGB)
	// u->sRGB = getGLXFBConfigAttrib(n, GLX_FRAMEBUFFER_SRGB_CAPABLE_ARB);

	// u->handle = (uintptr_t) n;
	// usableCount++;
	// }

	// closest = _glfwChooseFBConfig(desired, usableConfigs, usableCount);
	// if (closest)
	// *result = (GLXFBConfig) closest->handle;

	// XFree(nativeConfigs);
	// free(usableConfigs);

	// return closest != NULL;
	return true
}
