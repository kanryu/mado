package main

import (
	"fmt"
	"reflect"
)

func main() {
	numChans := 5
	var chans []chan struct{}

	for i := 0; i < numChans; i++ {
		tmp := make(chan struct{})
		chans = append(chans, tmp)
		go DoFunc(tmp)
	}

	cases := make([]reflect.SelectCase, len(chans))
	for i, ch := range chans {
		cases[i] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(ch)}
	}
	remaining := len(cases)
	for remaining > 0 {
		chosen, _, ok := reflect.Select(cases)
		if !ok {
			cases[chosen].Chan = reflect.ValueOf(nil)
			remaining -= 1
			continue
		}
		fmt.Println(chosen)
	}
}

func DoFunc(c chan<- struct{}) {
	c <- struct{}{}
	close(c)
}
