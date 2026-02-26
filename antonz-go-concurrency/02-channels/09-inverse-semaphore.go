package main

import (
	"fmt"
	"math/rand/v2"
	"strings"
	"time"
)

// Inverse semaphore
// In practice, you'll often find an "inverse" approach to implementing semaphores:
//  1. Create an empty channel with a buffer size of N.
//  2. Before starting, a goroutine puts a token into the channel.
//  3. Once finished, the goroutine takes a token from the channel.
func main() {
	phrases := []string{
		"a b c", "d e", "f", "g h", "i j k", "l m", "n", "o p q", "r s t",
	}

	// Create semaphore (no need to 'prime' it)
	const n = 2
	semaphore := make(chan struct{}, n)

	// Say phrases, limited by space available in semaphore
	for _, phrase := range phrases {
		// Try and put a token into semaphore
		semaphore <- struct{}{}
		go say(semaphore, phrase)
	}

	// Wait for workers to finish: all n tokens have been taken
	for range n {
		semaphore <- struct{}{}
	}
}

// say removes a token from the semaphore once its finished
// Note that this time the semaphore is read-only
func say(semaphore <-chan struct{}, phrase string) {
	for _, word := range strings.Fields(phrase) {
		fmt.Printf("Worker says: %s...\n", word)
		time.Sleep(rand.N(100 * time.Millisecond))
	}

	// Take the token from the channel.
	<-semaphore
}
