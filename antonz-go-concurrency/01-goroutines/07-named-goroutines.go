// Named Goroutines improves on Reader and Worker by creating named functions
// for each step.
package main

import (
	"fmt"
	"strings"
	"unicode"
)

func main() {
	phrase := "0ne 1wo thr33 4068"
	next := wordGenerator(phrase)
	stats := countDigitsInWords(next)
	printStats(stats)
}

// counter stores the number of digits in each word.
type counter map[string]int

// pair stores a word and the number of digits in it.
type pair struct {
	word  string
	count int
}

// countDigitsInWords counts the number of digits in words,
func countDigitsInWords(next func() string) counter {
	pending := make(chan string)
	go submitWords(next, pending)

	counted := make(chan pair)
	go countWords(pending, counted)

	return fillStats(counted)
}

// submitWords sends words to be counted.
func submitWords(next func() string, pending chan<- string) {
	for {
		word := next()
		if word == "" {
			close(pending)
			break
		}
		pending <- word
	}
}

// countWords counts digits in words.
func countWords(pending <-chan string, counted chan<- pair) {
	for word := range pending {
		count := countDigits(word)
		p := pair{word, count}
		counted <- p
	}
	close(counted)
}

// fillStats prepares the final statistics.
func fillStats(counted <-chan pair) counter {
	stats := counter{}
	for p := range counted {
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
