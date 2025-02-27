package main

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

func main() {
	phrases := []string{
		"go is awesome",
		"cats are cute",
		"rain is wet",
		"channels are hard",
		"floor is lava",
	}
	finished := make(chan bool)
	for idx, phrase := range phrases {
		go say(idx+1, phrase, finished)
	}

	// Wait until all goroutines are finished
	for range len(phrases) {
		<-finished
	}
}

// say does its work then writes a value to finished to indicate that it has... finished.
func say(id int, phrase string, finished chan<- bool) {
	for _, word := range strings.Fields(phrase) {
		fmt.Printf("Worker #%d says: %s...\n", id, word)
		dur := time.Duration(rand.Intn(100)) * time.Millisecond
		time.Sleep(dur)
	}
	finished <- true
}

// hardToReadSay is the exact same as say(), but using the zero-length struct{} type
// Note the dreadful syntax to 'create' an instance when writing to channel
func hardToReadSay(id int, phrase string, finished chan struct{}) {
	for _, word := range strings.Fields(phrase) {
		fmt.Printf("Worker #%d says: %s...\n", id, word)
		dur := time.Duration(rand.Intn(100)) * time.Millisecond
		time.Sleep(dur)
	}
	finished <- struct{}{}
}
