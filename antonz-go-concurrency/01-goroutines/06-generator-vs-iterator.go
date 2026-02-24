// Try converting generator function to new iterator type
package main

import (
	"fmt"
	"iter"
	"strings"
)

func main() {
	phrase := "0ne 1wo thr33 4068"

	// Old school generator
	next := wordGenerator(phrase)
	for {
		word := next()
		if word == "" {
			break
		}
		fmt.Println(word)
	}

	// Friendly iterator
	for word := range wordIterator(phrase) {
		fmt.Println(word)
	}

	// Stdlib iterator
	for word := range strings.FieldsSeq(phrase) {
		fmt.Println(word)
	}
}

// wordGenerator returns a generator that yields words from a phrase.
func wordGenerator(phrase string) func() string {
	words := strings.Fields(phrase)
	idx := 0
	return func() string {
		if idx == len(words) {
			return ""
		}
		word := words[idx]
		idx++
		return word
	}
}

// wordIterator reworks wordGenerator into a standard Go iterator
// This whole function is redundant, however. We could just return
// `words` as-is, just as soon as it was created!
func wordIterator(phrase string) iter.Seq[string] {
	words := strings.FieldsSeq(phrase)
	return func(yield func(string) bool) {
		for word := range words {
			if !yield(word) {
				return
			}
		}
	}
}
