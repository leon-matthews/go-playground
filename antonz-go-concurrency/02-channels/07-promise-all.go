// Promise.all()
package main

import (
	"fmt"
	"time"
)

// solution start

// gather runs all passed functions concurrently
// and returns the results when they are ready.
func gather(funcs []func() any) []any {
	results := make([]any, len(funcs))
	for i, f := range funcs {
		results[i] = f()
	}
	return results
}

// solution end

// squared returns a function that returns
// the square of the input number.
func squared(n int) func() any {
	return func() any {
		return n * n
	}
}

func main() {
	funcs := []func() any{squared(2), squared(3), squared(4)}

	start := time.Now()
	nums := gather(funcs)
	elapsed := time.Since(start)

	fmt.Println(nums)
	fmt.Printf("Took %v\n", elapsed)
}
