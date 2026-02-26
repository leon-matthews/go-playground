package main

import (
	"fmt"
	"math/rand/v2"
	"strings"
	"time"
)

// A semaphore restricts the number of concurrent goroutines by requiring
// each goroutine to wait for a token to become available.
// The results is many short-lived goroutines, but never more than N at once.
func main() {
	phrases := []string{
		"a b c", "d e", "f", "g h", "i j k", "l m", "n", "o p q", "r s t",
	}

	// Create semaphore and fill with tokens
	const n = 2
	semaphore := make(chan int, n)
	for i := range n {
		semaphore <- i
	}

	// Say phrases, limited by token availability
	for _, phrase := range phrases {
		// Get a token (blocking if none available)
		token := <-semaphore
		go say(semaphore, token, phrase)
	}

	// Wait for workers to finish: once they've all returned their tokens
	for range n {
		<-semaphore
	}
}

// say puts its token back into the semaphore once its work is finished
// Note that the semaphone is write-only
func say(semaphore chan<- int, token int, phrase string) {
	for word := range strings.FieldsSeq(phrase) {
		fmt.Printf("Worker #%d says: %s...\n", token, word)
		time.Sleep(rand.N(time.Millisecond * 100))
	}
	semaphore <- token
}
