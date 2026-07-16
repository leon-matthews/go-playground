// Package main implements a light-weight CLI to list all books available
package main

import (
	"fmt"
	"os"

	"local.dev/books"
)

func main() {
	catalogue, err := books.OpenCatalogue("testdata/catalogue.json")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	for _, book := range catalogue.AllBooks() {
		fmt.Println(book)
	}
}
