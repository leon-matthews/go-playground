package main

import (
	"fmt"
	"sync"
	"sync/atomic"
)

func main() {
	// Our counter
	var numOps atomic.Uint64

	// 50 goroutines that each increment the counter exactly 1000 times.
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)

		go func() {
			for c := 0; c < 1_000; c++ {
				numOps.Add(1)
			}
			wg.Done()
		}()
	}

	fmt.Println("Gotta wait. Only", numOps.Load(), "operations so far")
	wg.Wait()
	fmt.Println("Finished! Completed", numOps.Load(), "operations")
}
