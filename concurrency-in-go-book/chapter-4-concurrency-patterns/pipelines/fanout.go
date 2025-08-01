package main

import (
	"fmt"
	"math/rand"
	"runtime"
	"time"
)

func main() {
	numFinders := runtime.NumCPU()
	fmt.Printf("Using %d prime finders\n", numFinders)

	start := time.Now()
	done := make(chan any)
	out := randomStream(done, 50_000_000)
	out = primeFilter(done, out)
	for range 10 {
		fmt.Println(<-out)
	}
	close(done)
	fmt.Printf("Search took %v", time.Since(start))
}

// primeFilter removes non-primes from the integer stream
func primeFilter(done <-chan any, ints <-chan int) <-chan int {
	primes := make(chan int)

	go func() {
		defer close(primes)
		for i := range ints {
			// Check
			prime := true
			for divisor := i - 1; divisor > 1; divisor-- {
				if i%divisor == 0 {
					prime = false
					break
				}
			}

			// Send prime - or abort if cancelled
			if prime {
				select {
				case <-done:
					return
				case primes <- i:
				}
			}
		}
	}()

	return primes
}

// randomStream sends random integers into the ouput stream... forever!
func randomStream(done <-chan any, limit int) <-chan int {
	ints := make(chan int)

	go func() {
		defer close(ints)
		for {
			select {
			case ints <- rand.Intn(limit):
			case <-done:
				return
			}
		}
	}()

	return ints
}
