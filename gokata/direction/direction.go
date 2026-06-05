// Package direction shows idiomatic way to implement an enumerated type:
//
//  1. Create a new integer type.
//  2. List its values using iota.
//  3. Give the type a String function.
//
// Based on https://yourbasic.org/golang/iota.
package direction

type Cardinal int

// Cardinal directions.
const (
	North Cardinal = iota
	South
	East
	West
)

func (c Cardinal) String() string {
	// An array is strictly more efficient than a slice here.
	// I checked the assembly for Go 1.26.2. In both cases the strings are baked
	// into the binary's data section, but using a slice requires rebuilding the slice's
	// header (on the stack, not the heap) for every function invocation.
	return [...]string{"North", "South", "East", "West"}[c]
}
