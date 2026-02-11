// Fuzzing tutorial from: https://go.dev/doc/tutorial/fuzz
package main

import (
	"errors"
	"fmt"
	"unicode/utf8"
)

func main() {
	input := "Sphinx of black quartz, judge my vow"
	reversed, _ := Reverse(input)
	double, _ := Reverse(reversed)
	fmt.Println("original:", input)
	fmt.Println("reversed:", reversed)
	fmt.Println("reversed again:", double)
}

func Reverse(s string) (string, error) {
	if !utf8.ValidString(s) {
		return s, errors.New("input is not valid UTF-8")
	}
	b := []rune(s)
	for i, j := 0, len(b)-1; i < len(b)/2; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}
	return string(b), nil
}
