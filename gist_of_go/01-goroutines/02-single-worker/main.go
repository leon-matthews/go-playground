// Generator with goroutines.
// The goroutine went through the words and counted the digits, while the
// outer function received the counts and updated the counter:
// ┌───────────────┐
// │ loops through │               ┌────────────────┐
// │ words and     │ → (counted) → │ fills stats    │
// │ counts digits │               └────────────────┘
// └───────────────┘
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
func countDigitsInWords(next func() string) counter {
	counted := make(chan pair)

	go func() {
		// Fetch words from the generator,
		// count the number of digits in each,
		// and write it to the counted channel.
		for {
			word := next()
			counted <- pair{word, countDigits(word)}
			if word == "" {
				break
			}
		}
	}()

	// Read values from the counted channel
	// and fill stats.
	stats := counter{}
	for {
		p := <-counted
		if p.word == "" {
			break
		}
		stats[p.word] = p.count
	}

	// As a result, stats should contain words
	// and the number of digits in each.

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
