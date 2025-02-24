// Code exercises from:
// https://antonz.org/go-concurrency/goroutines/
package main

import (
	"fmt"
	"strings"
	"sync"
	"unicode"
)

func main() {
	phrase := "0ne 1wo thr33 4068"
	counts := countDigitsInWords(phrase)
	printStats(counts)
}

// counter stores the number of digits in each word.
type counter map[string]int

// asStats converts statistics from sync.Map to a regular map.
func asStats(m *sync.Map) counter {
	stats := counter{}
	m.Range(func(word, count any) bool {
		stats[word.(string)] = count.(int)
		return true
	})
	return stats
}

// printStats prints the number of digits in words.
func printStats(stats counter) {
	for word, count := range stats {
		fmt.Printf("%s: %d\n", word, count)
	}
}

// countDigitsInWords counts the number of digits in the words of a phrase.
func countDigitsInWords(phrase string) counter {
	var wg sync.WaitGroup
	syncStats := new(sync.Map)
	words := strings.Fields(phrase)

	// Count the number of digits in words,
	// using a separate goroutine for each word.
	for _, word := range words {
		wg.Add(1)
		go func() {
			defer wg.Done()
			syncStats.Store(word, countDigits(word))
		}()
	}

	wg.Wait()
	return asStats(syncStats)
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
