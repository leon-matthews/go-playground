package main

import (
	"cmp"
	"fmt"
)

func main() {
	// Tree of ints
	t := NewTree(cmp.Compare[int])
	t.Add(13)
	t.Add(16)
	t.Add(48)
	t.Add(49)
	fmt.Printf("Contains 13 %v\n", t.Contains(13))
	fmt.Printf("Contains 17 %v\n", t.Contains(17))

	// Tree of people

}
