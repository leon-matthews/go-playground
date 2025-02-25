package main

import (
	"fmt"
	"strings"
)

func main() {
	str := "one,two,,four"

	in := make(chan string)
	go func() {
		words := strings.Split(str, ",")
		for _, word := range words {
			in <- word
		}
		close(in)
	}()

	// For-range statement automatically breaks if channel is closed
	for word := range in {
		if word != "" {
			fmt.Printf("%s ", word)
		}
	}
}
