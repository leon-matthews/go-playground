package main

import (
	"fmt"
	"slices"
)

// Sorting functions are generic, and work for any ordered built-in type.
// For a list of ordered types, see [cmp.Ordered]
// [cmp.Ordered]: https://pkg.go.dev/cmp#Ordered
func main() {
	// [slices.Sort] sorts in-place
	ints := []int{7, 2, 4}
	slices.Sort(ints)
	fmt.Println("Ints:   ", ints)

	strs := []string{"c", "a", "b"}
	slices.Sort(strs)
	fmt.Println("Strings:", strs)

	// We can to a quick check to see if our slice (of any type) is sorted
	fmt.Println("Ints Sorted:", slices.IsSorted(ints))
	fmt.Println("Strings Sorted:", slices.IsSorted(strs))
}
