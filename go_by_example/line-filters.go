package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	// Scanner steps through 'tokens' in input
	// The defination of a token is configurable and is a line of text by default.
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		upper := strings.ToUpper(scanner.Text())
		fmt.Println(upper)
	}

	// Check for errors after Scan() returns false
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading standard input:", err)
	}
}
