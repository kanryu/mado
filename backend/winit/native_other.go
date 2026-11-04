//go:build !windows

package winit

import (
	"errors"
	"github.com/kanryu/mado"
)

var ErrNativeUnavailable = errors.New("mado/winit: native factory is only available on Windows")

type NativeFactory struct{}

func NewNativeFactory() (*NativeFactory, error) { return nil, ErrNativeUnavailable }

func NewFactory() (*NativeFactory, error) { return NewNativeFactory() }
func (*NativeFactory) CreateWindow(mado.WindowOptions) (mado.Window, error) {
	return nil, ErrNativeUnavailable
}
func (*NativeFactory) Run(func(mado.Event) error) error { return ErrNativeUnavailable }
func (*NativeFactory) Quit()                            {}
