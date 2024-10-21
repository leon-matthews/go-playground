// Examples and experiments from Section 4.1
package main

import (
	"crypto/sha256"
	"fmt"
)

func main() {
	basics()
	iterate()
	literals()
	cryptoHash()
	pointerOrValue()
}

func basics() {
	// Initialised with zero values for type
	var a [3]int
	fmt.Println(a)           // [0 0 0]
	fmt.Println(a[len(a)-1]) // 0, last element
}

func iterate() {
	// Iterate over...
	// ...indices and values
	var a [3]int
	for i, v := range a {
		fmt.Printf("%v:%v ", i, v)
	}
	fmt.Println()

	// ...just indices
	for i := range a {
		fmt.Printf("%v ", i)
	}
	fmt.Println()

	// ...just once per element
	for range a {
		fmt.Printf("element ")
	}
	fmt.Println()
}

func literals() {
	// Array literals
	var q [3]int = [3]int{1, 2, 3} // Note that length is part of type
	q2 := [...]int{1, 2, 3}        // Short declaration is short
	fmt.Println(q, q2)

	// Arrays ARE directly comparable, if their elements are, and
	// their types (including length, remember) match.
	fmt.Println(q == q2)

	// Indices can be specified, zero values will be added as needed
	big := [...]int{9: -1} // All zeros except index 9 which is -1
	fmt.Println(big)       // [0 0 0 0 0 0 0 0 0 -1]

	// This allows sort-of-kind-of map behaviour
	type Currency int
	const (
		NZD Currency = iota
		EUR
		GBP
	)
	symbol := [...]string{NZD: "$", EUR: "€", GBP: "£"}

	fmt.Println(EUR, symbol[EUR])
	fmt.Println(GBP, symbol[GBP])
	fmt.Println(NZD, symbol[NZD])
}

// Arrays are not often used directly, but cryptographic hash functions do
func cryptoHash() {
	hash1 := sha256.Sum256([]byte("l"))
	hash2 := sha256.Sum256([]byte("I"))
	fmt.Printf("%x\n%x\n", hash1, hash2)
	fmt.Printf("%T %T\n", hash1, hash2)		// [32]uint8 [32]uint8
	fmt.Println(hash1 == hash2)				// false
}

// Arrays are copied by default
func pointerOrValue() {
	name := []byte("Leon")
	hash := sha256.Sum256(name)
	fmt.Printf("%T %[1]v -> %T %[2]x\n", name, hash)

	zero(&hash)
	fmt.Printf("%[1]x\n", hash)
}

// Zero contents of a `[32]byte` array.
func zero(ptr *[32]byte) {
	for i := range ptr {
		ptr[i] = 0
	}
}

// Same as [zero] but creates new array
func zero2(ptr *[32]byte) {
	*ptr = [32]byte{}
}
