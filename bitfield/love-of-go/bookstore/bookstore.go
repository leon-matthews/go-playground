package bookstore

import (
	"errors"
	"fmt"
)

type Category int

const (
	CategoryAutobiography Category = iota
	CategoryLargePrintRomance
	CategoryParticlePhysics
)

var validCategory = map[Category]bool{
	CategoryAutobiography:     true,
	CategoryLargePrintRomance: true,
	CategoryParticlePhysics:   true,
}

type Book struct {
	Title           string
	Author          string
	Copies          int
	ID              int
	PriceCents      int
	DiscountPercent int

	category Category
}

type Catalogue map[int]Book

func Buy(b Book) (Book, error) {
	if b.Copies == 0 {
		return Book{}, errors.New("no copies left")
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

func (c Catalogue) GetBook(id int) (Book, error) {
	b, ok := c[id]
	if !ok {
		return b, fmt.Errorf("no book found with id=%d", id)
	}
	return b, nil
}

func (b *Book) Category() Category {
	return b.category
}

func (b *Book) NetPriceCents() int {
	saving := b.PriceCents * b.DiscountPercent / 100
	return b.PriceCents - saving
}

func (b *Book) SetCategory(category Category) error {
	if !validCategory[category] {
		return fmt.Errorf("unknown category: %v", category)
	}
	b.category = category
	return nil
}

func (b *Book) SetPriceCents(cents int) error {
	if cents < 0 {
		return errors.New("negative price cents")
	}
	b.PriceCents = cents
	return nil
}
