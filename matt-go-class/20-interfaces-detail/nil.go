// An interface's zero type is nil
package main

import (
	"bytes"
	"fmt"
	"io"
)

func main() {
	var r io.Reader     // interface type, nil until initialised
	var b *bytes.Buffer // also nil until initialised
	fmt.Printf("[%T]%+[1]v\n", r)
	fmt.Printf("[%T]%+[1]v\n", b)

	// Assign nil concrete type to interface type
	r = b

	// But now r is NOT nil, because it has a valid concrete type, even
	// though its value is still nil
	if r == nil {
		fmt.Println("r is nil")
	} else {
		fmt.Println("r is NOT nil")
	}
}
