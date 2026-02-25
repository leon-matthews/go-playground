// Four counters starts four countWords goroutines then uses a done channel to wait for them to finish
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
// When there are no more words, next() returns an empty string.
func countDigitsInWords(next func() string) counter {
	pending := submitWords(next)
	done := make(chan struct{})
	counted := make(chan pair)

	// Start four countWords goroutines instead of one.
	numCounters := 4
	for range numCounters {
		go countWords(done, pending, counted)
	}

	// Close counted channel once all goroutines are finished
	go func() {
		for range numCounters {
			<-done
		}
		close(counted)
	}()

	return fillStats(counted)
}

// submitWords sends words to be counted.
func submitWords(next func() string) <-chan string {
	out := make(chan string)
	go func() {
		for n := next(); n != ""; n = next() {
			out <- n
		}
		close(out)
	}()
	return out
}

// countWords counts digits in words.
func countWords(done chan<- struct{}, in <-chan string, out chan<- pair) {
	for word := range in {
		count := countDigits(word)
		p := pair{word, count}
		out <- p
	}
	done <- struct{}{}
}

// fillStats prepares the final statistics.
func fillStats(in <-chan pair) counter {
	stats := counter{}
	for p := range in {
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
	phrase := "1 22 333 4444 55555 666666 7777777 88888888"
	next := wordGenerator(phrase)
	stats := countDigitsInWords(next)
	printStats(stats)
}
