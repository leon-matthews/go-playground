package main

import (
	"fmt"
	"math/rand"
	"time"
)

func main() {
	done := make(chan any)
	ints := randomStream(done)
	for range 10 {
		fmt.Println(<-ints)
	}
	close(done)

	// Pretend we're off doing something useful
	time.Sleep(1 * time.Second)
}

// randomStream writes random integers out to returned channel... forever!
func randomStream(done <-chan any) <-chan int {
	ints := make(chan int)
	go func() {
		defer fmt.Println("randomStream closure exited")
		defer close(ints)
		for {
			select {
			case ints <- rand.Int():
			case <- done:
				return
			}
		}
	}()
	return ints
}
