// Dup prints the count and text of lines that appear more than once in the named
// input files. It reads in entire file at once ("slurp" mode). Adapted from
// https://github.com/adonovan/gopl.io/tree/master/ch1/dup3.
package main

import (
	"fmt"
	"os"
	"strings"
)

func main() {
	// Check args
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go [FILES...]")
	}

	// Read lines into map of counts
	counts := make(map[string]int)
	for _, arg := range os.Args[1:] {
		b, err := os.ReadFile(arg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "dup: %v\n", err)
			continue
		}
		for line := range strings.Lines(string(b)) {
			counts[line]++
		}
	}

	// Print any duplicate lines
	for line, n := range counts {
		if n > 1 {
			fmt.Printf("%d\t%s\n", n, line)
		}
	}
}
