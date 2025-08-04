package main

import (
	"flag"
	"fmt"
)

func main() {
	// Pointers to internal values
	word := flag.String("word", "foo", "a string")
	number := flag.Int("number", 42, "a number")
	boolean := flag.Bool("boolean", true, "a boolean")

	// Pointer to your own value
	var word2 string
	flag.StringVar(&word2, "w", "", "second string")

	// Call AFTER all flags defined, but BEFORE flags used
	flag.Parse()

	fmt.Println("word:", *word)
	fmt.Println("word2:", word2)
	fmt.Println("number:", *number)
	fmt.Println("boolean:", *boolean)
}
