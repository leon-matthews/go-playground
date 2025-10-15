package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
)

func main() {
	// Add -l flag to count lines instead of words
	lines := flag.Bool("l", false, "print the newline counts")
	flag.Parse()
	fmt.Println(count(os.Stdin, *lines))
}

// count returns number of words in the given reader
func count(r io.Reader, lines bool) int {
	// Create scanner
	scanner := bufio.NewScanner(r)

	// Use default lines split or change to words?
	if !lines {
		scanner.Split(bufio.ScanWords)
	}

	// Count!
	wc := 0
	for scanner.Scan() {
		wc++
	}
	return wc
}
