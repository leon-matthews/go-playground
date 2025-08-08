package main

import (
	"fmt"
	"sort"
)

// type sort.Interface interface {
//     Len() int
//     Less(i, j int) bool
//     Swap(i, j int)
// }

func main() {
	// Strings
	// sort.StringSlice has underlying type []string
	// It attaches methods to implement [sort.Interface]
	names := []string{"Leon", "Alyson", "Blake", "Stella", "Teddy", "Velma"}
	sort.Sort(sort.StringSlice(names)) // Type case attaches methods
	fmt.Println(names)                 // Sorted in-place
}
