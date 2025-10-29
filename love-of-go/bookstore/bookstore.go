package bookstore

import (
	"errors"
	"fmt"
)

var NoCopiesLeftError = errors.New("no copies left")
var NotFoundError = errors.New("no book found")

type Book struct {
	Title  string
	Author string
	Copies int
	ID     int
}

func Buy(b Book) (Book, error) {
	if b.Copies == 0 {
		return Book{}, NoCopiesLeftError
	}
	b.Copies--
	return b, nil
}

func GetAllBooks(catalogue map[int]Book) []Book {
	books := make([]Book, 0, len(catalogue))
	for _, b := range catalogue {
		books = append(books, b)
	}
	return books
}

func GetBook(catalogue map[int]Book, id int) (Book, error) {
	b, ok := catalogue[id]
	if !ok {
		return b, fmt.Errorf("with id=%d: %w", id, NotFoundError)
	}
	return b, nil
}
