package main

import (
	"fmt"
	"time"
)

func main() {
	timerExample()
	waitForWholeSecond()
	tickerExample()
}

// Set timer which expires when the current time has a whole number of seconds
func waitForWholeSecond() {
	now := time.Now()
	ns := int64(1e9 - now.Nanosecond())
	timer := time.NewTimer(time.Duration(ns))
	fired := <-timer.C
	fmt.Printf("Waited for %dns and fired at %v\n", ns, fired)
}

// When the Timer expires, the current time will be sent on channel 'C'
func timerExample() {
	// Can be used like [time.Sleep]
	timer1 := time.NewTimer(500 * time.Millisecond)
	<-timer1.C // blocks on the timer's channel C until it sends a value
	fmt.Println("timer1 fired")

	// Can also be cancelled before firing
	timer2 := time.NewTimer(1 * time.Second)

	go func() {
		<-timer2.C
		fmt.Println("timer2 fired")
	}()

	was_stopped := timer2.Stop()
	if was_stopped {
		fmt.Println("timer2 stopped")
	}

	fmt.Println("Finished")
}

// Will send the current time on the channel after each tick
func tickerExample() {
	ticker := time.NewTicker(100 * time.Millisecond)
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
