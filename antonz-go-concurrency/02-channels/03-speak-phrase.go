package main

import (
	"fmt"
	"math/rand/v2"
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

	done := make(chan struct{})
	for idx, phrase := range phrases {
		go say(done, idx+1, phrase)
	}

	// Wait until all goroutines have writen to done
	for range len(phrases) {
		<-done
	}
}

// say function writes to done when it's finished
func say(done chan<- struct{}, id int, phrase string) {
	for _, word := range strings.Fields(phrase) {
		fmt.Printf("Worker #%d says: %s...\n", id, word)
		time.Sleep(rand.N(100 * time.Millisecond))
	}
	done <- struct{}{}
}
