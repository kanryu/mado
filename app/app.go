// SPDX-License-Identifier: Unlicense OR MIT

package app

import (
	"image"

	"github.com/kanryu/mado"
	"github.com/kanryu/mado/layout"
	"github.com/kanryu/mado/op"
)

// // A FrameEvent requests a new frame in the form of a list of
// // operations that describes the window content.
// type FrameEvent struct {
// 	// Now is the current animation. Use Now instead of time.Now to
// 	// synchronize animation and to avoid the time.Now call overhead.
// 	Now time.Time
// 	// Metric converts device independent dp and sp to device pixels.
// 	Metric unit.Metric
// 	// Size is the dimensions of the window.
// 	Size image.Point
// 	// Insets represent the space occupied by system decorations and controls.
// 	Insets Insets
// 	// Frame completes the FrameEvent by drawing the graphical operations
// 	// from ops into the window.
// 	Frame func(frame *op.Ops)
// 	// Source is the interface between the window and widgets.
// 	Source input.Source
// }

// // Insets is the space taken up by
// // system decoration such as translucent
// // system bars and software keyboards.
// type Insets struct {
// 	// Values are in pixels.
// 	Top, Bottom, Left, Right unit.Dp
// }

// NewContext is shorthand for
//
//	layout.Context{
//	  Ops: ops,
//	  Now: e.Now,
//	  Source: e.Source,
//	  Metric: e.Metric,
//	  Constraints: layout.Exact(e.Size),
//	}
//
// NewContext calls ops.Reset and adjusts ops for e.Insets.
func NewContext(ops *op.Ops, e mado.FrameEvent) layout.Context {
	ops.Reset()

	size := e.Size

	if e.Insets != (mado.Insets{}) {
		left := e.Metric.Dp(e.Insets.Left)
		top := e.Metric.Dp(e.Insets.Top)
		op.Offset(image.Point{
			X: left,
			Y: top,
		}).Add(ops)

		size.X -= left + e.Metric.Dp(e.Insets.Right)
		size.Y -= top + e.Metric.Dp(e.Insets.Bottom)
	}

	return layout.Context{
		Ops:         ops,
		Now:         e.Now,
		Source:      e.Source,
		Metric:      e.Metric,
		Constraints: layout.Exact(size),
	}
}

// DataDir returns a path to use for application-specific
// configuration data.
// On desktop systems, DataDir use os.UserConfigDir.
// On iOS NSDocumentDirectory is queried.
// For Android Context.getFilesDir is used.
//
// BUG: DataDir blocks on Android until init functions
// have completed.
func DataDir() (string, error) {
	return dataDir()
}

// Main must be called last from the program main function.
// On most platforms Main blocks forever, for Android and
// iOS it returns immediately to give control of the main
// thread back to the system.
//
// Calling Main is necessary because some operating systems
// require control of the main thread of the program for
// running windows.
func Main() {
	mado.OsMain()
}

// func (FrameEvent) ImplementsEvent() {}

// func init() {
// 	if extraArgs != "" {
// 		args := strings.Split(extraArgs, "|")
// 		os.Args = append(os.Args, args...)
// 	}
// 	if ID == "" {
// 		ID = filepath.Base(os.Args[0])
// 	}
// }
