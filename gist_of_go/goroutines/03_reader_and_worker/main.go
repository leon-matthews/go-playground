// For more complex tasks, it's useful to have a goroutine for reading data
// (reader) and another for processing data (worker):
//
// ┌───────────────┐               ┌───────────────┐
// │ sends words   │               │ counts digits │               ┌────────────────┐
// │ to be counted │ → (pending) → │ in words      │ → (counted) → │ fills stats    │
// │               │               │               │               └────────────────┘
// └───────────────┘               └───────────────┘
package main

import (
	"fmt"
	"strings"
	"unicode"
)

// counter stores the number of digits in each word.
// The key is the word, and the value is the number of digits.
type counter map[string]int

// pair stores a word and the number of digits in it.
type pair struct {
	word  string
	count int
}

// countDigitsInWords counts the number of digits in words,
// fetching the next word with next().
func countDigitsInWords(next func() string) counter {
	pending := submitWords(next)
	counted := countWords(pending)
	return fillStats(counted)
}

// submitWords sends words to be counted
func submitWords(next func() string) <-chan string {
	out := make(chan string)
	go func() {
		for {
			word := next()
			out <- word
			if word == "" {
				break
			}
		}
	}()
	return out
}

// countWords counts digits in words
func countWords(pending <-chan string) <-chan pair {
	out := make(chan pair)
	go func() {
		for word := range pending {
			count := countDigits(word)
			out <- pair{word, count}
			if word == "" {
				break
			}
		}
	}()
	return out
}

// fillStats prepares the final statistics.
func fillStats(counted <-chan pair) counter {
	stats := counter{}
	for p := range counted {
		if p.word == "" {
			break
		}
		stats[p.word] = p.count
	}
	return stats
}

// countDigits returns the number of digits in a string.
func countDigits(str string) int {
	count := 0
	for _, char := range str {
		if unicode.IsDigit(char) {
			count++
		}
	}
	return count
}

// printStats prints the number of digits in words.
func printStats(stats counter) {
	for word, count := range stats {
		fmt.Printf("%s: %d\n", word, count)
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

func main() {
	phrase := "0ne 1wo thr33 4068"
	next := wordGenerator(phrase)
	stats := countDigitsInWords(next)
	printStats(stats)
}
