package main

import (
	"fmt"
	"slices"
)

func main() {
	ArrayDeclarations()
	SliceDeclarations()
	SliceAppend()
	SliceMake()
	SliceClear()
	SlicingSlices()
	CopySlices()
	ArrayToSlice()
	SliceToArray()
}

func ArrayDeclarations() {
	// Three elements with zero values
	var x [3]int
	fmt.Println(x)

	// Array literal
	var y = [3]int{5, 10, 20}
	fmt.Println(y)

	// Sparse array
	var sparse = [11]byte{1, 5: 1, 10: 1}
	fmt.Println(sparse)

	// Let the compilier count literal values
	var lazy = [...]uint64{2, 3, 5, 7, 11, 13, 17, 19}
	fmt.Println(lazy, len(lazy), "primes under 20")

	// Arrays are comparable
	var equivilant = [...]byte{1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1}
	fmt.Println("Are `sparse` and `equivilant` the same?", sparse == equivilant)
}

func SliceDeclarations() {
	// Slice literal (note absence of `...`)
	var primes = []byte{2, 3, 5, 7, 11, 13, 17, 19}
	fmt.Println(primes, len(primes), "primes under 20")

	// Sparse slice
	var sparse = []byte{1, 5: 1, 10: 1}
	fmt.Println(sparse)

	// Slice without literal is `nil` - a nil slice contains nothing
	var empty []uint
	fmt.Println(empty == nil)

	// Slices cannot be directly compared, except against `nil`
	var primes2 = []byte{2, 3, 5, 7, 11, 13, 17, 19}
	// invalid operation: primes == primes2 (slice can only be compared to nil)
	//~ fmt.Println(primes == primes2)

	// The package `slices` introduced in Go 1.21 does an
	// element-by-element comparison
	fmt.Println(slices.Equal(primes, primes2))
}

// Built-in `append()` function
func SliceAppend() {
	var grow []bool
	fmt.Println(grow, "length:", len(grow), "capacity:", cap(grow))

	// Append single value
	grow = append(grow, true)
	fmt.Println(grow, "length:", len(grow), "capacity:", cap(grow))

	// Append multiple values
	grow = append(grow, false, true)
	fmt.Println(grow, "length:", len(grow), "capacity:", cap(grow))

	// Extend slice using `...` operator
	thruthy := []bool{true, true, true}
	grow = append(grow, thruthy...)
	fmt.Println(grow, "length:", len(grow), "capacity:", cap(grow))
}

// Built-in `make()` function
func SliceMake() {
	// 8 zeros, room for zero more
	x := make([]int, 8)
	fmt.Println(x, "length:", len(x), "capacity:", cap(x))

	// 8 zeros, room for 8 more
	y := make([]int, 8, 16)
	fmt.Println(y, "length:", len(y), "capacity:", cap(y))

	// 0 zeros, room for 8 more
	z := make([]int, 0, 8)
	fmt.Println(z, "length:", len(z), "capacity:", cap(z))
}

// Built-in `clear()` function
func SliceClear() {
	s := []string{"first", "second", "third"}
	fmt.Println(s, "length:", len(s), "capacity:", cap(s))

	// Set all elements to their zero value (an empty string in this case)
	clear(s)

	// The length and capacity stay the same
	fmt.Println(s, "length:", len(s), "capacity:", cap(s))
}

// You can slice a slice! But slices share memory.
func SlicingSlices() {
	s := []uint{2, 3, 5, 7, 11, 13, 17, 19, 23, 29}
	first := s[0]
	fmt.Println(first)
}

func CopySlices() {
	x := []int{1, 2, 3, 4, 5}
	y := make([]int, 10)
	num := copy(y, x) // Think: y = x
	fmt.Println(num, y)
}

func ArrayToSlice() {
	x := [...]int{9, 8, 7, 6, 5, 4, 3, 2, 1}
	fmt.Printf("%T %[1]v\n", x)

	// Taking slice of an array makes a s
	s := x[:]
	fmt.Printf("%T %[1]p %[1]v\n", s)

	// Memory is shared
	s[5] = 42
	fmt.Println(cap(s))
	fmt.Printf("%T %[1]p %[1]v\n", s)
	fmt.Println(s, x)
}

func SliceToArray() {
	s := []int{1, 2, 3, 4, 5}
	x := [5]int(s)
	fmt.Println(s, x)
}
