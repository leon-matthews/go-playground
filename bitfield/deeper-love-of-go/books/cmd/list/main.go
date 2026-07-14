package main

import (
	"fmt"

	"local.dev/books"
)

func main() {
	catalogue := books.GetCatalogue()
	for _, book := range books.GetAllBooks(catalogue) {
		fmt.Println(books.BookToString(book))
	}
}
