// Package main implements an HTTP server
package main

import(
	"fmt"
	"os"

	"local.dev/books"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "Usage: server <CATALOG FILE>")
		return
	}

	catalogue, err := books.OpenCatalogue(os.Args[1])
	if err != nil {
		fmt.Printf("opening catalog: %v\n", err)
		return
	}

	books.ListenAndServe(":8000", catalogue)
}
