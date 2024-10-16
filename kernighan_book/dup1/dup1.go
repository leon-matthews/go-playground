// Print any lines from stdin that appear more than once
// Counts are printed, but lines do not appear in any particular order
package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	counts := make(map[string]int)
	input := bufio.NewScanner(os.Stdin)
	for input.Scan() {
		counts[input.Text()]++
	}

	// Ignore errors from input.Err()
	for line, count := range counts {
		if count > 1 {
			fmt.Printf("%d\t%v\n", count, line)
		}
	}
}
