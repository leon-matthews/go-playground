package main

import (
	"fmt"
	"strconv"
	"strings"
	"unsafe"
)

func main() {
	var v IntSlice = []int{1, 2, 3}
	fmt.Printf("IntSlice: [%T] %[1]v (%d bytes)\n", v, unsafe.Sizeof(v))

	// Converted to interface? Yes. %T still reports the concrete type, but size now 2 * sizeof(ptr)
	var s fmt.Stringer = v
	fmt.Printf("assign to fmt.Stringer: [%T] %[1]v (%d bytes)\n", s, unsafe.Sizeof(s))

	// Passing concrete type to interface has same effect as assignment
	printSize(v)

	// Assignment works if underlying type is the same - output format changes
	var is []int = v
	fmt.Printf("assign to []int: [%T] %[1]v (%d bytes)\n", is, unsafe.Sizeof(is))

	printSize(15)
}

// IntSlice is a basic user-declared type, with underying type []int
type IntSlice []int

// String adds semicolons to the output format
func (is IntSlice) String() string {
	var strs []string
	for _, v := range is {
		strs = append(strs, strconv.Itoa(v))
	}
	return "[" + strings.Join(strs, ";") + "]"
}

func printSize(s fmt.Stringer) {
	fmt.Printf("assign to fmt.Stringer: [%T] %[1]v (%d bytes)\n", s, unsafe.Sizeof(s))
}
