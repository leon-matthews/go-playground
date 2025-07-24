package main

import (
	"fmt"
	"slices"
)

func main() {
	// An uninitialized slice: equals to nil, length 0
	var i []int
	fmt.Println("uninitialised:", i, i == nil, len(i) == 0)

	// Create an empty slice with non-zero length
	i = make([]int, 3)
	fmt.Println("make:", i, "len:", len(i), "cap:", cap(i))

	// Set & get like arrays
	s := make([]string, 3)
	s[0] = "a"
	s[1] = "b"
	s[2] = "c"
	fmt.Println("set/get:", s, "len:", len(s), "cap:", cap(s))

	// Append one, then many elements
	s = append(s, "d")
	s = append(s, "e", "f", "g")
	fmt.Println(s, "len:", len(s), "cap:", cap(s))

	// Copy requires a non-empty destination slice
	c := make([]string, len(s))
	copy(c, s)
	fmt.Println(c, "len:", len(c), "cap:", cap(c))

	// Clone
	c2 := slices.Clone(s)
	fmt.Println(c2, "len:", len(c2), "cap:", cap(c2))

	// Slice the slice!
	l := s[2:5]
	fmt.Println(l)

	// The `slices` package contains useful utility functions
	fmt.Println("slice == copied", slices.Equal(s, c))
}
