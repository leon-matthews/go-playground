package main

import (
	"fmt"
	"os"
)

type point struct {
	x, y int
}

func main() {
	p := point{1, 2}

	// The %v (value) verb prints using a default format
	fmt.Printf("struct1: %v\n", p)

	// %+v verb variant will include the struct’s field names.
	fmt.Printf("struct2: %+v\n", p)

	// %#v verb variant prints a literal, like Python's repr()
	fmt.Printf("struct3: %#v\n", p)

	// %T prints the type of the value
	fmt.Printf("type: %T\n", p)

	// %t for booleans
	fmt.Printf("bool: %t\n", true)

	// %d for base-10 integers
	fmt.Printf("int: %d\n", 123)

	// %b for a binary
	fmt.Printf("bin: %b\n", 14)

	// %c character from integer
	fmt.Printf("char: %c\n", 33)

	// %x for hex (%X for upper-case)
	fmt.Printf("hex: %x\n", 456)

	// %f for basic decimal formatting
	fmt.Printf("float1: %f\n", 78.9)

	// %e for scientific notation
	fmt.Printf("float2: %e\n", 123400000.0)
	fmt.Printf("float3: %E\n", 123400000.0)

	// %s for basic string formatting
	fmt.Printf("str1: %s\n", "\"string\"")

	// %q to double quote strings
	fmt.Printf("str2: %q\n", "\"string\"")

	// %x for strings is base-16 with two characters per byte of input
	fmt.Printf("str3: %x\n", "hex this")

	// %p for pointers
	fmt.Printf("pointer: %p\n", &p)

	// %6d to add width to integer representation (space-padded, right-aligned)
	fmt.Printf("width1: |%6d|%6d|\n", 12, 345)

	// %6.2f adds width and decimal precision for floats
	fmt.Printf("width2: |%6.2f|%6.2f|\n", 1.2, 3.45)

	// Add hyphen for left-alignment
	fmt.Printf("width3: |%-6.2f|%-6.2f|\n", 1.2, 3.45)

	// Widths and alignment can also be specified for strings
	fmt.Printf("width4: |%6s|%6s|\n", "foo", "b")
	fmt.Printf("width5: |%-6s|%-6s|\n", "foo", "b")

	// Build a new string
	s := fmt.Sprintf("sprintf: a %s", "string")
	fmt.Println(s)

	// Write to io.Writers other than os.Stdout
	fmt.Fprintf(os.Stderr, "io: an %s\n", "error")
}
