package main

import (
	"fmt"
	"math/rand/v2"
	"strings"
	"sync"
	"time"
)

func main() {
	var wg sync.WaitGroup

	// WaitGroup has a handy-dandy Go() function
	wg.Go(func() {
		say(1, "The quick brown fox jumps over the lazy dog")
	})

	// Or you can use Add() and Done() manually
	wg.Add(1)
	go func() {
		defer wg.Done()
		say(2, "Sphinx of black quartz, judge my vow.")
	}()

	// Blocks until the Done() calls equal the Add() calls
	wg.Wait()

	fmt.Println("main exited")
}

func say(id int, phrase string) {
	for _, word := range strings.Fields(phrase) {
		fmt.Printf("%d says %q\n", id, word)
		time.Sleep(rand.N(100 * time.Millisecond))
	}
}
