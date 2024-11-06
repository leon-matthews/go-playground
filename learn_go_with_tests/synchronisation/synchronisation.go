package main

import (
	"fmt"
	"sync"
)

func main() {
	fmt.Println("Synchronisation Example")
}

func NewCounter() *Counter {
	return &Counter{}
}

type Counter struct {
	value int
	lock  sync.Mutex
}

func (c *Counter) Increment() {
	c.lock.Lock()
	c.value++
	c.lock.Unlock()
}

func (c *Counter) Value() int {
	return c.value
}
