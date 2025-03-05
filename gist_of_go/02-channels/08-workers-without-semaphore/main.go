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
	pending := make(chan string)

	// Write work to be done to pending
	go func() {
		for _, phrase := range phrases {
			pending <- phrase
		}
		close(pending)
	}()

	done := make(chan struct{})

	// Create two workers
	go say(done, pending, 1)
	go say(done, pending, 2)

	// Read twice, once for each worker
	<-done
	<-done
}

func say(done chan<- struct{}, pending <-chan string, id int) {
	// Pull work out of pending
	for phrase := range pending {
		for _, word := range strings.Fields(phrase) {
			fmt.Printf("Worker #%d says: %s...\n", id, word)
			dur := time.Duration(rand.Intn(100)) * time.Millisecond
			time.Sleep(dur)
		}
	}

	// Once no work remains, wait for a reader
	done <- struct{}{}
}
