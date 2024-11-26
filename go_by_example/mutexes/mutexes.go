// Mutexes are like atomic counters, but manage more complex state
package main

import (
	"github.com/k0kubun/pp/v3"
	"sync"
)

// Mutexes should not be copied, so be sure to pass by pointer
type Container struct {
	mu       sync.Mutex
	counters map[string]int
}

// Inc increments count for given name
func (c *Container) Inc(name string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.counters[name]++
}

func main() {
	c := Container{
		counters: map[string]int{"a": 2, "b": 3},
	}
	c.Inc("b")

	var wg sync.WaitGroup
	doIncrement := func(name string, n int) {
		for i := 0; i < n; i++ {
			c.Inc(name)
		}
		wg.Done()
	}

	wg.Add(3)
	go doIncrement("a", 10_000)
	go doIncrement("b", 20_000)
	go doIncrement("a", 10_000)
	wg.Wait()

	pp.Println(c)
}
