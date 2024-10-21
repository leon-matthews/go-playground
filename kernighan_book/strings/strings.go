// Experiments with strings, slices, and Unicode.
package main

import (
	"bytes"
	"fmt"
	"strconv"
)

func main() {
	substrings()
	unicode()
	byteSlices()
	byteBuffers()
	strconvExamples()
}

// Do we call them string slices in Go, or is that too confusing?
func substrings() {
	s := "hello, world"
	fmt.Println(len(s))     // 12
	fmt.Println(s[0], s[7]) // 104 113 (ASCII/UTF-8 for 'h' & 'w')

	// Index error causes panic
	// panic: runtime error: index out of range [12] with length 12
	//~ c := s[len(s)]

	// Strings are immutable
	// cannot assign to s[7] (neither addressable nor a map index expression)
	//~ s[7] = 'W'

	// Substring operations can share underlying data
	fmt.Println(s[0:5]) // hello
	fmt.Println(s[7:])  // world
	fmt.Println(s[:])   // hello, world

	// Raw strings
	fmt.Println(`This is a
		multiline
		string`)
}

func unicode() {
	// Runes
	fmt.Printf("%q %q\n", '\u4e16', '\U00004e16') // '世' '世'

	// Konnichiwa!
	s := "こんにちは"
	fmt.Println(len(s))

	// Loop over bytes
	for i := range len(s) {
		fmt.Printf("%v ", s[i])
	}
	fmt.Println()

	// Loop over runes
	for i, r := range s {
		fmt.Printf("%v: %d, %q\n", i, r, r)
	}

	// Invalid interior indices cause invalid data, not a panic
	fmt.Println(s[2:5]) // �� - two unicode replacement characters
	fmt.Println(s[3:6]) // ん - one valid rune
}

// While strings are immutable, byte slices are not
func byteSlices() {
	// Easy conversion, although data copied in both directions
	s := "abc"
	b := []byte(s)
	b[0] = 'A' // Constant used as byte/uint8, not rune
	s2 := string(b)
	fmt.Println(b, s2) // [65 98 99] Abc

	// Ni Hao!
	b2 := []byte("你好")
	fmt.Println(b2) // [228 189 160 229 165 189]
}

// A nice `io.Writer` implementation
func byteBuffers() {
	var buf bytes.Buffer
	buf.WriteByte('[')
	buf.WriteString("Leon")
	buf.WriteByte(']')
	fmt.Println(buf.String())	// "[Leon]"
}

// Basics of the `strconv` package
func strconvExamples() {
	// Integer-to-ASCII
	x := 123
	y := fmt.Sprintf("%d", x)
	fmt.Println(y, strconv.Itoa(x))	// 123 123

	// Change base
	for i := 2; i <= 16; i++ {
		s := strconv.FormatInt(int64(x), i)
		fmt.Printf("%d in base-%d = %s\n", x, i, s)
	}

	// ASCII to integer
	x, err := strconv.Atoi("12345")
	fmt.Println(x, err)		// 12345 <nil>

	// Base-10, up to 64-bits
	x2, err := strconv.ParseInt("54321", 10, 64)
	fmt.Println(x2, err)	// 54321 <nil>
}
