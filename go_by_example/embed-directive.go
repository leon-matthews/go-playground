package main

import (
	"embed"
	"fmt"
)

// embed directives accept paths relative to the directory containing the Go source file.

//go:embed README.md
var readme string

//go:embed embed-directive.go
var self2 []byte

// We can also embed multiple files or even folders with wildcards.
//
//go:embed *.go
var allGoFiles embed.FS

func main() {
	// Files
	fmt.Println(readme)
	fmt.Printf("This source file is %d bytes\n", len(self2))

	// File system
	entries, err := allGoFiles.ReadDir(".")
	if err != nil {
		panic(err)
	}
	fmt.Printf("Found %d Go files in embed.FS\n", len(entries))
}
