package main

import (
	"fmt"
	"sync"
	"sync/atomic"
)

func main() {
	// Our counter
	var numOps atomic.Uint64

	// Creat 50 goroutines that each...
	var wg sync.WaitGroup
	for range 50 {
		wg.Add(1)

		go func() {
			// ...increment the counter 1,000 times
			for range 1_000 {
				numOps.Add(1)
			}
			wg.Done()
		}()
	}

	fmt.Println("Gotta wait. Only", numOps.Load(), "operations so far")
	wg.Wait()
	fmt.Println("Finished! Completed", numOps.Load(), "add operations")
}
