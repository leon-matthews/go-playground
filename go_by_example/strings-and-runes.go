package main

import (
	"fmt"
	"unicode/utf8"
)

func main() {
	// 'Hello' in Thai
	const s = "สวัสดี"

	// Strings are equivalent to []byte, so len() gives encoded length
	fmt.Println(s, "len:", len(s))

	// Index to get byte values
	for i := 0; i < len(s); i++ {
		fmt.Printf("%x ", s[i])
	}
	fmt.Println()

	// Count runes using: utf8.RuneCountInString()
	fmt.Println("Rune count:", utf8.RuneCountInString(s))

	// Range iterates over runes, note effect combining characters
	for idx, runeValue := range s {
        fmt.Printf("%#U starts at %d\n", runeValue, idx)
    }
    fmt.Println()

    // utf8.DecodeRuneInString() returns the frist rune in the string and its width in bytes
    r, size := utf8.DecodeRuneInString(s)
    fmt.Printf("utf8.DecodeRuneInString(%q): %v (size %d)\n", s, r, size)
    for i, w := 0, 0; i < len(s); i += w {
		runeValue, width := utf8.DecodeRuneInString(s[i:])
		fmt.Printf("%#U starts at %d\n", runeValue, i)
		examineRune(runeValue)
        w = width
	}
}

// Rune literals (single quotes) can be compared directly
func examineRune(r rune) {
	if r == 't' {
		fmt.Println("found tee")
	} else if r == 'ส' {
		fmt.Println("found so sua")
	}
}
