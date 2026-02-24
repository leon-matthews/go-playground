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

	for {
		word, ok := <-in
		if !ok {
			break
		}

		if word != "" {
			fmt.Printf("%s ", word)
		}
	}
}
