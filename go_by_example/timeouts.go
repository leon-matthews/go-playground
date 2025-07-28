package main

import (
	"fmt"
	"time"
)

func main() {
	selectTimeout()
	selectTimeoutTriggered()
}

// Implement timeout in select, not triggered in this case
func selectTimeout() {
	ch1 := make(chan string, 1)
	go func() {
		time.Sleep(200 * time.Millisecond)
		ch1 <- "result 1"
	}()

	// Result arrives first; timeout not triggered
	select {
	case res := <-ch1:
		fmt.Println(res)
	case <-time.After(300 * time.Millisecond):
		fmt.Println("timeout1")
	}
}

// Timeout triggered
func selectTimeoutTriggered() {
	ch1 := make(chan string, 1)
	go func() {
		time.Sleep(300 * time.Millisecond)
		ch1 <- "result 2"
	}()

	// Timeout triggered
	select {
	case res := <-ch1:
		fmt.Println(res)
	case <-time.After(200 * time.Millisecond):
		fmt.Println("timeout2")
	}
}
