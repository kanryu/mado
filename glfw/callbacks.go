package glfw

import (
	"time"
	"unicode/utf8"

	"github.com/kanryu/mado"
	"github.com/kanryu/mado/app"
	"github.com/kanryu/mado/f32"
	"github.com/kanryu/mado/io/event"
	"github.com/kanryu/mado/io/input"
	"github.com/kanryu/mado/io/key"
	"github.com/kanryu/mado/io/pointer"
	"github.com/kanryu/mado/io/system"
	"github.com/kanryu/mado/io/window"
)

var _ mado.Callbacks = (*Callbacks)(nil)

type Callbacks struct {
	W               *app.Window
	Gw              *Window
	D               mado.Driver
	Busy            bool
	PrevWindowMode  mado.WindowMode
	PrevWindowStage mado.Stage
	WaitEvents      []event.Event
}

func (c *Callbacks) SetWindow(w mado.Window) {
	c.W = w.(*app.Window)
}

func (c *Callbacks) SetGlfwWindow(gw *Window) {
	c.Gw = gw
}

func (c *Callbacks) SetDriver(d mado.Driver) {
	c.D = d
	var wakeup func()
	if d != nil {
		wakeup = d.Wakeup
	}
	c.W.WakeupFuncs <- wakeup
}

func (c *Callbacks) Event(e event.Event) bool {
	if c.D == nil {
		panic("event while no driver active")
	}
	c.WaitEvents = append(c.WaitEvents, e)
	if c.Busy {
		return true
	}
	c.Busy = true
	var handled bool
	for len(c.WaitEvents) > 0 {
		e := c.WaitEvents[0]
		// copy(c.WaitEvents, c.WaitEvents[1:])
		// c.WaitEvents = c.WaitEvents[:len(c.WaitEvents)-1]
		c.WaitEvents = append([]event.Event{}, c.WaitEvents[1:]...)
		handled = c.W.ProcessEvent(c.D, e)
		// POST events to glfw callbacks
		switch e2 := e.(type) {
		case mado.StageEvent:
			switch e2.Stage {
			case mado.StagePaused:
				if c.W.Decorations.Config.Mode == mado.Minimized {
					c.Gw.fIconifyHolder(c.Gw, true)
					c.PrevWindowMode = mado.Minimized
				}
			case mado.StageRunning:
				if c.W.Decorations.Config.Mode == mado.Maximized {
					c.Gw.fMaximizeHolder(c.Gw, true)
					c.PrevWindowMode = mado.Maximized
				}
				if c.W.Decorations.Config.Mode == mado.Windowed {
					if c.PrevWindowStage == mado.StageInactive {
						c.Gw.fFocusHolder(c.Gw, true)
					} else if c.PrevWindowMode == mado.Minimized {
						c.Gw.fIconifyHolder(c.Gw, false)
					} else if c.PrevWindowMode != mado.Windowed {
						c.Gw.fMaximizeHolder(c.Gw, false)
					}
					c.PrevWindowMode = mado.Windowed
				}
			case mado.StageInactive:
				c.Gw.fFocusHolder(c.Gw, false)
				c.PrevWindowStage = mado.StageInactive
			}
		case window.MoveEvent:
			c.Gw.fPosHolder(c.Gw, e2.Pos.X, e2.Pos.Y)
		case window.SizeEvent:
			c.Gw.fSizeHolder(c.Gw, e2.Size.X, e2.Size.Y)
		case pointer.Event:
			c.Gw.fMouseButtonHolder(c.Gw, MouseButton(e2.Buttons), Action(e2.Kind), ModifierKey(e2.Modifiers))
		case key.Event:
			c.Gw.fKeyHolder(c.Gw, Key(e2.KeyCode), 0, Action(e2.State), ModifierKey(e2.Modifiers))
		}
	}
	c.Busy = false
	select {
	case <-c.W.Destroy:
		return handled
	default:
	}
	c.W.UpdateState(c.D)
	if _, ok := e.(mado.WakeupEvent); ok {
		select {
		case opts := <-c.W.Options:
			cnf := mado.Config{Decorated: c.W.Decorations.Enabled}
			for _, opt := range opts {
				opt(c.W.Metric, &cnf)
			}
			c.W.Decorations.Enabled = cnf.Decorated
			decoHeight := c.W.Decorations.Height
			if !c.W.Decorations.Enabled {
				decoHeight = 0
			}
			opts = append(opts, mado.DecoHeightOpt(decoHeight))
			c.D.Configure(opts)
		default:
		}
		select {
		case acts := <-c.W.Actions:
			c.D.Perform(acts)
		default:
		}
	}
	return handled
}

// SemanticRoot returns the ID of the semantic root.
func (c *Callbacks) SemanticRoot() input.SemanticID {
	c.W.UpdateSemantics()
	return c.W.Semantic.Root
}

// LookupSemantic looks up a semantic node from an ID. The zero ID denotes the root.
func (c *Callbacks) LookupSemantic(semID input.SemanticID) (input.SemanticNode, bool) {
	c.W.UpdateSemantics()
	n, found := c.W.Semantic.Ids[semID]
	return n, found
}

func (c *Callbacks) AppendSemanticDiffs(diffs []input.SemanticID) []input.SemanticID {
	c.W.UpdateSemantics()
	if tree := c.W.Semantic.PrevTree; len(tree) > 0 {
		c.W.CollectSemanticDiffs(&diffs, c.W.Semantic.PrevTree[0])
	}
	return diffs
}

func (c *Callbacks) SemanticAt(pos f32.Point) (input.SemanticID, bool) {
	c.W.UpdateSemantics()
	return c.W.Queue.SemanticAt(pos)
}

func (c *Callbacks) EditorState() mado.EditorState {
	return c.W.ImeState
}

func (c *Callbacks) SetComposingRegion(r key.Range) {
	c.W.ImeState.Compose = r
}

func (c *Callbacks) EditorInsert(text string) {
	sel := c.W.ImeState.Selection.Range
	c.EditorReplace(sel, text)
	start := sel.Start
	if sel.End < start {
		start = sel.End
	}
	sel.Start = start + utf8.RuneCountInString(text)
	sel.End = sel.Start
	c.SetEditorSelection(sel)
}

func (c *Callbacks) EditorReplace(r key.Range, text string) {
	c.W.ImeState.Replace(r, text)
	c.Event(key.EditEvent{Range: r, Text: text})
	c.Event(key.SnippetEvent(c.W.ImeState.Snippet.Range))
}

func (c *Callbacks) SetEditorSelection(r key.Range) {
	c.W.ImeState.Selection.Range = r
	c.Event(key.SelectionEvent(r))
}

func (c *Callbacks) SetEditorSnippet(r key.Range) {
	if sn := c.EditorState().Snippet.Range; sn == r {
		// No need to expand.
		return
	}
	c.Event(key.SnippetEvent(r))
}

func (c *Callbacks) ClickFocus() {
	c.W.Queue.ClickFocus()
	c.W.SetNextFrame(time.Time{})
	c.W.UpdateAnimation(c.D)
}

func (c *Callbacks) ActionAt(p f32.Point) (system.Action, bool) {
	return c.W.Queue.ActionAt(p)
}
