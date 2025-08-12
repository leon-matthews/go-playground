package main

import (
	"fmt"
	"strings"
	"sync"
	"unicode"
)

// counter stores the number of digits in each word.
type counter map[string]int

func main() {
	phrase := "0ne 1wo thr33 4068"
	counts := countDigitsInWords(phrase)
	printCounter(counts)
}

// countDigits returns the number of numeric characters in string
func countDigits(str string) int {
	count := 0
	for _, c := range str {
		if unicode.IsDigit(c) {
			count++
		}
	}
	return count
}

// countDigitsInWords counts numbers in each word in phrase - in parallel
func countDigitsInWords(phrase string) counter {
	var wg sync.WaitGroup
	data := new(sync.Map)
	words := strings.Fields(phrase)

	wg.Add(len(words))
	for _, word := range words {
		go func(word string) {
			defer wg.Done()
			data.Store(word, countDigits(word))
		}(word)
	}
	wg.Wait()

	return asCounter(data)
}

// asCounter converts statistics from sync.Map to a regular map.
func asCounter(m *sync.Map) counter {
	counts := counter{}
	m.Range(func(word, count any) bool {
		counts[word.(string)] = count.(int)
		return true
	})
	return counts
}

// printCounter prints the number of digits in words.
func printCounter(counts counter) {
	for word, count := range counts {
		fmt.Printf("%s: %d\n", word, count)
	}
}
