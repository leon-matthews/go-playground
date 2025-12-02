package main

import (
	"fmt"
	"math/rand/v2"
	"sync"
)

func main() {
	var numCreated int

	// The [New] member must return [any]
	pool := &sync.Pool{
		New: func() any {
			numCreated++
			mem := make([]byte, 1024)
			return &mem
		},
	}

	// Seed pool with four buffers
	pool.Put(pool.New())
	pool.Put(pool.New())
	pool.Put(pool.New())
	pool.Put(pool.New())

	const numWorkers = 1024 * 1024
	var wg sync.WaitGroup
	wg.Add(numWorkers)
	for range numWorkers {
		go func() {
			defer wg.Done()

			// Get object - maybe from pool, maybe new
			// We can may no assumptions about state of object
			mem := pool.Get().(*[]byte)

			// Do something with object
			(*mem)[rand.N(1024)] = rand.N[byte](255)

			// Put object back, otherwise there is no benefit!
			defer pool.Put(mem)
		}()

	}
	wg.Wait()

	// Fetch and print a buffer to see its state. Probably messy by now!
	mem := pool.Get().(*[]byte)
	fmt.Println(*mem)

	fmt.Println("Created", numCreated)
}
