//go:build (linux && !android) || freebsd || openbsd
// +build linux,!android freebsd openbsd

package app

import (
	"github.com/kanryu/mado/unix"
)

func init() {
	unix.InitUnix()
}
