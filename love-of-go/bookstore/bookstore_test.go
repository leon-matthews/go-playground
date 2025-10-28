package bookstore_test

import (
	"testing"

	"bookstore"
)

func TestBuy(t *testing.T) {
	t.Parallel()
	book := bookstore.Book{
		Title:  "Spark Joy",
		Author: "Marie Kondo",
		Copies: 2,
	}
	want := 1
	result := bookstore.Buy(book)
	got := result.Copies
	
}
