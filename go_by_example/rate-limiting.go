// Go elegantly supports rate limiting with goroutines, channels, and tickers.
package main

import (
	"fmt"
	"time"
)

func main() {
	fmt.Println("Blocking")
	blocking()
	fmt.Println()

	fmt.Println("Bursty (note timestamps)")
	bursty()
}

// Limits rate by blocking on reads from ticker
func blocking() {
	// Simulate queue of incoming requests
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
}

func bursty() {
	// New limiter allows bursts of up to three events
	// Every requests takes a value out of the limiter channel
	burstyLimiter := make(chan time.Time, 3)

	// Fill up channel to simulate quiescent server
	for range 3 {
		burstyLimiter <- time.Now()
	}

	// Add a new value to limiter every 200 milliseconds
	go func() {
		for t := range time.Tick(200 * time.Millisecond) {
			burstyLimiter <- t
		}
	}()

	// Simulate 5 more incoming requests
	burstyRequests := make(chan int, 6)
	for i := range 6 {
		burstyRequests <- i
	}
	close(burstyRequests)

	for req := range burstyRequests {
		<-burstyLimiter
		fmt.Println("request", req, time.Now())
	}
}
