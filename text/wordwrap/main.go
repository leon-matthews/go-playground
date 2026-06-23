// Command wordwrap demonstrates a low-allocation word-wrapping function.
package main

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

const nbsp = 0xA0

// WrapString wraps s so that no line exceeds lim runes, breaking only at whitespace.
//
// Existing newlines are preserved, as are runs of whitespace within a line.
// Non-breaking spaces (U+00A0) join words rather than allowing a break, and a
// word longer than lim overflows rather than being split.
//
// s is expected to be valid UTF-8; malformed bytes are passed through unchanged.
func WrapString(s string, lim uint) string {
	var b strings.Builder
	b.Grow(len(s)) // wrapping never lengthens the text, so this is the final size

	// current is the line column; the pending word and space runs are [start:end] slices of s
	var current, wordLen, spaceLen uint
	wordStart, wordEnd := 0, 0
	spaceStart, spaceEnd := 0, 0

	// flush helpers append a pending run straight from the source string
	flushSpace := func() {
		if spaceLen > 0 {
			b.WriteString(s[spaceStart:spaceEnd])
		}
	}
	flushWord := func() {
		if wordLen > 0 {
			b.WriteString(s[wordStart:wordEnd])
		}
	}

	for i := 0; i < len(s); {
		char, size := utf8.DecodeRuneInString(s[i:]) // size advances i at the foot of the loop
		switch {
		case char == '\n': // hard break: emit any pending runs that fit, then reset the column
			if wordLen == 0 {
				if current+spaceLen > lim {
					current = 0
				} else {
					current += spaceLen
					flushSpace()
				}
				spaceLen = 0
			} else {
				current += spaceLen + wordLen
				flushSpace()
				spaceLen = 0
				flushWord()
				wordLen = 0
			}
			b.WriteByte('\n')
			current = 0
		case unicode.IsSpace(char) && char != nbsp: // whitespace (nbsp counts as a word char)
			if spaceLen == 0 || wordLen > 0 { // a word just ended: flush both runs
				current += spaceLen + wordLen
				flushSpace()
				spaceLen = 0
				flushWord()
				wordLen = 0
			}
			if spaceLen == 0 {
				spaceStart = i
			}
			spaceEnd = i + size
			spaceLen++
		default: // word character: extend the word, wrapping if the line would overrun
			if wordLen == 0 {
				wordStart = i
			}
			wordEnd = i + size
			wordLen++
			if current+wordLen+spaceLen > lim && wordLen < lim { // wrap mid-word, drop pending spaces
				b.WriteByte('\n')
				current = 0
				spaceLen = 0
			}
		}
		i += size
	}

	// flush whatever run is still pending at end of input
	if wordLen == 0 {
		if current+spaceLen <= lim {
			flushSpace()
		}
	} else {
		flushSpace()
		flushWord()
	}
	return b.String()
}

func main() {
	input := "Go is an open source programming language. It makes it easy to build simple, reliable, and efficient software."
	fmt.Println(WrapString(input, 20))
}
