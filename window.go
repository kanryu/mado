package mado

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"runtime"
	"time"

	"github.com/kanryu/mado/font/gofont"
	"github.com/kanryu/mado/gpu"
	"github.com/kanryu/mado/internal/debug"
	"github.com/kanryu/mado/internal/ops"
	"github.com/kanryu/mado/io/event"
	"github.com/kanryu/mado/io/input"
	"github.com/kanryu/mado/io/key"
	"github.com/kanryu/mado/io/pointer"
	"github.com/kanryu/mado/io/system"
	"github.com/kanryu/mado/layout"
	"github.com/kanryu/mado/op"
	"github.com/kanryu/mado/text"
	"github.com/kanryu/mado/unit"
	"github.com/kanryu/mado/widget"
	"github.com/kanryu/mado/widget/material"
)

// Window represents a window.
type Window struct {
	ctx Context
	gpu gpu.GPU

	// driverFuncs is a channel of functions to run when
	// the Window has a valid driver.
	driverFuncs chan func(d Driver)
	// wakeups wakes up the native event loop to send a
	// WakeupEvent that flushes driverFuncs.
	wakeups chan struct{}
	// wakeupFuncs is sent wakeup functions when the driver changes.
	WakeupFuncs chan func()
	// redraws is notified when a redraw is requested by the client.
	redraws chan struct{}
	// immediateRedraws is like redraw but doesn't need a wakeup.
	immediateRedraws chan struct{}
	// scheduledRedraws is sent the most recent delayed redraw time.
	scheduledRedraws chan time.Time
	// options are the options waiting to be applied.
	Options chan []Option
	// actions are the actions waiting to be performed.
	Actions chan system.Action

	// out is where the platform backend delivers events bound for the
	// user program.
	out      chan event.Event
	frames   chan *op.Ops
	frameAck chan struct{}
	Destroy  chan struct{}

	stage        Stage
	animating    bool
	hasNextFrame bool
	nextFrame    time.Time
	// viewport is the latest frame size with insets applied.
	viewport image.Rectangle
	// metric is the metric from the most recent frame.
	Metric unit.Metric

	Queue       input.Router
	cursor      pointer.Cursor
	Decorations struct {
		op.Ops
		// enabled tracks the Decorated option as
		// given to the Option method. It may differ
		// from Config.Decorated depending on platform
		// capability.
		Enabled bool
		Config
		Height        unit.Dp
		currentHeight int
		*material.Theme
		*widget.Decorations
	}

	Callbacks Callbacks

	nocontext bool

	// semantic data, lazily evaluated if requested by a backend to speed up
	// the cases where semantic data is not needed.
	Semantic struct {
		// uptodate tracks whether the fields below are up to date.
		uptodate bool
		Root     input.SemanticID
		PrevTree []input.SemanticNode
		tree     []input.SemanticNode
		Ids      map[input.SemanticID]input.SemanticNode
	}

	ImeState EditorState

	// event stores the state required for processing and delivering events
	// from NextEvent. If we had support for range over func, this would
	// be the iterator state.
	EventState struct {
		created     bool
		InitialOpts []Option
		wakeup      func()
		timer       *time.Timer
	}
	// }
}

// NewWindow creates a new window for a set of window
// options. The options are hints; the platform is free to
// ignore or adjust them.
//
// If the current program is running on iOS or Android,
// NewWindow returns the window previously created by the
// platform.
//
// Calling NewWindow more than once is not supported on
// iOS, Android, WebAssembly.
func NewWindow(options ...Option) *Window {
	debug.Parse()
	// Measure decoration height.
	deco := new(widget.Decorations)
	theme := material.NewTheme()
	theme.Shaper = text.NewShaper(text.NoSystemFonts(), text.WithCollection(gofont.Regular()))
	decoStyle := material.Decorations(theme, deco, 0, "")
	gtx := layout.Context{
		Ops: new(op.Ops),
		// Measure in Dp.
		Metric: unit.Metric{},
	}
	// Allow plenty of space.
	gtx.Constraints.Max.Y = 200
	dims := decoStyle.Layout(gtx)
	decoHeight := unit.Dp(dims.Size.Y)
	defaultOptions := []Option{
		Size(800, 600),
		Title("Gio"),
		Decorated(true),
		DecoHeightOpt(decoHeight),
	}
	options = append(defaultOptions, options...)
	var cnf Config
	cnf.Apply(unit.Metric{}, options)

	w := &Window{
		out:              make(chan event.Event),
		immediateRedraws: make(chan struct{}),
		redraws:          make(chan struct{}, 1),
		scheduledRedraws: make(chan time.Time, 1),
		frames:           make(chan *op.Ops),
		frameAck:         make(chan struct{}),
		driverFuncs:      make(chan func(d Driver), 1),
		wakeups:          make(chan struct{}, 1),
		WakeupFuncs:      make(chan func()),
		Destroy:          make(chan struct{}),
		Options:          make(chan []Option, 1),
		Actions:          make(chan system.Action, 1),
		nocontext:        cnf.CustomRenderer,
	}
	w.Decorations.Theme = theme
	w.Decorations.Decorations = deco
	w.Decorations.Enabled = cnf.Decorated
	w.Decorations.Height = decoHeight
	w.ImeState.Compose = key.Range{Start: -1, End: -1}
	w.Semantic.Ids = make(map[input.SemanticID]input.SemanticNode)
	w.Callbacks.SetWindow(w)
	w.EventState.InitialOpts = options
	return w
}

// update the window contents, input operations declare input handlers,
// and so on. The supplied operations list completely replaces the window state
// from previous calls.
func (w *Window) update(frame *op.Ops) {
	w.frames <- frame
	<-w.frameAck
}

func (w *Window) validateAndProcess(d Driver, size image.Point, sync bool, frame *op.Ops, sigChan chan<- struct{}) error {
	signal := func() {
		if sigChan != nil {
			// We're done with frame, let the client continue.
			sigChan <- struct{}{}
			// Signal at most once.
			sigChan = nil
		}
	}
	defer signal()
	for {
		if w.gpu == nil && !w.nocontext {
			var err error
			if w.ctx == nil {
				w.ctx, err = d.NewContext()
				if err != nil {
					return err
				}
				sync = true
			}
		}
		if sync && w.ctx != nil {
			if err := w.ctx.Refresh(); err != nil {
				if errors.Is(err, ErrOutOfDate) {
					// Surface couldn't be created for transient reasons. Skip
					// this frame and wait for the next.
					return nil
				}
				w.destroyGPU()
				if errors.Is(err, gpu.ErrDeviceLost) {
					continue
				}
				return err
			}
		}
		if w.ctx != nil {
			if err := w.ctx.Lock(); err != nil {
				w.destroyGPU()
				return err
			}
		}
		if w.gpu == nil && !w.nocontext {
			gpu, err := gpu.New(w.ctx.API())
			if err != nil {
				w.ctx.Unlock()
				w.destroyGPU()
				return err
			}
			w.gpu = gpu
		}
		if w.gpu != nil {
			if err := w.frame(frame, size); err != nil {
				w.ctx.Unlock()
				if errors.Is(err, ErrOutOfDate) {
					// GPU surface needs refreshing.
					sync = true
					continue
				}
				w.destroyGPU()
				if errors.Is(err, gpu.ErrDeviceLost) {
					continue
				}
				return err
			}
		}
		w.Queue.Frame(frame)
		// Let the client continue as soon as possible, in particular before
		// a potentially blocking Present.
		signal()
		var err error
		if w.gpu != nil {
			err = w.ctx.Present()
			w.ctx.Unlock()
		}
		return err
	}
}

func (w *Window) frame(frame *op.Ops, viewport image.Point) error {
	if runtime.GOOS == "js" {
		// Use transparent black when Gio is embedded, to allow mixing of Gio and
		// foreign content below.
		w.gpu.Clear(color.NRGBA{A: 0x00, R: 0x00, G: 0x00, B: 0x00})
	} else {
		w.gpu.Clear(color.NRGBA{A: 0xff, R: 0xff, G: 0xff, B: 0xff})
	}
	target, err := w.ctx.RenderTarget()
	if err != nil {
		return err
	}
	return w.gpu.Frame(frame, target, viewport)
}

func (w *Window) processFrame(d Driver) {
	for k := range w.Semantic.Ids {
		delete(w.Semantic.Ids, k)
	}
	w.Semantic.uptodate = false
	q := &w.Queue
	switch q.TextInputState() {
	case input.TextInputOpen:
		d.ShowTextInput(true)
	case input.TextInputClose:
		d.ShowTextInput(false)
	}
	if hint, ok := q.TextInputHint(); ok {
		d.SetInputHint(hint)
	}
	if mime, txt, ok := q.WriteClipboard(); ok {
		d.WriteClipboard(mime, txt)
	}
	if q.ClipboardRequested() {
		d.ReadClipboard()
	}
	oldState := w.ImeState
	newState := oldState
	newState.EditorState = q.EditorState()
	if newState != oldState {
		w.ImeState = newState
		d.EditorStateChanged(oldState, newState)
	}
	if t, ok := q.WakeupTime(); ok {
		w.SetNextFrame(t)
	}
	w.UpdateAnimation(d)
}

// Invalidate the window such that a [FrameEvent] will be generated immediately.
// If the window is inactive, the event is sent when the window becomes active.
//
// Note that Invalidate is intended for externally triggered updates, such as a
// response from a network request. The [op.InvalidateCmd] command is more efficient
// for animation.
//
// Invalidate is safe for concurrent use.
func (w *Window) Invalidate() {
	select {
	case w.immediateRedraws <- struct{}{}:
		return
	default:
	}
	select {
	case w.redraws <- struct{}{}:
		w.Wakeup()
	default:
	}
}

// Option applies the options to the window.
func (w *Window) Option(opts ...Option) {
	if len(opts) == 0 {
		return
	}
	for {
		select {
		case old := <-w.Options:
			opts = append(old, opts...)
		case w.Options <- opts:
			w.Wakeup()
			return
		}
	}
}

// Run f in the same thread as the native window event loop, and wait for f to
// return or the window to close. Run is guaranteed not to deadlock if it is
// invoked during the handling of a [ViewEvent], [FrameEvent],
// [StageEvent]; call Run in a separate goroutine to avoid deadlock in all
// other cases.
//
// Note that most programs should not call Run; configuring a Window with
// [CustomRenderer] is a notable exception.
func (w *Window) Run(f func()) {
	done := make(chan struct{})
	w.driverDefer(func(d Driver) {
		defer close(done)
		f()
	})
	select {
	case <-done:
	case <-w.Destroy:
	}
}

// driverDefer is like Run but can be run from any context. It doesn't wait
// for f to return.
func (w *Window) driverDefer(f func(d Driver)) {
	select {
	case w.driverFuncs <- f:
		w.Wakeup()
	case <-w.Destroy:
	}
}

func (w *Window) UpdateAnimation(d Driver) {
	animate := false
	if w.stage >= StageInactive && w.hasNextFrame {
		if dt := time.Until(w.nextFrame); dt <= 0 {
			animate = true
		} else {
			// Schedule redraw.
			select {
			case <-w.scheduledRedraws:
			default:
			}
			w.scheduledRedraws <- w.nextFrame
		}
	}
	if animate != w.animating {
		w.animating = animate
		d.SetAnimating(animate)
	}
}

func (w *Window) Wakeup() {
	select {
	case w.wakeups <- struct{}{}:
	default:
	}
}

func (w *Window) SetNextFrame(at time.Time) {
	if !w.hasNextFrame || at.Before(w.nextFrame) {
		w.hasNextFrame = true
		w.nextFrame = at
	}
}

func (w *Window) moveFocus(dir key.FocusDirection) {
	w.Queue.MoveFocus(dir)
	if _, handled := w.Queue.WakeupTime(); handled {
		w.Queue.RevealFocus(w.viewport)
	} else {
		var v image.Point
		switch dir {
		case key.FocusRight:
			v = image.Pt(+1, 0)
		case key.FocusLeft:
			v = image.Pt(-1, 0)
		case key.FocusDown:
			v = image.Pt(0, +1)
		case key.FocusUp:
			v = image.Pt(0, -1)
		default:
			return
		}
		const scrollABit = unit.Dp(50)
		dist := v.Mul(int(w.Metric.Dp(scrollABit)))
		w.Queue.ScrollFocus(dist)
	}
}

func (w *Window) waitAck(d Driver) {
	for {
		select {
		case f := <-w.driverFuncs:
			f(d)
		case w.out <- theFlushEvent:
			// A dummy event went through, so we know the application has processed the previous event.
			return
		case <-w.immediateRedraws:
			// Invalidate was called during frame processing.
			w.SetNextFrame(time.Time{})
			w.UpdateAnimation(d)
		}
	}
}

func (w *Window) destroyGPU() {
	if w.gpu != nil {
		w.ctx.Lock()
		w.gpu.Release()
		w.ctx.Unlock()
		w.gpu = nil
	}
	if w.ctx != nil {
		w.ctx.Release()
		w.ctx = nil
	}
}

// waitFrame waits for the client to either call [FrameEvent.Frame]
// or to continue event handling.
func (w *Window) waitFrame(d Driver) *op.Ops {
	for {
		select {
		case f := <-w.driverFuncs:
			f(d)
		case frame := <-w.frames:
			// The client called FrameEvent.Frame.
			return frame
		case w.out <- theFlushEvent:
			// The client ignored FrameEvent and continued processing
			// events.
			return nil
		case <-w.immediateRedraws:
			// Invalidate was called during frame processing.
			w.SetNextFrame(time.Time{})
		}
	}
}

// UpdateSemantics refreshes the semantics tree, the id to node map and the ids of
// updated nodes.
func (w *Window) UpdateSemantics() {
	if w.Semantic.uptodate {
		return
	}
	w.Semantic.uptodate = true
	w.Semantic.PrevTree, w.Semantic.tree = w.Semantic.tree, w.Semantic.PrevTree
	w.Semantic.tree = w.Queue.AppendSemantics(w.Semantic.tree[:0])
	w.Semantic.Root = w.Semantic.tree[0].ID
	for _, n := range w.Semantic.tree {
		w.Semantic.Ids[n.ID] = n
	}
}

// CollectSemanticDiffs traverses the previous semantic tree, noting changed nodes.
func (w *Window) CollectSemanticDiffs(diffs *[]input.SemanticID, n input.SemanticNode) {
	newNode, exists := w.Semantic.Ids[n.ID]
	// Ignore deleted nodes, as their disappearance will be reported through an
	// ancestor node.
	if !exists {
		return
	}
	diff := newNode.Desc != n.Desc || len(n.Children) != len(newNode.Children)
	for i, ch := range n.Children {
		if !diff {
			newCh := newNode.Children[i]
			diff = ch.ID != newCh.ID
		}
		w.CollectSemanticDiffs(diffs, ch)
	}
	if diff {
		*diffs = append(*diffs, n.ID)
	}
}

func (w *Window) UpdateState(d Driver) {
	for {
		select {
		case f := <-w.driverFuncs:
			f(d)
		case <-w.redraws:
			w.SetNextFrame(time.Time{})
			w.UpdateAnimation(d)
		default:
			return
		}
	}
}

func (w *Window) ProcessEvent(d Driver, e event.Event) bool {
	select {
	case <-w.Destroy:
		return false
	default:
	}
	switch e2 := e.(type) {
	case StageEvent:
		if e2.Stage < StageInactive {
			if w.gpu != nil {
				w.ctx.Lock()
				w.gpu.Release()
				w.gpu = nil
				w.ctx.Unlock()
			}
		}
		w.stage = e2.Stage
		w.UpdateAnimation(d)
		w.out <- e
		w.waitAck(d)
	case FrameEvent:
		if e2.Size == (image.Point{}) {
			panic(errors.New("internal error: zero-sized Draw"))
		}
		if w.stage < StageInactive {
			// No drawing if not visible.
			break
		}
		w.Metric = e2.Metric
		w.hasNextFrame = false
		e2.Frame = w.update
		e2.Source = w.Queue.Source()

		// Prepare the decorations and update the frame insets.
		wrapper := &w.Decorations.Ops
		wrapper.Reset()
		viewport := image.Rectangle{
			Min: image.Point{
				X: e2.Metric.Dp(e2.Insets.Left),
				Y: e2.Metric.Dp(e2.Insets.Top),
			},
			Max: image.Point{
				X: e2.Size.X - e2.Metric.Dp(e2.Insets.Right),
				Y: e2.Size.Y - e2.Metric.Dp(e2.Insets.Bottom),
			},
		}
		// Scroll to focus if viewport is shrinking in any dimension.
		if old, new := w.viewport.Size(), viewport.Size(); new.X < old.X || new.Y < old.Y {
			w.Queue.RevealFocus(viewport)
		}
		w.viewport = viewport
		viewSize := e2.Size
		m := op.Record(wrapper)
		size, offset := w.decorate(d, e2, wrapper)
		e2.Size = size
		deco := m.Stop()
		w.out <- e2
		frame := w.waitFrame(d)
		var signal chan<- struct{}
		if frame != nil {
			signal = w.frameAck
			off := op.Offset(offset).Push(wrapper)
			ops.AddCall(&wrapper.Internal, &frame.Internal, ops.PC{}, ops.PCFor(&frame.Internal))
			off.Pop()
		}
		deco.Add(wrapper)
		if err := w.validateAndProcess(d, viewSize, e2.Sync, wrapper, signal); err != nil {
			w.destroyGPU()
			w.out <- DestroyEvent{Err: err}
			close(w.Destroy)
			break
		}
		w.processFrame(d)
		w.updateCursor(d)
	case DestroyEvent:
		w.destroyGPU()
		w.out <- e2
		close(w.Destroy)
	case ViewEvent:
		w.out <- e
		w.waitAck(d)
	case ConfigEvent:
		w.Decorations.Config = e2.Config
		e2.Config = w.effectiveConfig()
		w.out <- e2
	case WakeupEvent:
	case event.Event:
		focusDir := key.FocusDirection(-1)
		if e, ok := e2.(key.Event); ok && e.State == key.Press {
			isMobile := runtime.GOOS == "ios" || runtime.GOOS == "android"
			switch {
			case e.Name == key.NameTab && e.Modifiers == 0:
				focusDir = key.FocusForward
			case e.Name == key.NameTab && e.Modifiers == key.ModShift:
				focusDir = key.FocusBackward
			case e.Name == key.NameUpArrow && e.Modifiers == 0 && isMobile:
				focusDir = key.FocusUp
			case e.Name == key.NameDownArrow && e.Modifiers == 0 && isMobile:
				focusDir = key.FocusDown
			case e.Name == key.NameLeftArrow && e.Modifiers == 0 && isMobile:
				focusDir = key.FocusLeft
			case e.Name == key.NameRightArrow && e.Modifiers == 0 && isMobile:
				focusDir = key.FocusRight
			}
		}
		e := e2
		if focusDir != -1 {
			e = input.SystemEvent{Event: e}
		}
		w.Queue.Queue(e)
		t, handled := w.Queue.WakeupTime()
		if focusDir != -1 && !handled {
			w.moveFocus(focusDir)
			t, handled = w.Queue.WakeupTime()
		}
		w.updateCursor(d)
		if handled {
			w.SetNextFrame(t)
			w.UpdateAnimation(d)
		}
		return handled
	}
	return true
}

// NextEvent blocks until an event is received from the window, such as
// [FrameEvent]. It blocks forever if called after [DestroyEvent]
// has been returned.
func (w *Window) NextEvent() event.Event {
	state := &w.EventState
	// if !state.created {
	// 	state.created = true
	// 	if err := newWindow(&w.callbacks, state.initialOpts); err != nil {
	// 		close(w.destroy)
	// 		return DestroyEvent{Err: err}
	// 	}
	// }
	for {
		var (
			wakeups <-chan struct{}
			timeC   <-chan time.Time
		)
		if state.wakeup != nil {
			wakeups = w.wakeups
			if state.timer != nil {
				timeC = state.timer.C
			}
		}
		select {
		case t := <-w.scheduledRedraws:
			if state.timer != nil {
				state.timer.Stop()
			}
			state.timer = time.NewTimer(time.Until(t))
		case e := <-w.out:
			// Receiving a flushEvent indicates to the platform backend that
			// all previous events have been processed by the user program.
			if _, ok := e.(flushEvent); ok {
				break
			}
			return e
		case <-timeC:
			select {
			case w.redraws <- struct{}{}:
				state.wakeup()
			default:
			}
		case <-wakeups:
			state.wakeup()
		case state.wakeup = <-w.WakeupFuncs:
		}
	}
}

func (w *Window) updateCursor(d Driver) {
	if c := w.Queue.Cursor(); c != w.cursor {
		w.cursor = c
		d.SetCursor(c)
	}
}

func (w *Window) fallbackDecorate() bool {
	cnf := w.Decorations.Config
	return w.Decorations.Enabled && !cnf.Decorated && cnf.Mode != Fullscreen && !w.nocontext
}

// decorate the window if enabled and returns the corresponding Insets.
func (w *Window) decorate(d Driver, e FrameEvent, o *op.Ops) (size, offset image.Point) {
	if !w.fallbackDecorate() {
		return e.Size, image.Pt(0, 0)
	}
	deco := w.Decorations.Decorations
	allActions := system.ActionMinimize | system.ActionMaximize | system.ActionUnmaximize |
		system.ActionClose | system.ActionMove
	style := material.Decorations(w.Decorations.Theme, deco, allActions, w.Decorations.Config.Title)
	// Update the decorations based on the current window mode.
	var actions system.Action
	switch m := w.Decorations.Config.Mode; m {
	case Windowed:
		actions |= system.ActionUnmaximize
	case Minimized:
		actions |= system.ActionMinimize
	case Maximized:
		actions |= system.ActionMaximize
	case Fullscreen:
		actions |= system.ActionFullscreen
	default:
		panic(fmt.Errorf("unknown WindowMode %v", m))
	}
	deco.Perform(actions)
	gtx := layout.Context{
		Ops:         o,
		Now:         e.Now,
		Source:      e.Source,
		Metric:      e.Metric,
		Constraints: layout.Exact(e.Size),
	}
	// Update the window based on the actions on the decorations.
	w.Perform(deco.Update(gtx))
	style.Layout(gtx)
	// Offset to place the frame content below the decorations.
	decoHeight := gtx.Dp(w.Decorations.Config.DecoHeight)
	if w.Decorations.currentHeight != decoHeight {
		w.Decorations.currentHeight = decoHeight
		w.out <- ConfigEvent{Config: w.effectiveConfig()}
	}
	e.Size.Y -= w.Decorations.currentHeight
	return e.Size, image.Pt(0, decoHeight)
}

func (w *Window) effectiveConfig() Config {
	cnf := w.Decorations.Config
	cnf.Size.Y -= w.Decorations.currentHeight
	cnf.Decorated = w.Decorations.Enabled || cnf.Decorated
	return cnf
}

// Perform the actions on the window.
func (w *Window) Perform(actions system.Action) {
	WalkActions(actions, func(action system.Action) {
		switch action {
		case system.ActionMinimize:
			w.Option(Minimized.Option())
		case system.ActionMaximize:
			w.Option(Maximized.Option())
		case system.ActionUnmaximize:
			w.Option(Windowed.Option())
		default:
			return
		}
		actions &^= action
	})
	if actions == 0 {
		return
	}
	for {
		select {
		case old := <-w.Actions:
			actions |= old
		case w.Actions <- actions:
			w.Wakeup()
			return
		}
	}
}

// Title sets the title of the window.
func Title(t string) Option {
	return func(_ unit.Metric, cnf *Config) {
		cnf.Title = t
	}
}

// Size sets the size of the window. The mode will be changed to Windowed.
func Size(w, h unit.Dp) Option {
	if w <= 0 {
		panic("width must be larger than or equal to 0")
	}
	if h <= 0 {
		panic("height must be larger than or equal to 0")
	}
	return func(m unit.Metric, cnf *Config) {
		cnf.Mode = Windowed
		cnf.Size = image.Point{
			X: m.Dp(w),
			Y: m.Dp(h),
		}
	}
}

// Decorated controls whether Gio and/or the platform are responsible
// for drawing window decorations. Providing false indicates that
// the application will either be undecorated or will draw its own decorations.
func Decorated(enabled bool) Option {
	return func(_ unit.Metric, cnf *Config) {
		cnf.Decorated = enabled
	}
}

func DecoHeightOpt(h unit.Dp) Option {
	return func(m unit.Metric, c *Config) {
		c.DecoHeight = h
	}
}

// flushEvent is sent to detect when the user program
// has completed processing of all prior events. Its an
// [io/event.Event] but only for internal use.
type flushEvent struct{}

func (t flushEvent) ImplementsEvent() {}

// theFlushEvent avoids allocating garbage when sending
// flushEvents.
var theFlushEvent flushEvent
