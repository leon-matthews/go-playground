package books

import (
	"fmt"
	"maps"
	"slices"
)

type Book struct {
	ID     string
	Title  string
	Author string
	Copies int
}

var catalogue = map[string]Book{
	"abc": {
		ID:     "abc",
		Title:  "In the Company of Cheerful Ladies",
		Author: "Alexander McCall Smith",
		Copies: 1,
	},
	"def": {
		ID:     "def",
		Title:  "White Heat",
		Author: "Dominic Sandbrook",
		Copies: 2,
	},
}

func AddBook(catalogue map[string]Book, book Book) {
	catalogue[book.ID] = book
}

func BookToString(book Book) string {
	return fmt.Sprintf("%v by %v (copies: %v)", book.Title, book.Author, book.Copies)
}

func GetAllBooks(catalogue map[string]Book) []Book {
	return slices.Collect(maps.Values(catalogue))
}

func GetBook(catalogue map[string]Book, id string) (Book, bool) {
	book, ok := catalogue[id]
	return book, ok
}

func GetCatalogue() map[string]Book {
	return catalogue
}
