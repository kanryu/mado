//go:build darwin && !ios
// +build darwin,!ios

package app

import (
	"github.com/kanryu/mado/darwin"
)

func init() {
	darwin.InitDarwin()
}
