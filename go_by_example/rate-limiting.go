// Go elegantly supports rate limiting with goroutines, channels, and tickers.
package main

import (
	"fmt"
	"time"
)

func main() {
	// Create buffered requests channel and fill it
	requests := make(chan int, 5)
	for i := 1; i <= 5; i++ {
		requests <- i
	}
	close(requests)

	// Ticker limiter sends time on its channel
	// As of Go 1.23, the garbage collector can recover unreferenced tickers,
	// even if they haven't been explicitly stopped
	limiter := time.Tick(200 * time.Millisecond)

	// Block before each receive to limit request to 1 per 200ms
	for req := range requests {
		now := <-limiter
		fmt.Println(req, now)
	}

	// New limiter allows bursts of up to three events
	burstyLimiter := make(chan time.Time, 3)

	go func() {
		for t := range time.Tick(200 * time.Millisecond) {
			burstyLimiter <- t
			fmt.Println("burstyLimiter", <-burstyLimiter)
		}
	}()

	burstyRequests := make(chan int, 5)
	for i := 1; i <= 5; i++ {
		burstyRequests <- i
	}
}
