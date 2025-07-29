package main

import (
	"fmt"
)

// Abort entire program if a function returns an error value that we don’t
// know how to (or want to) handle.
func main() {
	panic("Hello, death")

	fmt.Println("You'll never see me")
}
