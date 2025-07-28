package main

import (
	"fmt"
	"math/rand/v2"
	"sync"
	"time"
)

// Note that this approach has no straightforward way to propagate errors from
// workers. For more advanced use cases, consider using the [errgroup] package.
// [errgroup]: https://pkg.go.dev/golang.org/x/sync/errgroup
func main() {
	var wg sync.WaitGroup

	for i := range 5 {
		// Increment count of running goroutines
		wg.Add(1)

		go func() {
			// Decrement count of running goroutines
			defer wg.Done()
			worker(i)
		}()
	}

	// Block until count of running goroutines returns to zero
	wg.Wait()
}

func worker(id int) {
	fmt.Printf("Worker %d starting...\n", id)
	time.Sleep(rand.N(1000 * time.Millisecond))
	fmt.Printf("...worker %d done\n", id)
}
