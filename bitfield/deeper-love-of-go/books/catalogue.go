package books

import (
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"slices"
	"sync"
)

// Catalogue manages many books, keyed by ID
type Catalogue struct {
	Path string
	
	data map[string]Book
	mu *sync.RWMutex
}

// AllBooks build a plain slice of Book values
func (c *Catalogue) AllBooks() []Book {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return slices.Collect(maps.Values(c.data))
}

// AddBook inserts a new title into catalogue
func (c *Catalogue) AddBook(book Book) error {
	c.mu.Lock()
	c.mu.Unlock()
	_, ok := c.data[book.ID]
	if ok {
		return fmt.Errorf("ID %q already exists", book.ID)
	}
	c.data[book.ID] = book
	return nil
}

// GetBook tries to find the book with the given ID, using the comma-ok idiom.
func (c *Catalogue) GetBook(id string) (Book, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	book, ok := c.data[id]
	return book, ok
}

// Copies returns the number of copies held for the given ID
func (c *Catalogue) Copies(id string) (int, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	book, ok := c.data[id]
	if !ok {
		return 0, fmt.Errorf("ID %q not found", id)
	}
	return book.Copies, nil
}

// SetCopies updates the count of held copies for the given ID
// Returns the updated book value
func (c *Catalogue) SetCopies(id string, copies int) (Book, error) {
	book, ok := c.GetBook(id)
	if !ok {
		return Book{}, fmt.Errorf("ID %q not found", id)
	}
	
	c.mu.Lock()
	defer c.mu.Unlock()
	err := book.SetCopies(copies)
	if err != nil {
		return Book{}, err
	}
	c.data[id] = book
	return book, nil
}

// Sync writes catalogue data to JSON file
func (c *Catalogue) Sync() error {
	// Create new file or truncate existing
	file, err := os.Create(c.Path)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write catalogue data as JSON file
	c.mu.RLock()
	defer c.mu.RUnlock()
	err = json.NewEncoder(file).Encode(c.data)
	if err != nil {
		return err
	}

	return nil
}

// OpenCatalogue reads catalogue data from a JSON file.
func OpenCatalogue(path string) (*Catalogue, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening catalogue: %w", err)
	}
	defer file.Close()

	catalogue := NewCatalogue()
	err = json.NewDecoder(file).Decode(&catalogue.data)
	if err != nil {
		return nil, fmt.Errorf("reading catalogue: %w", err)
	}
	catalogue.Path = path
	return catalogue, nil
}

// NewCatalogue creates a new, empty catalogue
func NewCatalogue() *Catalogue {
	c := Catalogue{
		mu: &sync.RWMutex{},
		data: map[string]Book{},
	}
	return &c
}
