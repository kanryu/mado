//go:build windows
// +build windows

package app

import (
	"github.com/kanryu/mado/mswindows"
)

func init() {
	mswindows.InitWindows()
}
