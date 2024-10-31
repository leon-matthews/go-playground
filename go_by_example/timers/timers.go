package main

import (
	"fmt"
	"time"
)

func main() {
	timerExample()
	tickerExample()
}

// When the Timer expires, the current time will be sent on C
func timerExample() {
	// Can be used like [time.Sleep]
	timer1 := time.NewTimer(500 * time.Millisecond)
	<-timer1.C // blocks on the timer's channel C until it sends a value

	// Can also be cancelled before firing
	timer2 := time.NewTimer(1 * time.Second)

	go func() {
		<-timer2.C
		fmt.Println("timer2 fired")
	}()

	stop2 := timer2.Stop()
	if stop2 {
		fmt.Println("timer2 stopped")
	}

	fmt.Println("Finished")
}

// Will send the current time on the channel after each tick
func tickerExample() {
	ticker := time.NewTicker(500 * time.Millisecond)
	done := make(chan bool)

	go func() {
		for {
			select {
			case <-done:
				return
			case t := <-ticker.C:
				fmt.Println("Tick at:", t)
			}
		}
	}()

	time.Sleep(1600 * time.Millisecond)
	ticker.Stop()
	done <- true
	fmt.Println("ticker stopped")
}
