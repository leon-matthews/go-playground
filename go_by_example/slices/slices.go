package main

import (
	"fmt"
	"slices"
)

func main() {
	// An uninitialized slice equals to nil and has length 0
	var s[]string
	fmt.Println("uninitialised:", s, s == nil, len(s) == 0)

	// Create an empty slice with non-zero length
	s = make([]string, 3)
    fmt.Println("make:", s, "len:", len(s), "cap:", cap(s))

    // Set & get like arrays
    s[0] = "a"
    s[1] = "b"
    s[2] = "c"
    fmt.Println("set/get:", s, "len:", len(s), "cap:", cap(s))

    // Append one, then many elements
    s = append(s, "d")
    s = append(s, "e", "f", "g")
    fmt.Println(s)

    // Copy requires a non-empty destination slice
	c := make([]string, len(s))
    copy(c, s)
    fmt.Println(c)

    // Slice the slice!
    l := s[2:5]
	fmt.Println(l)

    // The `slices` package contains useful utility functions
	fmt.Println("slice == copied", slices.Equal(s, c))
}
