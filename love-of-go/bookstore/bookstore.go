package bookstore

import (
	"errors"
	"fmt"
)

var NoCopiesLeftError = errors.New("no copies left")
var NotFoundError = errors.New("no book found")

type Book struct {
	Title           string
	Author          string
	Copies          int
	ID              int
	PriceCents      int
	DiscountPercent int
}

type Catalogue map[int]Book

func Buy(b Book) (Book, error) {
	if b.Copies == 0 {
		return Book{}, NoCopiesLeftError
	}
	b.Copies--
	return b, nil
}

func (c Catalogue) GetAllBooks() []Book {
	books := make([]Book, 0, len(c))
	for _, b := range c {
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

func (b Book) NetPriceCents() int {
	saving := b.PriceCents * b.DiscountPercent / 100
	return b.PriceCents - saving
}
