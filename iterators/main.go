package main

import (
	"fmt"
	"iter"
)

func main() {
	for prefix := range HexStrings(5) {
		fmt.Println(prefix)
	}
}

// HexStrings generates all hexadecimal strings of the given length, zero-padded.
func HexStrings(length int) iter.Seq[string] {
	limit := 0x01 << (length * 4)
	return func(yield func(string) bool) {
		for v := range limit {
			hex := fmt.Sprintf("%0*x", length, v)
			if !yield(hex) {
				return
			}
		}
	}
}
