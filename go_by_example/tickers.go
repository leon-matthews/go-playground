package main

import (
	"fmt"
	"time"
)

func main() {
	ticker := time.NewTicker(100 * time.Millisecond)
	done := make(chan bool)

	go func() {
		defer func() { fmt.Println("Goroutine finished") }()
		for {
			select {
			case <-done:
				return
			case t := <-ticker.C:
				fmt.Println(t)
			}
		}
	}()

	time.Sleep(1 * time.Second)
	ticker.Stop()
	fmt.Println("Ticker stopped")
	done <- true
}
