// Package books manages stock for a bookshop
package books

import (
	"fmt"
)

// Book hold details for a unique title
type Book struct {
	ID     string
	Title  string
	Author string
	Copies int

	path string
}

// SetCopies sets the number of copies held
func (b *Book) SetCopies(copies int) error {
	if copies < 0 {
		return fmt.Errorf("negative copies: %d", copies)
	}
	b.Copies = copies
	return nil
}

func (b Book) String() string {
	return fmt.Sprintf("%v by %v (copies: %v)", b.Title, b.Author, b.Copies)
}
