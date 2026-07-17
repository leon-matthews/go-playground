// Package main implements CLI to change copies count for given id
package main

import (
	"fmt"
	"os"
	"strconv"

	"local.dev/books"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "Usage: copies <ID> <COPIES>")
		return
	}
	id := os.Args[1]
	copies, err := strconv.Atoi(os.Args[2])
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid value for copies: %q", os.Args[2])
		return
	}

	catalogue, err := books.OpenCatalogue("testdata/catalogue.json")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	book, err := catalogue.SetCopies(id, copies)
	if err != nil {
		fmt.Fprintf(os.Stderr, "updating copies: %v\n", err)
		return
	}
	fmt.Println(book)

	err = catalogue.Sync()
	if err != nil {
		fmt.Fprintf(os.Stderr, "saving catalogue: %v\n", err)
	}
}
