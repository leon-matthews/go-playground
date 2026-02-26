// N workers.
package main

import (
	"fmt"
	"math/rand/v2"
	"strings"
	"time"
)

// solution start

// makeSema uses a semaphore to run the given function using only n goroutines concurrently
// Returns the handle and wait functions, which should be used as follows:
//
//	const numGoroutines = 4
//	handle, wait := makeSema(numGoroutines, some_function)
//	for _, phrase := range phrases {
//	  handle(phrase)
//	}
//	wait()
//
// handle() will run the given function in a new goroutine - or block if there are
// already n goroutines running. wait() blocks until all workers are finished.
func makeSema(n int, handler func(int, string)) (handle func(string), wait func()) {
	// Create and populate semaphore with integers
	semaphore := make(chan int, n)
	for i := range n {
		semaphore <- i
	}

	// handle() takes a token from the channel
	// and processes the given phrase with handler().
	handle = func(s string) {
		id := <-semaphore // Take semaphore or block
		go func() {
			handler(id, s)
			semaphore <- id
		}()
	}

	// wait() waits until all tokens return to the channel.
	wait = func() {
		for range n {
			<-semaphore
		}
	}

	return handle, wait
}

// solution end

// say prints each word of a phrase.
func say(id int, phrase string) {
	for _, word := range strings.Fields(phrase) {
		fmt.Printf("Worker #%d says: %s...\n", id, word)
		time.Sleep(rand.N(100 * time.Millisecond))
	}
}

func main() {
	phrases := []string{
		"go is awesome",
		"cats are cute",
		"rain is wet",
		"channels are hard",
		"floor is lava",
	}

	handle, wait := makeSema(4, say)
	for _, phrase := range phrases {
		handle(phrase)
	}
	wait()
}
