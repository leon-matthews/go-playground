package main

import (
	//~ "bytes"
	"fmt"
	"regexp"
)

func main() {
	matchExamples()
	findExamples()
	replaceExamples()
}

func matchExamples() {
	// Simple match, no compilation
	match, _ := regexp.MatchString("p([a-z]+)ch", "peach")
	fmt.Println("Peach found:", match)

	// Compile first, then use MatchString() as method
	r, _ := regexp.Compile("p([a-z]+)ch")
	fmt.Println("Another peach found:", r.MatchString("peach"))
}

// There are 16 methods of Regexp that match a regular expression and identify
// the matched text. Their names are matched by this regular expression:
//
//	Find(All)?(String)?(Submatch)?(Index)?
func findExamples() {
	r, _ := regexp.Compile("p([a-z]+)ch")

	// FindString() returns a string holding the text of the leftmost match, or an empty string.
	// The all variant returns string slice of all matches, optionally limited if n >= 0
	fmt.Println(r.FindString("peach punch"))
	fmt.Println(r.FindAllString("peach punch", -1))

	// Returns slice of strings, first-match, then subexpression
	fmt.Println(r.FindStringSubmatch("peach punch"))
	fmt.Println(r.FindAllStringSubmatch("peach punch", -1))

	// FindStringIndex() returns start and end indexes for first match, or nil.
	fmt.Println("idx:", r.FindStringIndex("peach punch"))
	fmt.Println("idx:", r.FindAllStringIndex("peach punch", -1))

	// Dropping 'String' from the name matches against []byte instead
	fmt.Println(r.Match([]byte("peach")))
	fmt.Println(r.FindAll([]byte("peach punch"), -1))
}

func replaceExamples() {
	r, _ := regexp.Compile("p([a-z]+)h")
	// ReplaceAll expands matches $name or ${name}, with indexes from one for unnamed groups
	fmt.Println(r.ReplaceAllString("life is a peach", "b${1}h"))

	// The Literal variants perform no expansion
	fmt.Println(r.ReplaceAllLiteralString("life is a peach", "beach"))
}
