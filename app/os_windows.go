//go:build windows
// +build windows

package app

import (
	"github.com/kanryu/mado/windows"
)

func init() {
	windows.InitWindows()
}
