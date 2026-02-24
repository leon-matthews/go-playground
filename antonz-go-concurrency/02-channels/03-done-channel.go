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

	for idx, phrase := range phrases {
		go say(idx+1, phrase)
	}
}

// say function writes to done when it's finished
func say(id int, phrase string) {
	for _, word := range strings.Fields(phrase) {
		fmt.Printf("Worker #%d says: %s...\n", id, word)
		time.Sleep(rand.N(100 * time.Millisecond))
	}
}
