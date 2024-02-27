package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/kanryu/mado"
	"github.com/kanryu/mado/app"
	"github.com/kanryu/mado/io/event"
)

func main() {
	var wg sync.WaitGroup
	ch1 := make(chan rune)
	ch2 := make(chan event.Event)
	done := make(chan struct{})

	go func() {
		for {
			select {
			case c := <-ch1:
				fmt.Printf("[R1] %c\n", c)
			case e := <-ch2:
				fmt.Println("[R2]", e)
				switch e2 := e.(type) {
				// case app.ViewEvent:
				// 	fmt.Println("app.ViewEvent", e2)
				case mado.ViewEvent:
					fmt.Println("mado.ViewEvent", e2)
				}
			case <-done:
				fmt.Println("Done!")
				return
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for c := 'A'; c <= 'C'; c++ {
			ch1 <- c
			time.Sleep(time.Second)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 1; i < 5; i++ {
			ch2 <- app.ViewEvent{}
			time.Sleep(500 * time.Millisecond)
		}
	}()

	wg.Wait()

	close(done)

	time.Sleep(time.Second)
}
