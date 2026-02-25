// Promise.all()
package main

import (
	"fmt"
	"sync"
	"time"
)

// solution start

// gather runs all passed functions concurrently, preserving order.
// Allows different goroutines to write to same slice (!) but at different indices
func gather(funcs []func() any) []any {
	results := make([]any, len(funcs))
	var wg sync.WaitGroup
	for i, f := range funcs {
		wg.Go(func() {
			results[i] = f()
		})
	}
	wg.Wait()
	return results
}

type result struct {
	index int
	value any
}

// gather runs all passed functions concurrently, preserving order.
// Write results to buffered channel (in arbitrary order) then use index
// to build slice out results.
func gather2(funcs []func() any) []any {
	// Run one goroutine per given function
	// Feed results into channel in any order
	out := make(chan result, len(funcs))
	for i, f := range funcs {
		go func() {
			out <- result{i, f()}
		}()
	}
	close(out)

	// Collect results into correct order
	results := make([]any, 0, len(funcs))
	for r := range out {
		results[r.index] = r.value
	}
	return results
}

// solution end

// squared returns a function that returns
// the square of the input number.
func squared(n int) func() any {
	return func() any {
		time.Sleep(100 * time.Millisecond)
		return n * n
	}
}

func main() {
	start := time.Now()
	funcs := []func() any{squared(2), squared(3), squared(4)}
	nums := gather(funcs)
	nums2 := gather(funcs)
	elapsed := time.Since(start)
	fmt.Println(nums)
	fmt.Println(nums2)
	fmt.Printf("Took %v\n", elapsed)
}
