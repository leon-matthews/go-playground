package main

import (
	//~ "bytes"
	"fmt"
	"regexp"
)



func main() {
	// Simple match
	match, _ := regexp.MatchString("p([a-z]+)ch", "peach")
	fmt.Println("Peach found:", match)

	// Compile first
	r, _ := regexp.Compile("p([a-z]+)ch")

	// MatchString() again, as a method this time.
    fmt.Println("Another peach found:", r.MatchString("peach"))

    // There are 16 methods of Regexp that match a regular expression and identify
	// the matched text. Their names are matched by this regular expression:
	//   Find(All)?(String)?(Submatch)?(Index)?

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
