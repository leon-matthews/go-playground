// Package main implements a light-weight CLI to find a particular book
package main

import (
	"fmt"
	"os"

	"local.dev/books"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: find <ID>")
		return
	}
	id := os.Args[1]

	catalogue, err := books.OpenCatalogue("testdata/catalogue.json")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	book, ok := catalogue.GetBook(id)
	if !ok {
		fmt.Fprintln(os.Stderr, "Sorry, I couldn't find that book in the catalog.")
		return
	}
	fmt.Println(book)
}
