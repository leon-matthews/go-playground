// Let's say we want only N say goroutines to exist at the same time.
// A buffered channel can help achieve this. Here's the idea:
//
//   - Create a channel with a buffer size of N and fill it with "tokens" (arbitrary values).
//   - Before starting, a goroutine takes a token from the channel.
//   - Once finished, the goroutine returns the token to the channel.
//
// If there are no tokens left in the channel, the next goroutine will not
// start and will wait until someone returns a token to the channel. In this
// way, no more than N goroutines will run simultaneously. This setup is
// called a semaphore.
package main

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

func main() {
	phrases := []string{
		"a b c", "d e", "f", "g h", "i j k", "l m", "n",
	}

	// Semaphore for 2 goroutines.
	sema := make(chan int, 2)
	sema <- 1
	sema <- 2

	for index, phrase := range phrases {
		// Get a token from the channel (if there are any).
		tok := <-sema
		go say(sema, index, tok, phrase)
	}

	// Wait for all goroutines to finish their work
	// (all tokens are returned to the channel).
	<-sema
	<-sema
	fmt.Println("done")
}

// say prints each word of a phrase.
func say(sema chan<- int, worker, tok int, phrase string) {
	for _, word := range strings.Fields(phrase) {
		fmt.Printf("Worker %d with token #%d says: %s...\n", worker, tok, word)
		dur := time.Duration(rand.Intn(100)) * time.Millisecond
		time.Sleep(dur)
	}
	// Return the token to the channel.
	sema <- tok
}
