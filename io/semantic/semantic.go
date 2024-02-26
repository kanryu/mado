// SPDX-License-Identifier: Unlicense OR MIT

// Package semantic provides operations for semantic descriptions of a user
// interface, to facilitate presentation and interaction in external software
// such as screen readers.
//
// Semantic descriptions are organized in a tree, with clip operations as
// nodes. Operations in this package are associated with the current semantic
// node, that is the most recent pushed clip operation.
package semantic

import (
	"github.com/kanryu/mado/internal/ops"
	"github.com/kanryu/mado/op"
)

// LabelOp provides the content of a textual component.
type LabelOp string

// DescriptionOp describes a component.
type DescriptionOp string

// ClassOp provides the component class.
type ClassOp int

const (
	Unknown ClassOp = iota
	Button
	CheckBox
	Editor
	RadioButton
	Switch
)

// SelectedOp describes the selected state for components that have
// boolean state.
type SelectedOp bool

// EnabledOp describes the enabled state.
type EnabledOp bool

func (l LabelOp) Add(o *op.Ops) {
	data := ops.Write1String(&o.Internal, ops.TypeSemanticLabelLen, string(l))
	data[0] = byte(ops.TypeSemanticLabel)
}

func (d DescriptionOp) Add(o *op.Ops) {
	data := ops.Write1String(&o.Internal, ops.TypeSemanticDescLen, string(d))
	data[0] = byte(ops.TypeSemanticDesc)
}

func (c ClassOp) Add(o *op.Ops) {
	data := ops.Write(&o.Internal, ops.TypeSemanticClassLen)
	data[0] = byte(ops.TypeSemanticClass)
	data[1] = byte(c)
}

func (s SelectedOp) Add(o *op.Ops) {
	data := ops.Write(&o.Internal, ops.TypeSemanticSelectedLen)
	data[0] = byte(ops.TypeSemanticSelected)
	if s {
		data[1] = 1
	}
}

func (e EnabledOp) Add(o *op.Ops) {
	data := ops.Write(&o.Internal, ops.TypeSemanticEnabledLen)
	data[0] = byte(ops.TypeSemanticEnabled)
	if e {
		data[1] = 1
	}
}

func (c ClassOp) String() string {
	switch c {
	case Unknown:
		return "Unknown"
	case Button:
		return "Button"
	case CheckBox:
		return "CheckBox"
	case Editor:
		return "Editor"
	case RadioButton:
		return "RadioButton"
	case Switch:
		return "Switch"
	default:
		panic("invalid ClassOp")
	}
}
