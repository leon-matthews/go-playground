package main

import (
	"fmt"
    "math"
	"runtime"
	"sync"
)

var kB = math.Pow(2, 10)
var MB = math.Pow(2, 20)

func main() {
	var c <-chan any
	var wg sync.WaitGroup
	noop := func() { wg.Done(); <-c }
	const numGoroutines = 1e6
	wg.Add(numGoroutines)

	before := memConsumed()
	for i := numGoroutines; i > 0; i-- {
		go noop()
	}
	wg.Wait()
	after := memConsumed()

	fmt.Printf("Before: %.3fMib\n", float64(before)/MB)
    fmt.Printf("After: %.3fMib\n", float64(after)/MB)
	fmt.Printf("Difference: %.3fkiB", float64(after-before)/numGoroutines/kB)
}

func memConsumed() uint64 {
	runtime.GC()
	var s runtime.MemStats
	runtime.ReadMemStats(&s)
	return s.Sys
}
