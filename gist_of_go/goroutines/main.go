// Code exercises from:
// https://antonz.org/go-concurrency/goroutines/
package main

import (
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"
)

var pangrams = []string{
	"The quick brown fox jumps over the lazy dog",
	"Pack my box with five dozen liquor jugs",
	"The five boxing wizards jump quickly",
	"Sphinx of black quartz judge my vow",
}

func main() {
	var wg sync.WaitGroup

	for id, pangram := range pangrams {
		wg.Add(1)
		go func() {
			defer wg.Done()
			say(id+1, pangram)
		}()
	}

	// Wait for goroutines to finish
	wg.Wait()
}

func say(id int, phrase string) {
	for _, word := range strings.Fields(phrase) {
		fmt.Printf("#%d %s\n", id, word)
		ms := time.Duration(rand.Intn(100)) * time.Millisecond
		time.Sleep(ms)
	}
}
