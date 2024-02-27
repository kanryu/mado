package mado

import "github.com/kanryu/mado/io/system"

type WakeupEvent struct{}

func (WakeupEvent) ImplementsEvent() {}

func WalkActions(actions system.Action, do func(system.Action)) {
	for a := system.Action(1); actions != 0; a <<= 1 {
		if actions&a != 0 {
			actions &^= a
			do(a)
		}
	}
}
