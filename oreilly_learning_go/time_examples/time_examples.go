package main

import (
	"fmt"
	"time"
)

func main() {
	limit := 1_000_000
	var n uint64
	var sum uint64

	start := time.Now()
	for range 1_000_000 {
		n++
		sum += n
	}
	elapsed := time.Since(start) // same as `time.Now().Sub(start)`
	fmt.Printf("Calculated sum of 1 to %d = %d in %v", limit, sum, elapsed)
}
