package main

import (
	"strings"
)

var Base36Alphabet = []byte("0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ")

// Base36Decode decodes a string to a byte slice
func Base36Decode(encoded string) []byte {
	return []byte{}
}

// Base36Encode encodes a slice of bytes to an ascii string
func Base36Encode(bytes []byte) string {
	if len(bytes) == 0 {
		return ""
	}

	// Each input byte results in one or more characters added to string
	// (Actually `math.Log2(len(alphabet))`)
	var encoded strings.Builder
	push := func(value int) int {
		index := value % len(Base36Alphabet)
		encoded.WriteByte(Base36Alphabet[index])
		return index
	}

	var value int = 1
	for _, b := range bytes {
		value *= int(b)
		for {
			value /= push(value)
			if value < len(Base36Alphabet) {
				break
			}
		}
		push(value)
	}
	return encoded.String()
}
