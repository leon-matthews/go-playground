package main

import (
	"fmt"
	"math/rand/v2"
	"sync"
	"time"
)

func main() {
	wg := sync.WaitGroup{}
	c := counter{}
	c2 := counter{}

	for _ = range 100 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c.increment()
			c2.increment2()
		}()
	}

	for _ = range 100 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c.decrement()
			c2.decrement2()
		}()
	}
	wg.Wait()
	fmt.Println("Synchronised value is", c.count)
	fmt.Println("Unsynchronised value is", c2.count)
}

type counter struct {
	count int
	lock  sync.Mutex
}

// decrement count with terrible race-condition
func (c *counter) decrement() {
	c.lock.Lock()
	defer c.lock.Unlock()
	current := c.count
	time.Sleep(rand.N(10 * time.Millisecond))
	c.count = current - 1
}

// increment count with terrible race-condition
func (c *counter) increment() {
	c.lock.Lock()
	defer c.lock.Unlock()
	current := c.count
	time.Sleep(rand.N(10 * time.Millisecond))
	c.count = current + 1
}

// decrement2 removes use of Mutex
func (c *counter) decrement2() {
	current := c.count
	time.Sleep(rand.N(10 * time.Millisecond))
	c.count = current - 1
}

// increment2 removes use of Mutex
func (c *counter) increment2() {
	current := c.count
	time.Sleep(rand.N(10 * time.Millisecond))
	c.count = current + 1
}
