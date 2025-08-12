// Fixing directions.
package main

import (
	"fmt"
	"strings"
)

func main() {
	src := "go is awesome"
	res := encode(src)
	fmt.Println(res)
}

// encode encrypts sentences using the Caesar cipher.
func encode(str string) string {
	pending := submitter(str)
	encoded := encoder(pending)
	words := receiver(encoded)
	return strings.Join(words, " ")
}

// solution start

// submitter sends words to be encrypted.
func submitter(str string) <-chan string {
	ch := make(chan string)
	go func() {
		words := strings.Fields(str)
		for _, word := range words {
			ch <- word
		}
		close(ch)
	}()
	return ch
}

// encoder encrypts words.
func encoder(ch1 <-chan string) <-chan string {
	ch2 := make(chan string)
	go func() {
		for word := range ch1 {
			ch2 <- encodeWord(word)
		}
		close(ch2)
	}()
	return ch2
}

// receiver builds a slice of encrypted words.
func receiver(ch <-chan string) []string {
	words := []string{}
	for word := range ch {
		words = append(words, word)
	}
	return words
}

// solution end

// encodeWord encrypts a word using the Caesar cipher.
func encodeWord(word string) string {
	const shift = 13
	const char_a byte = 'a'
	encoded := make([]byte, len(word))
	for idx, char := range []byte(word) {
		delta := (char - char_a + shift) % 26
		encoded[idx] = char_a + delta
	}
	return string(encoded)
}
