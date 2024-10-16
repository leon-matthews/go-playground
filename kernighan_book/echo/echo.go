// Print command-line argument separated by a single space.
package main

import (
	"fmt"
	"os"
)

func main() {
	// Elegant use of zero-value to handle first use of `sep`
	var s, sep string
	for _, arg := range os.Args[1:] {
		s += sep + arg
		sep = " "
	}
	fmt.Println(s)
}
