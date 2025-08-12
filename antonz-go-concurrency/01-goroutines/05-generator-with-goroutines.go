// Generator with goroutines.
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
// fetching the next word with next().
func countDigitsInWords(next func() string) counter {
	counted := make(chan pair)

	go func() {
		for {
			word := next()
			count := countDigits(word)
			counted <- pair{word, count}
			if word == "" {
				break
			}
		}
	}()

	// Read values from the counted channel and fill stats.
	stats := counter{}
	for {
		p := <-counted
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
