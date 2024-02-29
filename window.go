package mado

import (
	"image"
	"time"
	"unicode"
	"unicode/utf16"

	"github.com/kanryu/mado/f32"
	"github.com/kanryu/mado/io/event"
	"github.com/kanryu/mado/io/input"
	"github.com/kanryu/mado/io/key"
	"github.com/kanryu/mado/io/system"
	"github.com/kanryu/mado/op"
	"github.com/kanryu/mado/unit"
)

// Option configures a window.
type Option func(unit.Metric, *Config)

type EditorState struct {
	input.EditorState
	Compose key.Range
}

func (e *EditorState) Replace(r key.Range, text string) {
	if r.Start > r.End {
		r.Start, r.End = r.End, r.Start
	}
	runes := []rune(text)
	newEnd := r.Start + len(runes)
	adjust := func(pos int) int {
		switch {
		case newEnd < pos && pos <= r.End:
			return newEnd
		case r.End < pos:
			diff := newEnd - r.End
			return pos + diff
		}
		return pos
	}
	e.Selection.Start = adjust(e.Selection.Start)
	e.Selection.End = adjust(e.Selection.End)
	if e.Compose.Start != -1 {
		e.Compose.Start = adjust(e.Compose.Start)
		e.Compose.End = adjust(e.Compose.End)
	}
	s := e.Snippet
	if r.End < s.Start || r.Start > s.End {
		// Discard snippet if it doesn't overlap with replacement.
		s = key.Snippet{
			Range: key.Range{
				Start: r.Start,
				End:   r.Start,
			},
		}
	}
	var newSnippet []rune
	snippet := []rune(s.Text)
	// Append first part of existing snippet.
	if end := r.Start - s.Start; end > 0 {
		newSnippet = append(newSnippet, snippet[:end]...)
	}
	// Append replacement.
	newSnippet = append(newSnippet, runes...)
	// Append last part of existing snippet.
	if start := r.End; start < s.End {
		newSnippet = append(newSnippet, snippet[start-s.Start:]...)
	}
	// Adjust snippet range to include replacement.
	if r.Start < s.Start {
		s.Start = r.Start
	}
	s.End = s.Start + len(newSnippet)
	s.Text = string(newSnippet)
	e.Snippet = s
}

// UTF16Index converts the given index in runes into an index in utf16 characters.
func (e *EditorState) UTF16Index(runes int) int {
	if runes == -1 {
		return -1
	}
	if runes < e.Snippet.Start {
		// Assume runes before sippet are one UTF-16 character each.
		return runes
	}
	chars := e.Snippet.Start
	runes -= e.Snippet.Start
	for _, r := range e.Snippet.Text {
		if runes == 0 {
			break
		}
		runes--
		chars++
		if r1, _ := utf16.EncodeRune(r); r1 != unicode.ReplacementChar {
			chars++
		}
	}
	// Assume runes after snippets are one UTF-16 character each.
	return chars + runes
}

// RunesIndex converts the given index in utf16 characters to an index in runes.
func (e *EditorState) RunesIndex(chars int) int {
	if chars == -1 {
		return -1
	}
	if chars < e.Snippet.Start {
		// Assume runes before offset are one UTF-16 character each.
		return chars
	}
	runes := e.Snippet.Start
	chars -= e.Snippet.Start
	for _, r := range e.Snippet.Text {
		if chars == 0 {
			break
		}
		chars--
		runes++
		if r1, _ := utf16.EncodeRune(r); r1 != unicode.ReplacementChar {
			chars--
		}
	}
	// Assume runes after snippets are one UTF-16 character each.
	return runes + chars
}

type Callbacks interface {
	SetDriver(d Driver)
	Event(e event.Event) bool
	SemanticRoot() input.SemanticID
	LookupSemantic(semID input.SemanticID) (input.SemanticNode, bool)
	AppendSemanticDiffs(diffs []input.SemanticID) []input.SemanticID
	SemanticAt(pos f32.Point) (input.SemanticID, bool)
	EditorState() EditorState
	SetComposingRegion(r key.Range)
	EditorInsert(text string)
	EditorReplace(r key.Range, text string)
	SetEditorSelection(r key.Range)
	SetEditorSnippet(r key.Range)
	ClickFocus()
	ActionAt(p f32.Point) (system.Action, bool)
}

type Window interface {
	Update(frame *op.Ops)
	ValidateAndProcess(d Driver, size image.Point, sync bool, frame *op.Ops, sigChan chan<- struct{}) error
	Frame(frame *op.Ops, viewport image.Point) error
	ProcessFrame(d Driver)
	Invalidate()
	Option(opts ...Option)
	Run(f func())
	DriverDefer(f func(d Driver))
	UpdateAnimation(d Driver)
	Wakeup()
	SetNextFrame(at time.Time)
	WaitAck(d Driver)
	DestroyGPU()
	UpdateSemantics()
	CollectSemanticDiffs(diffs *[]input.SemanticID, n input.SemanticNode)
	UpdateState(d Driver)
	ProcessEvent(d Driver, e event.Event) bool
	NextEvent() event.Event
	UpdateCursor(d Driver)
	FallbackDecorate() bool
	Decorate(d Driver, e FrameEvent, o *op.Ops) (size, offset image.Point)
	EffectiveConfig() Config
	Perform(actions system.Action)
}

func DecoHeightOpt(h unit.Dp) Option {
	return func(m unit.Metric, c *Config) {
		c.DecoHeight = h
	}
}
