// Mutexes are like atomic counters, but manage more complex state
package main

import (
	"fmt"
	"sync"
)

// Mutexes should not be copied, so be sure to pass by pointer
type Counter struct {
	mu       sync.Mutex
	counts map[string]int
}

func NewCounter(initial map[string]int) *Counter {
	return &Counter{counts: initial}
}

// Increment safely adds one to the count for the given name
func (c *Counter) Increment(name string) {
	// Block until lock available
	c.mu.Lock()

	// Unlock as soon as function exists
	defer c.mu.Unlock()

	// We can be sure that we're the only goroutine with
	c.counts[name]++
}

func main() {
	c := NewCounter(map[string]int{"a": 2, "b": 3})

	// A single increment
	c.Increment("b")
	fmt.Println(c.counts)

	var wg sync.WaitGroup
	doIncrement := func(name string, numTimes int) {
		for range numTimes {
			c.Increment(name)
		}
		wg.Done()
	}

	wg.Add(3)
	go doIncrement("a", 10_000)
	go doIncrement("b", 20_000)
	go doIncrement("c", 10_000)
	wg.Wait()

	fmt.Println(c.counts)
}
