// Examples, exercises, and experiments from Chapter 13, The Standard Library
package main

import (
	"fmt"
	"io"
	"strings"
)

func main() {
	readerWriter()
}

// Both [io.Reader] and [io.Writer] define a single method:
//
//	 type Reader interface {
//			Read(p []byte) (n int, err error)
//	 }
//
//	 type Writer interface {
//	 	Write(p []byte) (n int, err error)
//	 }
//
// The functions take a slice of bytes, returns the number of bytes processed,
// and an error if something went wrong.
func readerWriter() error {
	pangrams :=
		"The quick brown fox jumped over the lazy dog\n" +
		"Pack my box with five dozen liquor jugs\n" +
		"Jackdaws love my big sphinx of quartz\n" +
		"The five boxing wizards jump quickly\n" +
		"Sphinx of black quartz, judge my vow\n" +
		"Blowzy night-frumps vex'd Jack Q"

	reader := strings.NewReader(pangrams)
	counts, err := countLetters(reader)
	if err != nil {
		return err
	}
	fmt.Println(counts)

	return nil
}

// Build
func countLetters(r io.Reader) (map[string]int, error) {
	buf := make([]byte, 64) // Create once then reuse
	out := map[string]int{}
	for {
		n, err := r.Read(buf)
		fmt.Println("Read", n, "bytes. Error:", err)
		for _, b := range buf[:n] {
			if (b >= 'A' && b <= 'Z') || (b >= 'a' && b <= 'z') {
				out[string(b)]++
			}
		}

		// Expected error
		if err == io.EOF {
			return out, nil
		}

		// Unexpected error
		if err != nil {
			return nil, err
		}
	}

	return out, nil
}
