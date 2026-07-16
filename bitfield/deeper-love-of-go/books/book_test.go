package books_test

import (
	"testing"

	"local.dev/books"
)

func TestString_FormatsBookInfoAsString(t *testing.T) {
	t.Parallel()
	b := books.Book{
		Title:  "Sea Room",
		Author: "Adam Nicholson",
		Copies: 2,
	}
	want := "Sea Room by Adam Nicholson (copies: 2)"
	got := b.String()
	if want != got {
		t.Fatalf("want %v, got %v", want, got)
	}
}

func TestSetCopies_(t *testing.T) {
	t.Parallel()

	t.Run("SetsNumberOfCopiesToGivenValue", func(t *testing.T) {
		t.Parallel()
		book := books.Book{
			Copies: 5,
		}
		err := book.SetCopies(12)
		if err != nil {
			t.Fatal(err)
		}
		if book.Copies != 12 {
			t.Errorf("want 12 copies, got %d", book.Copies)
		}
	})

	t.Run("ReturnsErrorIfCopiesNegative", func(t *testing.T) {
		t.Parallel()
		book := books.Book{}
		err := book.SetCopies(-1)
		if err == nil {
			t.Error("want error for negative copies, got nil")
		}
	})
}
