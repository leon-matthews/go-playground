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

	catalogue := books.GetCatalogue()
	book, ok := books.GetBook(catalogue, id)
	if !ok {
		fmt.Fprintln(os.Stderr, "Sorry, I couldn't find that book in the catalog.")
		return
	}
	fmt.Println(books.BookToString(book))
}
