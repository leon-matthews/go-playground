// In practice, you'll often find an "inverse" approach to semaphores:
//
//	Create an empty channel with a buffer size of N.
//	Before starting, a goroutine puts a token into the channel.
//	Once finished, the goroutine takes a token from the channel.
//
// If the channel is full of tokens, the next goroutine will not start and
// will wait until someone takes a token from the channel. In this way, no
// more than N goroutines will run simultaneously.
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
	sema := make(chan struct{}, 2)

	for i, phrase := range phrases {
		// Put a token into the channel (if there is space).
		sema <- struct{}{}
		go say(sema, i, phrase)
	}

	// Wait for all goroutines to finish their work
	// (all tokens taken from the channel).
	sema <- struct{}{}
	sema <- struct{}{}
	fmt.Println("done")
}

// say prints each word of a phrase.
func say(sema <-chan struct{}, worker int, phrase string) {
	for _, word := range strings.Fields(phrase) {
		fmt.Printf("Worker #%d says: %s...\n", worker, word)
		dur := time.Duration(rand.Intn(100)) * time.Millisecond
		time.Sleep(dur)
	}
	// Take the token from the channel.
	<-sema
}
