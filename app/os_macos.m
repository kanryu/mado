// SPDX-License-Identifier: Unlicense OR MIT

// +build darwin,!ios

#import <AppKit/AppKit.h>

#include "_cgo_export.h"

__attribute__ ((visibility ("hidden"))) CALayer *gio_layerFactory(void);

bool withPollEvents = false;


@interface GioAppDelegate : NSObject<NSApplicationDelegate>
@end

@interface GioWindowDelegate : NSObject<NSWindowDelegate> {
	NSSize windowSize;
	NSSize fbSize;
}
@end

@implementation GioWindowDelegate
- (void)windowWillMiniaturize:(NSNotification *)notification {
	NSWindow *window = (NSWindow *)[notification object];
	gio_onWindowIconify((__bridge CFTypeRef)window.contentView, true);
	gio_onHide((__bridge CFTypeRef)window.contentView);
}
- (void)windowDidDeminiaturize:(NSNotification *)notification {
	NSWindow *window = (NSWindow *)[notification object];
	gio_onWindowIconify((__bridge CFTypeRef)window.contentView, false);
	gio_onShow((__bridge CFTypeRef)window.contentView);
}
- (void)windowWillEnterFullScreen:(NSNotification *)notification {
	NSWindow *window = (NSWindow *)[notification object];
	gio_onFullscreen((__bridge CFTypeRef)window.contentView);
}
- (void)windowWillExitFullScreen:(NSNotification *)notification {
	NSWindow *window = (NSWindow *)[notification object];
	gio_onWindowed((__bridge CFTypeRef)window.contentView);
}
- (void)windowDidChangeScreen:(NSNotification *)notification {
	NSWindow *window = (NSWindow *)[notification object];
	CGDirectDisplayID dispID = [[[window screen] deviceDescription][@"NSScreenNumber"] unsignedIntValue];
	CFTypeRef view = (__bridge CFTypeRef)window.contentView;
	gio_onChangeScreen(view, dispID);
}
- (void)windowDidBecomeKey:(NSNotification *)notification {
	NSWindow *window = (NSWindow *)[notification object];
	gio_onFocus((__bridge CFTypeRef)window.contentView, 1);
}
- (void)windowDidResignKey:(NSNotification *)notification {
	NSWindow *window = (NSWindow *)[notification object];
	gio_onFocus((__bridge CFTypeRef)window.contentView, 0);
}
- (void)windowDidMove:(NSNotification *)notification
{
	NSWindow *window = (NSWindow *)[notification object];
    int x, y;
	@autoreleasepool {
		const NSRect contentRect =
			[window contentRectForFrameRect:[window frame]];

		x = contentRect.origin.x;
		y = CGDisplayBounds(CGMainDisplayID()).size.height - contentRect.origin.y - contentRect.size.height;
    } // autoreleasepool
	gio_onWindowPos((__bridge CFTypeRef)window.contentView, x, y);
}
- (BOOL)windowShouldClose:(NSWindow*)window
{
    gio_onWindowClose((__bridge CFTypeRef)window.contentView);
    return YES;
}
- (void)windowDidResize:(NSNotification *)notification
{
	NSWindow *window = (NSWindow *)[notification object];
	@autoreleasepool {
		const NSRect contentRect =
			[window contentRectForFrameRect:[window frame]];
	    const NSRect fbRect = [window convertRectToBacking:contentRect];
		if (fbRect.size.width != fbSize.width ||
			fbRect.size.height != fbSize.height)
		{
			fbSize.width  = fbRect.size.width;
			fbSize.height = fbRect.size.height;
			gio_onFramebufferSize((__bridge CFTypeRef)window.contentView, fbRect.size.width, fbRect.size.height);
		}
		if (contentRect.size.width != windowSize.width ||
			contentRect.size.height != windowSize.height)
		{
			windowSize.width  = contentRect.size.width;
			windowSize.height = contentRect.size.height;
			gio_onWindowSize((__bridge CFTypeRef)window.contentView, contentRect.size.width, contentRect.size.height);
		}
	}
}
@end

@interface GioView : NSView <CALayerDelegate,NSTextInputClient> {
	bool cursorTracked;
	NSSize windowSize;
	NSSize fbSize;
	float xscale;
	float yscale;
}
- (void)handleMouse:(NSEvent*) event
			   type:(int) typ
			     dx:(CGFloat) dx
				 dy:(CGFloat) dy;
@end

@implementation GioView
- (void)setFrameSize:(NSSize)newSize {
	[super setFrameSize:newSize];
	[self setNeedsDisplay:YES];
}
// drawRect is called when OpenGL is used, displayLayer otherwise.
// Don't know why.
- (void)drawRect:(NSRect)r {
	gio_onDraw((__bridge CFTypeRef)self);
}
- (void)displayLayer:(CALayer *)layer {
	layer.contentsScale = self.window.backingScaleFactor;
	gio_onDraw((__bridge CFTypeRef)self);
}
- (CALayer *)makeBackingLayer {
	CALayer *layer = gio_layerFactory();
	layer.delegate = self;
	return layer;
}
// - (void)viewDidChangeBackingProperties
// {
//     const NSRect contentRect = [self frame];
//     const NSRect fbRect = [self convertRectToBacking:contentRect];
//     const float xscale = fbRect.size.width / contentRect.size.width;
//     const float yscale = fbRect.size.height / contentRect.size.height;
// 	@autoreleasepool {
// 		if (xscale != self.xscale || yscale != self.yscale)
// 		{
// 			if (self.retina && self.layer)
// 				[self.layer setContentsScale:[self.object backingScaleFactor]];

// 			self.xscale = xscale;
// 			wself.yscale = yscale;
// 			gio_onContentScale((__bridge CFTypeRef)self, xscale, yscale);
// 		}

// 		if (fbRect.size.width != self.fbWidth ||
// 			fbRect.size.height != self.fbHeight)
// 		{
// 			self.fbWidth  = fbRect.size.width;
// 			self.fbHeight = fbRect.size.height;
// 			gio_onFramebufferSize((__bridge CFTypeRef)self, fbRect.size.width, fbRect.size.height);
// 		}
// 		if (contentRect.size.width != windowSize.width ||
// 			contentRect.size.height != windowSize.height)
// 		{
// 			windowSize.width  = contentRect.size.width;
// 			windowSize.height = contentRect.size.height;
// 			gio_onWindowSize((__bridge CFTypeRef)window.contentView, contentRect.size.width, contentRect.size.height);
// 		}
// 	}
// }
- (void)viewDidMoveToWindow {
	if (self.window == nil) {
		gio_onClose((__bridge CFTypeRef)self);
	}
}
- (void)handleMouse:(NSEvent*) event
			   type:(int) typ
			     dx:(CGFloat) dx
				 dy:(CGFloat) dy {
	NSPoint p = [self convertPoint:[event locationInWindow] fromView:nil];
	if (!event.hasPreciseScrollingDeltas) {
		// dx and dy are in rows and columns.
		dx *= 10;
		dy *= 10;
	}
	if (!NSMouseInRect(p, self.bounds, NO)) {
		if (cursorTracked) {
			gio_onCursorEnter((__bridge CFTypeRef)self, false);
		}
		cursorTracked = false;
		return;
	}
	if (!cursorTracked) {
		gio_onCursorEnter((__bridge CFTypeRef)self, true);
	}
	cursorTracked = true;
	// Origin is in the lower left corner. Convert to upper left.
	CGFloat height = self.bounds.size.height;
	gio_onMouse((__bridge CFTypeRef)self, (__bridge CFTypeRef)event, typ, event.buttonNumber, p.x, height - p.y, dx, dy, [event timestamp], [event modifierFlags]);
}
- (void)mouseDown:(NSEvent *)event {
	[self handleMouse:event type:MOUSE_DOWN dx:0 dy:0];
}
- (void)mouseUp:(NSEvent *)event {
	[self handleMouse:event type:MOUSE_UP dx:0 dy:0];
}
- (void)rightMouseDown:(NSEvent *)event {
	[self handleMouse:event type:MOUSE_DOWN dx:0 dy:0];
}
- (void)rightMouseUp:(NSEvent *)event {
	[self handleMouse:event type:MOUSE_UP dx:0 dy:0];
}
- (void)otherMouseDown:(NSEvent *)event {
	[self handleMouse:event type:MOUSE_DOWN dx:0 dy:0];
}
- (void)otherMouseUp:(NSEvent *)event {
	[self handleMouse:event type:MOUSE_UP dx:0 dy:0];
}
- (void)mouseMoved:(NSEvent *)event {
	[self handleMouse:event type:MOUSE_MOVE dx:0 dy:0];
}
- (void)mouseDragged:(NSEvent *)event {
	[self handleMouse:event type:MOUSE_MOVE dx:0 dy:0];
}
- (void)rightMouseDragged:(NSEvent *)event {
	[self handleMouse:event type:MOUSE_MOVE dx:0 dy:0];
}
- (void)otherMouseDragged:(NSEvent *)event {
	[self handleMouse:event type:MOUSE_MOVE dx:0 dy:0];
}
- (void)scrollWheel:(NSEvent *)event {
	CGFloat dx = -event.scrollingDeltaX;
	CGFloat dy = -event.scrollingDeltaY;
	[self handleMouse:event type:MOUSE_SCROLL dx:dx dy:dy];
}
- (void)keyDown:(NSEvent *)event {
	[self interpretKeyEvents:[NSArray arrayWithObject:event]];
	NSString *keys = [event charactersIgnoringModifiers];
	gio_onKeys((__bridge CFTypeRef)self, (__bridge CFTypeRef)keys, [event keyCode], [event timestamp], [event modifierFlags], true);
}
- (void)keyUp:(NSEvent *)event {
	NSString *keys = [event charactersIgnoringModifiers];
	gio_onKeys((__bridge CFTypeRef)self, (__bridge CFTypeRef)keys, [event keyCode], [event timestamp], [event modifierFlags], false);
}
- (void)insertText:(id)string {
	gio_onText((__bridge CFTypeRef)self, (__bridge CFTypeRef)string);
}
- (void)doCommandBySelector:(SEL)sel {
	// Don't pass commands up the responder chain.
	// They will end up in a beep.
}

- (BOOL)hasMarkedText {
	int res = gio_hasMarkedText((__bridge CFTypeRef)self);
	return res ? YES : NO;
}
- (NSRange)markedRange {
	return gio_markedRange((__bridge CFTypeRef)self);
}
- (NSRange)selectedRange {
	return gio_selectedRange((__bridge CFTypeRef)self);
}
- (void)unmarkText {
	gio_unmarkText((__bridge CFTypeRef)self);
}
- (void)setMarkedText:(id)string
        selectedRange:(NSRange)selRange
     replacementRange:(NSRange)replaceRange {
	NSString *str;
	// string is either an NSAttributedString or an NSString.
	if ([string isKindOfClass:[NSAttributedString class]]) {
		str = [string string];
	} else {
		str = string;
	}
	gio_setMarkedText((__bridge CFTypeRef)self, (__bridge CFTypeRef)str, selRange, replaceRange);
}
- (NSArray<NSAttributedStringKey> *)validAttributesForMarkedText {
	return nil;
}
- (NSAttributedString *)attributedSubstringForProposedRange:(NSRange)range
                                                actualRange:(NSRangePointer)actualRange {
	NSString *str = CFBridgingRelease(gio_substringForProposedRange((__bridge CFTypeRef)self, range, actualRange));
	return [[NSAttributedString alloc] initWithString:str attributes:nil];
}
- (void)insertText:(id)string
  replacementRange:(NSRange)replaceRange {
	NSString *str;
	// string is either an NSAttributedString or an NSString.
	if ([string isKindOfClass:[NSAttributedString class]]) {
		str = [string string];
	} else {
		str = string;
	}
	gio_insertText((__bridge CFTypeRef)self, (__bridge CFTypeRef)str, replaceRange);
}
- (NSUInteger)characterIndexForPoint:(NSPoint)p {
	return gio_characterIndexForPoint((__bridge CFTypeRef)self, p);
}
- (NSRect)firstRectForCharacterRange:(NSRange)rng
                         actualRange:(NSRangePointer)actual {
    NSRect r = gio_firstRectForCharacterRange((__bridge CFTypeRef)self, rng, actual);
    r = [self convertRect:r toView:nil];
    return [[self window] convertRectToScreen:r];
}
@end

// Delegates are weakly referenced from their peers. Nothing
// else holds a strong reference to our window delegate, so
// keep a single global reference instead.
static GioWindowDelegate *globalWindowDel;

static CVReturn displayLinkCallback(CVDisplayLinkRef dl, const CVTimeStamp *inNow, const CVTimeStamp *inOutputTime, CVOptionFlags flagsIn, CVOptionFlags *flagsOut, void *handle) {
	gio_onFrameCallback(dl);
	return kCVReturnSuccess;
}

CFTypeRef gio_createDisplayLink(void) {
	CVDisplayLinkRef dl;
	CVDisplayLinkCreateWithActiveCGDisplays(&dl);
	CVDisplayLinkSetOutputCallback(dl, displayLinkCallback, nil);
	return dl;
}

int gio_startDisplayLink(CFTypeRef dl) {
	return CVDisplayLinkStart((CVDisplayLinkRef)dl);
}

int gio_stopDisplayLink(CFTypeRef dl) {
	return CVDisplayLinkStop((CVDisplayLinkRef)dl);
}

void gio_releaseDisplayLink(CFTypeRef dl) {
	CVDisplayLinkRelease((CVDisplayLinkRef)dl);
}

void gio_setDisplayLinkDisplay(CFTypeRef dl, uint64_t did) {
	CVDisplayLinkSetCurrentCGDisplay((CVDisplayLinkRef)dl, (CGDirectDisplayID)did);
}

void gio_hideCursor() {
	@autoreleasepool {
		[NSCursor hide];
	}
}

void gio_showCursor() {
	@autoreleasepool {
		[NSCursor unhide];
	}
}

// some cursors are not public, this tries to use a private cursor
// and uses fallback when the use of private cursor fails.
void gio_trySetPrivateCursor(SEL cursorName, NSCursor* fallback) {
	if ([NSCursor respondsToSelector:cursorName]) {
		id object = [NSCursor performSelector:cursorName];
		if ([object isKindOfClass:[NSCursor class]]) {
			[(NSCursor*)object set];
			return;
		}
	}
	[fallback set];
}

void gio_setCursor(NSUInteger curID) {
	@autoreleasepool {
		switch (curID) {
			case 0: // pointer.CursorDefault
				[NSCursor.arrowCursor set];
				break;
			// case 1: // pointer.CursorNone
			case 2: // pointer.CursorText
				[NSCursor.IBeamCursor set];
				break;
			case 3: // pointer.CursorVerticalText
				[NSCursor.IBeamCursorForVerticalLayout set];
				break;
			case 4: // pointer.CursorPointer
				[NSCursor.pointingHandCursor set];
				break;
			case 5: // pointer.CursorCrosshair
				[NSCursor.crosshairCursor set];
				break;
			case 6: // pointer.CursorAllScroll
				// For some reason, using _moveCursor fails on Monterey.
				// gio_trySetPrivateCursor(@selector(_moveCursor), NSCursor.arrowCursor);
				[NSCursor.arrowCursor set];
				break;
			case 7: // pointer.CursorColResize
				[NSCursor.resizeLeftRightCursor set];
				break;
			case 8: // pointer.CursorRowResize
				[NSCursor.resizeUpDownCursor set];
				break;
			case 9: // pointer.CursorGrab
				// [NSCursor.openHandCursor set];
				gio_trySetPrivateCursor(@selector(openHandCursor), NSCursor.arrowCursor);
				break;
			case 10: // pointer.CursorGrabbing
				// [NSCursor.closedHandCursor set];
				gio_trySetPrivateCursor(@selector(closedHandCursor), NSCursor.arrowCursor);
				break;
			case 11: // pointer.CursorNotAllowed
				[NSCursor.operationNotAllowedCursor set];
				break;
			case 12: // pointer.CursorWait
				gio_trySetPrivateCursor(@selector(busyButClickableCursor), NSCursor.arrowCursor);
				break;
			case 13: // pointer.CursorProgress
				gio_trySetPrivateCursor(@selector(busyButClickableCursor), NSCursor.arrowCursor);
				break;
			case 14: // pointer.CursorNorthWestResize
				gio_trySetPrivateCursor(@selector(_windowResizeNorthWestCursor), NSCursor.resizeUpDownCursor);
				break;
			case 15: // pointer.CursorNorthEastResize
				gio_trySetPrivateCursor(@selector(_windowResizeNorthEastCursor), NSCursor.resizeUpDownCursor);
				break;
			case 16: // pointer.CursorSouthWestResize
				gio_trySetPrivateCursor(@selector(_windowResizeSouthWestCursor), NSCursor.resizeUpDownCursor);
				break;
			case 17: // pointer.CursorSouthEastResize
				gio_trySetPrivateCursor(@selector(_windowResizeSouthEastCursor), NSCursor.resizeUpDownCursor);
				break;
			case 18: // pointer.CursorNorthSouthResize
				[NSCursor.resizeUpDownCursor set];
				break;
			case 19: // pointer.CursorEastWestResize
				[NSCursor.resizeLeftRightCursor set];
				break;
			case 20: // pointer.CursorWestResize
				[NSCursor.resizeLeftCursor set];
				break;
			case 21: // pointer.CursorEastResize
				[NSCursor.resizeRightCursor set];
				break;
			case 22: // pointer.CursorNorthResize
				[NSCursor.resizeUpCursor set];
				break;
			case 23: // pointer.CursorSouthResize
				[NSCursor.resizeDownCursor set];
				break;
			case 24: // pointer.CursorNorthEastSouthWestResize
				gio_trySetPrivateCursor(@selector(_windowResizeNorthEastSouthWestCursor), NSCursor.resizeUpDownCursor);
				break;
			case 25: // pointer.CursorNorthWestSouthEastResize
				gio_trySetPrivateCursor(@selector(_windowResizeNorthWestSouthEastCursor), NSCursor.resizeUpDownCursor);
				break;
			default:
				[NSCursor.arrowCursor set];
				break;
		}
	}
}

CFTypeRef gio_createWindow(CFTypeRef viewRef, CGFloat width, CGFloat height, CGFloat minWidth, CGFloat minHeight, CGFloat maxWidth, CGFloat maxHeight) {
	@autoreleasepool {
		NSRect rect = NSMakeRect(0, 0, width, height);
		NSUInteger styleMask = NSTitledWindowMask |
			NSResizableWindowMask |
			NSMiniaturizableWindowMask |
			NSClosableWindowMask;

		NSWindow* window = [[NSWindow alloc] initWithContentRect:rect
													   styleMask:styleMask
														 backing:NSBackingStoreBuffered
														   defer:NO];
		if (minWidth > 0 || minHeight > 0) {
			window.contentMinSize = NSMakeSize(minWidth, minHeight);
		}
		if (maxWidth > 0 || maxHeight > 0) {
			window.contentMaxSize = NSMakeSize(maxWidth, maxHeight);
		}
		[window setAcceptsMouseMovedEvents:YES];
		NSView *view = (__bridge NSView *)viewRef;
		[window setContentView:view];
		[window makeFirstResponder:view];
		window.delegate = globalWindowDel;
		return (__bridge_retained CFTypeRef)window;
	}
}

CFTypeRef gio_createView(void) {
	@autoreleasepool {
		NSRect frame = NSMakeRect(0, 0, 0, 0);
		GioView* view = [[GioView alloc] initWithFrame:frame];
		view.wantsLayer = YES;
		view.layerContentsRedrawPolicy = NSViewLayerContentsRedrawDuringViewResize;
		return CFBridgingRetain(view);
	}
}

@implementation GioAppDelegate
- (void)applicationDidFinishLaunching:(NSNotification *)aNotification {
	[NSApp setActivationPolicy:NSApplicationActivationPolicyRegular];
	[NSApp activateIgnoringOtherApps:YES];
	if (withPollEvents) {
		[NSApp stop:nil];
	}
	gio_onFinishLaunching();
}
- (void)applicationDidHide:(NSNotification *)aNotification {
	gio_onAppHide();
}
- (void)applicationWillUnhide:(NSNotification *)notification {
	gio_onAppShow();
}
@end

void gio_main() {
	@autoreleasepool {
		[NSApplication sharedApplication];
		GioAppDelegate *del = [[GioAppDelegate alloc] init];
		[NSApp setDelegate:del];

		NSMenuItem *mainMenu = [NSMenuItem new];

		NSMenu *menu = [NSMenu new];
		NSMenuItem *hideMenuItem = [[NSMenuItem alloc] initWithTitle:@"Hide"
															  action:@selector(hide:)
													   keyEquivalent:@"h"];
		[menu addItem:hideMenuItem];
		NSMenuItem *quitMenuItem = [[NSMenuItem alloc] initWithTitle:@"Quit"
															  action:@selector(terminate:)
													   keyEquivalent:@"q"];
		[menu addItem:quitMenuItem];
		[mainMenu setSubmenu:menu];
		NSMenu *menuBar = [NSMenu new];
		[menuBar addItem:mainMenu];
		[NSApp setMainMenu:menuBar];

		globalWindowDel = [[GioWindowDelegate alloc] init];

		[NSApp run];
	}
}

void gio_enablePollEvents(void)
{
	withPollEvents = true;
}

void gio_PollEvents(void)
{
    @autoreleasepool {

    for (;;)
    {
        NSEvent* event = [NSApp nextEventMatchingMask:NSEventMaskAny
                                            untilDate:[NSDate distantPast]
                                               inMode:NSDefaultRunLoopMode
                                              dequeue:YES];
        if (event == nil)
            break;

        [NSApp sendEvent:event];
    }

    } // autoreleasepool
}