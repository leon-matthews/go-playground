package books

import (
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"slices"
)

// Catalogue manages many books, keyed by ID
type Catalogue map[string]Book

// AllBooks build a plain slice of Book values
func (c Catalogue) AllBooks() []Book {
	return slices.Collect(maps.Values(c))
}

// AddBook inserts a new title into catalogue
func (c Catalogue) AddBook(book Book) error {
	_, ok := c[book.ID]
	if ok {
		return fmt.Errorf("ID %q already exists", book.ID)
	}
	c[book.ID] = book
	return nil
}

// GetBook tries to find the book with the given ID, using the comma-ok idiom.
func (c Catalogue) GetBook(id string) (Book, bool) {
	book, ok := c[id]
	return book, ok
}

// SetCopies updates the count of held copies for the given ID
func (c Catalogue) SetCopies(id string, copies int) (Book, error) {
	book, ok := c.GetBook(id)
	if !ok {
		return Book{}, fmt.Errorf("ID %q not found", id)
	}
	err := book.SetCopies(copies)
	if err != nil {
		return Book{}, err
	}
	c[id] = book
	return book, nil
}

// Sync writes catalogue data to JSON file
func (c Catalogue) Sync(path string) error {
	// Create new file or truncate existing
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write catalogue as JSON
	err = json.NewEncoder(file).Encode(c)
	if err != nil {
		return err
	}

	return nil
}

// OpenCatalogue reads catalogue data from a JSON file.
func OpenCatalogue(path string) (Catalogue, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening catalogue: %w", err)
	}
	defer file.Close()

	catalog := Catalogue{}
	err = json.NewDecoder(file).Decode(&catalog)
	if err != nil {
		return nil, fmt.Errorf("reading catalogue: %w", err)
	}
	return catalog, nil
}
