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
	"github.com/kanryu/mado/io/system"
)

var _ mado.Callbacks = (*Callbacks)(nil)

type Callbacks struct {
	W          *app.Window
	D          mado.Driver
	Busy       bool
	WaitEvents []event.Event
}

func (c *Callbacks) SetWindow(w *app.Window) {
	c.W = w
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
		copy(c.WaitEvents, c.WaitEvents[1:])
		c.WaitEvents = c.WaitEvents[:len(c.WaitEvents)-1]
		handled = c.W.ProcessEvent(c.D, e)
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
