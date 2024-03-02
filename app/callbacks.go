package app

import (
	"time"
	"unicode/utf8"

	"github.com/kanryu/mado"
	"github.com/kanryu/mado/f32"
	"github.com/kanryu/mado/io/event"
	"github.com/kanryu/mado/io/input"
	"github.com/kanryu/mado/io/key"
	"github.com/kanryu/mado/io/system"
)

var _ mado.Callbacks = (*Callbacks)(nil)

type Callbacks struct {
	w          *Window
	d          mado.Driver
	busy       bool
	waitEvents []event.Event
}

func (c *Callbacks) SetWindow(w mado.Window) {
	c.w = w.(*Window)
}

func (c *Callbacks) SetDriver(d mado.Driver) {
	c.d = d
	var wakeup func()
	if d != nil {
		wakeup = d.Wakeup
	}
	c.w.WakeupFuncs <- wakeup
}

func (c *Callbacks) Event(e event.Event) bool {
	if c.d == nil {
		panic("event while no driver active")
	}
	c.waitEvents = append(c.waitEvents, e)
	if c.busy {
		return true
	}
	c.busy = true
	var handled bool
	for len(c.waitEvents) > 0 {
		e := c.waitEvents[0]
		copy(c.waitEvents, c.waitEvents[1:])
		c.waitEvents = c.waitEvents[:len(c.waitEvents)-1]
		handled = c.w.ProcessEvent(c.d, e)
	}
	c.busy = false
	select {
	case <-c.w.Destroy:
		return handled
	default:
	}
	c.w.UpdateState(c.d)
	if _, ok := e.(mado.WakeupEvent); ok {
		select {
		case opts := <-c.w.Options:
			cnf := mado.Config{Decorated: c.w.Decorations.Enabled}
			for _, opt := range opts {
				opt(c.w.Metric, &cnf)
			}
			c.w.Decorations.Enabled = cnf.Decorated
			decoHeight := c.w.Decorations.Height
			if !c.w.Decorations.Enabled {
				decoHeight = 0
			}
			opts = append(opts, mado.DecoHeightOpt(decoHeight))
			c.d.Configure(opts)
		default:
		}
		select {
		case acts := <-c.w.Actions:
			c.d.Perform(acts)
		default:
		}
	}
	return handled
}

// SemanticRoot returns the ID of the semantic root.
func (c *Callbacks) SemanticRoot() input.SemanticID {
	c.w.UpdateSemantics()
	return c.w.Semantic.Root
}

// LookupSemantic looks up a semantic node from an ID. The zero ID denotes the root.
func (c *Callbacks) LookupSemantic(semID input.SemanticID) (input.SemanticNode, bool) {
	c.w.UpdateSemantics()
	n, found := c.w.Semantic.Ids[semID]
	return n, found
}

func (c *Callbacks) AppendSemanticDiffs(diffs []input.SemanticID) []input.SemanticID {
	c.w.UpdateSemantics()
	if tree := c.w.Semantic.PrevTree; len(tree) > 0 {
		c.w.CollectSemanticDiffs(&diffs, c.w.Semantic.PrevTree[0])
	}
	return diffs
}

func (c *Callbacks) SemanticAt(pos f32.Point) (input.SemanticID, bool) {
	c.w.UpdateSemantics()
	return c.w.Queue.SemanticAt(pos)
}

func (c *Callbacks) EditorState() mado.EditorState {
	return c.w.ImeState
}

func (c *Callbacks) SetComposingRegion(r key.Range) {
	c.w.ImeState.Compose = r
}

func (c *Callbacks) EditorInsert(text string, preedit bool) {
	sel := c.w.ImeState.Selection.Range
	c.EditorReplace(sel, text, preedit)
	start := sel.Start
	if sel.End < start {
		start = sel.End
	}
	sel.Start = start + utf8.RuneCountInString(text)
	sel.End = sel.Start
	c.SetEditorSelection(sel)
}

func (c *Callbacks) EditorReplace(r key.Range, text string, preedit bool) {
	c.w.ImeState.Replace(r, text)
	c.Event(key.EditEvent{Range: r, Text: text, Preedit: preedit})
	c.Event(key.SnippetEvent(c.w.ImeState.Snippet.Range))
}

func (c *Callbacks) SetEditorSelection(r key.Range) {
	c.w.ImeState.Selection.Range = r
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
	c.w.Queue.ClickFocus()
	c.w.SetNextFrame(time.Time{})
	c.w.UpdateAnimation(c.d)
}

func (c *Callbacks) ActionAt(p f32.Point) (system.Action, bool) {
	return c.w.Queue.ActionAt(p)
}
