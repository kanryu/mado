// SPDX-License-Identifier: Unlicense OR MIT

package app

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"runtime"
	"time"

	"github.com/kanryu/mado"
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

	_ "github.com/kanryu/mado/app/internal/log"
)

// // Option configures a window.
// type Option func(unit.Metric, *Config)

var _ mado.Window = (*Window)(nil)

// Window represents an operating system window.
type Window struct {
	ctx mado.Context
	gpu gpu.GPU

	// driverFuncs is a channel of functions to run when
	// the Window has a valid driver.
	driverFuncs chan func(d mado.Driver)
	// wakeups wakes up the native event loop to send a
	// WakeupEvent that flushes driverFuncs.
	wakeups chan struct{}
	// wakeupFuncs is sent wakeup functions when the driver changes.
	wakeupFuncs chan func()
	// redraws is notified when a redraw is requested by the client.
	redraws chan struct{}
	// immediateRedraws is like redraw but doesn't need a wakeup.
	immediateRedraws chan struct{}
	// scheduledRedraws is sent the most recent delayed redraw time.
	scheduledRedraws chan time.Time
	// options are the options waiting to be applied.
	options chan []mado.Option
	// actions are the actions waiting to be performed.
	actions chan system.Action

	// out is where the platform backend delivers events bound for the
	// user program.
	out      chan event.Event
	frames   chan *op.Ops
	frameAck chan struct{}
	destroy  chan struct{}

	stage        Stage
	animating    bool
	hasNextFrame bool
	nextFrame    time.Time
	// viewport is the latest frame size with insets applied.
	viewport image.Rectangle
	// metric is the metric from the most recent frame.
	metric unit.Metric

	queue       input.Router
	cursor      pointer.Cursor
	decorations struct {
		op.Ops
		// enabled tracks the Decorated option as
		// given to the Option method. It may differ
		// from Config.Decorated depending on platform
		// capability.
		enabled bool
		mado.Config
		height        unit.Dp
		currentHeight int
		*material.Theme
		*widget.Decorations
	}

	callbacks callbacks

	nocontext bool

	// semantic data, lazily evaluated if requested by a backend to speed up
	// the cases where semantic data is not needed.
	semantic struct {
		// uptodate tracks whether the fields below are up to date.
		uptodate bool
		root     input.SemanticID
		prevTree []input.SemanticNode
		tree     []input.SemanticNode
		ids      map[input.SemanticID]input.SemanticNode
	}

	imeState mado.EditorState

	// event stores the state required for processing and delivering events
	// from NextEvent. If we had support for range over func, this would
	// be the iterator state.
	eventState struct {
		created     bool
		initialOpts []mado.Option
		wakeup      func()
		timer       *time.Timer
	}
}

// type editorState struct {
// 	input.EditorState
// 	compose key.Range
// }

// var _ mado.Callbacks = (*callbacks)(nil)

// type callbacks struct {
// 	w          *Window
// 	d          mado.Driver
// 	busy       bool
// 	waitEvents []event.Event
// }

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
func NewWindow(options ...mado.Option) *Window {
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
	defaultOptions := []mado.Option{
		Size(800, 600),
		Title("Gio"),
		Decorated(true),
		decoHeightOpt(decoHeight),
	}
	options = append(defaultOptions, options...)
	var cnf mado.Config
	cnf.Apply(unit.Metric{}, options)

	w := &Window{
		out:              make(chan event.Event),
		immediateRedraws: make(chan struct{}),
		redraws:          make(chan struct{}, 1),
		scheduledRedraws: make(chan time.Time, 1),
		frames:           make(chan *op.Ops),
		frameAck:         make(chan struct{}),
		driverFuncs:      make(chan func(d mado.Driver), 1),
		wakeups:          make(chan struct{}, 1),
		wakeupFuncs:      make(chan func()),
		destroy:          make(chan struct{}),
		options:          make(chan []mado.Option, 1),
		actions:          make(chan system.Action, 1),
		nocontext:        cnf.CustomRenderer,
	}
	w.decorations.Theme = theme
	w.decorations.Decorations = deco
	w.decorations.enabled = cnf.Decorated
	w.decorations.height = decoHeight
	w.imeState.Compose = key.Range{Start: -1, End: -1}
	w.semantic.ids = make(map[input.SemanticID]input.SemanticNode)
	w.callbacks.w = w
	w.eventState.initialOpts = options
	return w
}

func decoHeightOpt(h unit.Dp) mado.Option {
	return func(m unit.Metric, c *mado.Config) {
		c.DecoHeight = h
	}
}

// update the window contents, input operations declare input handlers,
// and so on. The supplied operations list completely replaces the window state
// from previous calls.
func (w *Window) Update(frame *op.Ops) {
	w.frames <- frame
	<-w.frameAck
}

func (w *Window) ValidateAndProcess(d mado.Driver, size image.Point, sync bool, frame *op.Ops, sigChan chan<- struct{}) error {
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
				if errors.Is(err, errOutOfDate) {
					// Surface couldn't be created for transient reasons. Skip
					// this frame and wait for the next.
					return nil
				}
				w.DestroyGPU()
				if errors.Is(err, gpu.ErrDeviceLost) {
					continue
				}
				return err
			}
		}
		if w.ctx != nil {
			if err := w.ctx.Lock(); err != nil {
				w.DestroyGPU()
				return err
			}
		}
		if w.gpu == nil && !w.nocontext {
			gpu, err := gpu.New(w.ctx.API())
			if err != nil {
				w.ctx.Unlock()
				w.DestroyGPU()
				return err
			}
			w.gpu = gpu
		}
		if w.gpu != nil {
			if err := w.Frame(frame, size); err != nil {
				w.ctx.Unlock()
				if errors.Is(err, errOutOfDate) {
					// GPU surface needs refreshing.
					sync = true
					continue
				}
				w.DestroyGPU()
				if errors.Is(err, gpu.ErrDeviceLost) {
					continue
				}
				return err
			}
		}
		w.queue.Frame(frame)
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

func (w *Window) Frame(frame *op.Ops, viewport image.Point) error {
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

func (w *Window) ProcessFrame(d mado.Driver) {
	for k := range w.semantic.ids {
		delete(w.semantic.ids, k)
	}
	w.semantic.uptodate = false
	q := &w.queue
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
	oldState := w.imeState
	newState := oldState
	newState.EditorState = q.EditorState()
	if newState != oldState {
		w.imeState = newState
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
func (w *Window) Option(opts ...mado.Option) {
	if len(opts) == 0 {
		return
	}
	for {
		select {
		case old := <-w.options:
			opts = append(old, opts...)
		case w.options <- opts:
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
	w.DriverDefer(func(d mado.Driver) {
		defer close(done)
		f()
	})
	select {
	case <-done:
	case <-w.destroy:
	}
}

// driverDefer is like Run but can be run from any context. It doesn't wait
// for f to return.
func (w *Window) DriverDefer(f func(d mado.Driver)) {
	select {
	case w.driverFuncs <- f:
		w.Wakeup()
	case <-w.destroy:
	}
}

func (w *Window) UpdateAnimation(d mado.Driver) {
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

// func (c *callbacks) SetWindow(w *Window) {
// 	c.w = w
// }

// func (c *callbacks) SetDriver(d mado.Driver) {
// 	c.d = d
// 	var wakeup func()
// 	if d != nil {
// 		wakeup = d.Wakeup
// 	}
// 	c.w.wakeupFuncs <- wakeup
// }

// func (c *callbacks) Event(e event.Event) bool {
// 	if c.d == nil {
// 		panic("event while no driver active")
// 	}
// 	c.waitEvents = append(c.waitEvents, e)
// 	if c.busy {
// 		return true
// 	}
// 	c.busy = true
// 	var handled bool
// 	for len(c.waitEvents) > 0 {
// 		e := c.waitEvents[0]
// 		copy(c.waitEvents, c.waitEvents[1:])
// 		c.waitEvents = c.waitEvents[:len(c.waitEvents)-1]
// 		handled = c.w.ProcessEvent(c.d, e)
// 	}
// 	c.busy = false
// 	select {
// 	case <-c.w.destroy:
// 		return handled
// 	default:
// 	}
// 	c.w.UpdateState(c.d)
// 	if _, ok := e.(mado.WakeupEvent); ok {
// 		select {
// 		case opts := <-c.w.options:
// 			cnf := mado.Config{Decorated: c.w.decorations.enabled}
// 			for _, opt := range opts {
// 				opt(c.w.metric, &cnf)
// 			}
// 			c.w.decorations.enabled = cnf.Decorated
// 			decoHeight := c.w.decorations.height
// 			if !c.w.decorations.enabled {
// 				decoHeight = 0
// 			}
// 			opts = append(opts, decoHeightOpt(decoHeight))
// 			c.d.Configure(opts)
// 		default:
// 		}
// 		select {
// 		case acts := <-c.w.actions:
// 			c.d.Perform(acts)
// 		default:
// 		}
// 	}
// 	return handled
// }

// // SemanticRoot returns the ID of the semantic root.
// func (c *callbacks) SemanticRoot() input.SemanticID {
// 	c.w.UpdateSemantics()
// 	return c.w.semantic.root
// }

// // LookupSemantic looks up a semantic node from an ID. The zero ID denotes the root.
// func (c *callbacks) LookupSemantic(semID input.SemanticID) (input.SemanticNode, bool) {
// 	c.w.UpdateSemantics()
// 	n, found := c.w.semantic.ids[semID]
// 	return n, found
// }

// func (c *callbacks) AppendSemanticDiffs(diffs []input.SemanticID) []input.SemanticID {
// 	c.w.UpdateSemantics()
// 	if tree := c.w.semantic.prevTree; len(tree) > 0 {
// 		c.w.CollectSemanticDiffs(&diffs, c.w.semantic.prevTree[0])
// 	}
// 	return diffs
// }

// func (c *callbacks) SemanticAt(pos f32.Point) (input.SemanticID, bool) {
// 	c.w.UpdateSemantics()
// 	return c.w.queue.SemanticAt(pos)
// }

// func (c *callbacks) EditorState() mado.EditorState {
// 	return c.w.imeState
// }

// func (c *callbacks) SetComposingRegion(r key.Range) {
// 	c.w.imeState.Compose = r
// }

// func (c *callbacks) EditorInsert(text string) {
// 	sel := c.w.imeState.Selection.Range
// 	c.EditorReplace(sel, text)
// 	start := sel.Start
// 	if sel.End < start {
// 		start = sel.End
// 	}
// 	sel.Start = start + utf8.RuneCountInString(text)
// 	sel.End = sel.Start
// 	c.SetEditorSelection(sel)
// }

// func (c *callbacks) EditorReplace(r key.Range, text string) {
// 	c.w.imeState.Replace(r, text)
// 	c.Event(key.EditEvent{Range: r, Text: text})
// 	c.Event(key.SnippetEvent(c.w.imeState.Snippet.Range))
// }

// func (c *callbacks) SetEditorSelection(r key.Range) {
// 	c.w.imeState.Selection.Range = r
// 	c.Event(key.SelectionEvent(r))
// }

// func (c *callbacks) SetEditorSnippet(r key.Range) {
// 	if sn := c.EditorState().Snippet.Range; sn == r {
// 		// No need to expand.
// 		return
// 	}
// 	c.Event(key.SnippetEvent(r))
// }

func (w *Window) moveFocus(dir key.FocusDirection) {
	w.queue.MoveFocus(dir)
	if _, handled := w.queue.WakeupTime(); handled {
		w.queue.RevealFocus(w.viewport)
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
		dist := v.Mul(int(w.metric.Dp(scrollABit)))
		w.queue.ScrollFocus(dist)
	}
}

// func (c *callbacks) ClickFocus() {
// 	c.w.queue.ClickFocus()
// 	c.w.SetNextFrame(time.Time{})
// 	c.w.UpdateAnimation(c.d)
// }

// func (c *callbacks) ActionAt(p f32.Point) (system.Action, bool) {
// 	return c.w.queue.ActionAt(p)
// }

// func (e *editorState) Replace(r key.Range, text string) {
// 	if r.Start > r.End {
// 		r.Start, r.End = r.End, r.Start
// 	}
// 	runes := []rune(text)
// 	newEnd := r.Start + len(runes)
// 	adjust := func(pos int) int {
// 		switch {
// 		case newEnd < pos && pos <= r.End:
// 			return newEnd
// 		case r.End < pos:
// 			diff := newEnd - r.End
// 			return pos + diff
// 		}
// 		return pos
// 	}
// 	e.Selection.Start = adjust(e.Selection.Start)
// 	e.Selection.End = adjust(e.Selection.End)
// 	if e.compose.Start != -1 {
// 		e.compose.Start = adjust(e.compose.Start)
// 		e.compose.End = adjust(e.compose.End)
// 	}
// 	s := e.Snippet
// 	if r.End < s.Start || r.Start > s.End {
// 		// Discard snippet if it doesn't overlap with replacement.
// 		s = key.Snippet{
// 			Range: key.Range{
// 				Start: r.Start,
// 				End:   r.Start,
// 			},
// 		}
// 	}
// 	var newSnippet []rune
// 	snippet := []rune(s.Text)
// 	// Append first part of existing snippet.
// 	if end := r.Start - s.Start; end > 0 {
// 		newSnippet = append(newSnippet, snippet[:end]...)
// 	}
// 	// Append replacement.
// 	newSnippet = append(newSnippet, runes...)
// 	// Append last part of existing snippet.
// 	if start := r.End; start < s.End {
// 		newSnippet = append(newSnippet, snippet[start-s.Start:]...)
// 	}
// 	// Adjust snippet range to include replacement.
// 	if r.Start < s.Start {
// 		s.Start = r.Start
// 	}
// 	s.End = s.Start + len(newSnippet)
// 	s.Text = string(newSnippet)
// 	e.Snippet = s
// }

// // UTF16Index converts the given index in runes into an index in utf16 characters.
// func (e *editorState) UTF16Index(runes int) int {
// 	if runes == -1 {
// 		return -1
// 	}
// 	if runes < e.Snippet.Start {
// 		// Assume runes before sippet are one UTF-16 character each.
// 		return runes
// 	}
// 	chars := e.Snippet.Start
// 	runes -= e.Snippet.Start
// 	for _, r := range e.Snippet.Text {
// 		if runes == 0 {
// 			break
// 		}
// 		runes--
// 		chars++
// 		if r1, _ := utf16.EncodeRune(r); r1 != unicode.ReplacementChar {
// 			chars++
// 		}
// 	}
// 	// Assume runes after snippets are one UTF-16 character each.
// 	return chars + runes
// }

// // RunesIndex converts the given index in utf16 characters to an index in runes.
// func (e *editorState) RunesIndex(chars int) int {
// 	if chars == -1 {
// 		return -1
// 	}
// 	if chars < e.Snippet.Start {
// 		// Assume runes before offset are one UTF-16 character each.
// 		return chars
// 	}
// 	runes := e.Snippet.Start
// 	chars -= e.Snippet.Start
// 	for _, r := range e.Snippet.Text {
// 		if chars == 0 {
// 			break
// 		}
// 		chars--
// 		runes++
// 		if r1, _ := utf16.EncodeRune(r); r1 != unicode.ReplacementChar {
// 			chars--
// 		}
// 	}
// 	// Assume runes after snippets are one UTF-16 character each.
// 	return runes + chars
// }

func (w *Window) WaitAck(d mado.Driver) {
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

func (w *Window) DestroyGPU() {
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
func (w *Window) waitFrame(d mado.Driver) *op.Ops {
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

// updateSemantics refreshes the semantics tree, the id to node map and the ids of
// updated nodes.
func (w *Window) UpdateSemantics() {
	if w.semantic.uptodate {
		return
	}
	w.semantic.uptodate = true
	w.semantic.prevTree, w.semantic.tree = w.semantic.tree, w.semantic.prevTree
	w.semantic.tree = w.queue.AppendSemantics(w.semantic.tree[:0])
	w.semantic.root = w.semantic.tree[0].ID
	for _, n := range w.semantic.tree {
		w.semantic.ids[n.ID] = n
	}
}

// CollectSemanticDiffs traverses the previous semantic tree, noting changed nodes.
func (w *Window) CollectSemanticDiffs(diffs *[]input.SemanticID, n input.SemanticNode) {
	newNode, exists := w.semantic.ids[n.ID]
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

func (w *Window) UpdateState(d mado.Driver) {
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

func (w *Window) ProcessEvent(d mado.Driver, e event.Event) bool {
	select {
	case <-w.destroy:
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
		w.WaitAck(d)
	case mado.FrameEvent:
		if e2.Size == (image.Point{}) {
			panic(errors.New("internal error: zero-sized Draw"))
		}
		if w.stage < StageInactive {
			// No drawing if not visible.
			break
		}
		w.metric = e2.Metric
		w.hasNextFrame = false
		e2.Frame = w.Update
		e2.Source = w.queue.Source()

		// Prepare the decorations and update the frame insets.
		wrapper := &w.decorations.Ops
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
			w.queue.RevealFocus(viewport)
		}
		w.viewport = viewport
		viewSize := e2.Size
		m := op.Record(wrapper)
		size, offset := w.Decorate(d, e2, wrapper)
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
		if err := w.ValidateAndProcess(d, viewSize, e2.Sync, wrapper, signal); err != nil {
			w.DestroyGPU()
			w.out <- mado.DestroyEvent{Err: err}
			close(w.destroy)
			break
		}
		w.ProcessFrame(d)
		w.UpdateCursor(d)
	case mado.DestroyEvent:
		w.DestroyGPU()
		w.out <- e2
		close(w.destroy)
	case ViewEvent:
		w.out <- e2
		w.WaitAck(d)
	case mado.ConfigEvent:
		w.decorations.Config = e2.Config
		e2.Config = w.EffectiveConfig()
		w.out <- e2
	case mado.WakeupEvent:
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
		w.queue.Queue(e)
		t, handled := w.queue.WakeupTime()
		if focusDir != -1 && !handled {
			w.moveFocus(focusDir)
			t, handled = w.queue.WakeupTime()
		}
		w.UpdateCursor(d)
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
	state := &w.eventState
	if !state.created {
		state.created = true
		if err := newWindow(&w.callbacks, state.initialOpts); err != nil {
			close(w.destroy)
			return mado.DestroyEvent{Err: err}
		}
	}
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
		case state.wakeup = <-w.wakeupFuncs:
		}
	}
}

func (w *Window) UpdateCursor(d mado.Driver) {
	if c := w.queue.Cursor(); c != w.cursor {
		w.cursor = c
		d.SetCursor(c)
	}
}

func (w *Window) FallbackDecorate() bool {
	cnf := w.decorations.Config
	return w.decorations.enabled && !cnf.Decorated && cnf.Mode != mado.Fullscreen && !w.nocontext
}

// decorate the window if enabled and returns the corresponding Insets.
func (w *Window) Decorate(d mado.Driver, e mado.FrameEvent, o *op.Ops) (size, offset image.Point) {
	if !w.FallbackDecorate() {
		return e.Size, image.Pt(0, 0)
	}
	deco := w.decorations.Decorations
	allActions := system.ActionMinimize | system.ActionMaximize | system.ActionUnmaximize |
		system.ActionClose | system.ActionMove
	style := material.Decorations(w.decorations.Theme, deco, allActions, w.decorations.Config.Title)
	// Update the decorations based on the current window mode.
	var actions system.Action
	switch m := w.decorations.Config.Mode; m {
	case mado.Windowed:
		actions |= system.ActionUnmaximize
	case mado.Minimized:
		actions |= system.ActionMinimize
	case mado.Maximized:
		actions |= system.ActionMaximize
	case mado.Fullscreen:
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
	decoHeight := gtx.Dp(w.decorations.Config.DecoHeight)
	if w.decorations.currentHeight != decoHeight {
		w.decorations.currentHeight = decoHeight
		w.out <- mado.ConfigEvent{Config: w.EffectiveConfig()}
	}
	e.Size.Y -= w.decorations.currentHeight
	return e.Size, image.Pt(0, decoHeight)
}

func (w *Window) EffectiveConfig() mado.Config {
	cnf := w.decorations.Config
	cnf.Size.Y -= w.decorations.currentHeight
	cnf.Decorated = w.decorations.enabled || cnf.Decorated
	return cnf
}

// Perform the actions on the window.
func (w *Window) Perform(actions system.Action) {
	mado.WalkActions(actions, func(action system.Action) {
		switch action {
		case system.ActionMinimize:
			w.Option(mado.Minimized.Option())
		case system.ActionMaximize:
			w.Option(mado.Maximized.Option())
		case system.ActionUnmaximize:
			w.Option(mado.Windowed.Option())
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
		case old := <-w.actions:
			actions |= old
		case w.actions <- actions:
			w.Wakeup()
			return
		}
	}
}

// Title sets the title of the window.
func Title(t string) mado.Option {
	return func(_ unit.Metric, cnf *mado.Config) {
		cnf.Title = t
	}
}

// Size sets the size of the window. The mode will be changed to Windowed.
func Size(w, h unit.Dp) mado.Option {
	if w <= 0 {
		panic("width must be larger than or equal to 0")
	}
	if h <= 0 {
		panic("height must be larger than or equal to 0")
	}
	return func(m unit.Metric, cnf *mado.Config) {
		cnf.Mode = mado.Windowed
		cnf.Size = image.Point{
			X: m.Dp(w),
			Y: m.Dp(h),
		}
	}
}

// MaxSize sets the maximum size of the window.
func MaxSize(w, h unit.Dp) mado.Option {
	if w <= 0 {
		panic("width must be larger than or equal to 0")
	}
	if h <= 0 {
		panic("height must be larger than or equal to 0")
	}
	return func(m unit.Metric, cnf *mado.Config) {
		cnf.MaxSize = image.Point{
			X: m.Dp(w),
			Y: m.Dp(h),
		}
	}
}

// MinSize sets the minimum size of the window.
func MinSize(w, h unit.Dp) mado.Option {
	if w <= 0 {
		panic("width must be larger than or equal to 0")
	}
	if h <= 0 {
		panic("height must be larger than or equal to 0")
	}
	return func(m unit.Metric, cnf *mado.Config) {
		cnf.MinSize = image.Point{
			X: m.Dp(w),
			Y: m.Dp(h),
		}
	}
}

// StatusColor sets the color of the Android status bar.
func StatusColor(color color.NRGBA) mado.Option {
	return func(_ unit.Metric, cnf *mado.Config) {
		cnf.StatusColor = color
	}
}

// NavigationColor sets the color of the navigation bar on Android, or the address bar in browsers.
func NavigationColor(color color.NRGBA) mado.Option {
	return func(_ unit.Metric, cnf *mado.Config) {
		cnf.NavigationColor = color
	}
}

// CustomRenderer controls whether the window contents is
// rendered by the client. If true, no GPU context is created.
//
// Caller must assume responsibility for rendering which includes
// initializing the render backend, swapping the framebuffer and
// handling frame pacing.
func CustomRenderer(custom bool) mado.Option {
	return func(_ unit.Metric, cnf *mado.Config) {
		cnf.CustomRenderer = custom
	}
}

// Decorated controls whether Gio and/or the platform are responsible
// for drawing window decorations. Providing false indicates that
// the application will either be undecorated or will draw its own decorations.
func Decorated(enabled bool) mado.Option {
	return func(_ unit.Metric, cnf *mado.Config) {
		cnf.Decorated = enabled
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
