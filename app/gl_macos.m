// SPDX-License-Identifier: Unlicense OR MIT

// +build darwin,!ios,nometal

#import <AppKit/AppKit.h>
#include <CoreFoundation/CoreFoundation.h>
#include <OpenGL/OpenGL.h>
#include "_cgo_export.h"

CALayer *gio_layerFactory(void) {
	@autoreleasepool {
		return [CALayer layer];
	}
}

CFTypeRef gio_createGLContext(void) {
	@autoreleasepool {
		NSOpenGLPixelFormatAttribute attr[] = {
			NSOpenGLPFAOpenGLProfile, NSOpenGLProfileVersion3_2Core,
			NSOpenGLPFAColorSize,     24,
			NSOpenGLPFAAccelerated,
			// Opt-in to automatic GPU switching. CGL-only property.
			kCGLPFASupportsAutomaticGraphicsSwitching,
			NSOpenGLPFAAllowOfflineRenderers,
			0
		};
		NSOpenGLPixelFormat *pixFormat = [[NSOpenGLPixelFormat alloc] initWithAttributes:attr];

		NSOpenGLContext *ctx = [[NSOpenGLContext alloc] initWithFormat:pixFormat shareContext: nil];
		return CFBridgingRetain(ctx);
	}
}

CFTypeRef gio_createGLContext2(NSOpenGLPixelFormatAttribute *attribs) {
	@autoreleasepool {
		NSOpenGLPixelFormat *pixFormat = [[NSOpenGLPixelFormat alloc] initWithAttributes:attribs];

		NSOpenGLContext *ctx = [[NSOpenGLContext alloc] initWithFormat:pixFormat shareContext: nil];
		return CFBridgingRetain(ctx);
	}
}

void gio_swapBuffers(CFTypeRef ctxRef)
{
    @autoreleasepool {

	NSOpenGLContext *ctx = (__bridge NSOpenGLContext *)ctxRef;
    [ctx flushBuffer];

    } // autoreleasepool
}

void gio_swapInterval(CFTypeRef ctxRef, int interval)
{
    @autoreleasepool {

	NSOpenGLContext *ctx = (__bridge NSOpenGLContext *)ctxRef;
    [ctx setValues:&interval
                              forParameter:NSOpenGLContextParameterSwapInterval];

    } // autoreleasepool
}

void gio_setContextView(CFTypeRef ctxRef, CFTypeRef viewRef) {
	NSOpenGLContext *ctx = (__bridge NSOpenGLContext *)ctxRef;
	NSView *view = (__bridge NSView *)viewRef;
	[view setWantsBestResolutionOpenGLSurface:YES];
	[ctx setView:view];
}

void gio_clearCurrentContext(void) {
	@autoreleasepool {
		[NSOpenGLContext clearCurrentContext];
	}
}

void gio_updateContext(CFTypeRef ctxRef) {
	@autoreleasepool {
		NSOpenGLContext *ctx = (__bridge NSOpenGLContext *)ctxRef;
		[ctx update];
	}
}

void gio_makeCurrentContext(CFTypeRef ctxRef) {
	@autoreleasepool {
		NSOpenGLContext *ctx = (__bridge NSOpenGLContext *)ctxRef;
		[ctx makeCurrentContext];
	}
}

void gio_lockContext(CFTypeRef ctxRef) {
	@autoreleasepool {
		NSOpenGLContext *ctx = (__bridge NSOpenGLContext *)ctxRef;
		CGLLockContext([ctx CGLContextObj]);
	}
}

void gio_unlockContext(CFTypeRef ctxRef) {
	@autoreleasepool {
		NSOpenGLContext *ctx = (__bridge NSOpenGLContext *)ctxRef;
		CGLUnlockContext([ctx CGLContextObj]);
	}
}

CFBundleRef bundleOpengl;

void gio_initNSGL(void)
{
    if (bundleOpengl)
        return;

    bundleOpengl =
        CFBundleGetBundleWithIdentifier(CFSTR("com.apple.opengl"));
}

void * gio_getProcAddress(const char* procname)
{
	if (!bundleOpengl) {
		gio_initNSGL();
	}
    CFStringRef symbolName = CFStringCreateWithCString(kCFAllocatorDefault,
                                                       procname,
                                                       kCFStringEncodingASCII);
	void *symbol;
	symbol = CFBundleGetFunctionPointerForName(bundleOpengl,
                                                          symbolName);

    CFRelease(symbolName);

    return symbol;
}