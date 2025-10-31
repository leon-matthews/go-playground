package bookstore_test

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"

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
	result, err := bookstore.Buy(book)
	if err != nil {
		t.Errorf("got err %v", err)
	}
	got := result.Copies
	if want != got {
		t.Errorf("want %d, got %d", want, got)
	}
}

func TestBuyNoCopies(t *testing.T) {
	t.Parallel()
	b := bookstore.Book{
		Title:  "Spark Joy",
		Author: "Marie Kondo",
		Copies: 0,
	}
	_, err := bookstore.Buy(b)
	if err == nil {
		t.Error("want error buying from zero copies, got nil")
	}
}

func TestGetAllBooks(t *testing.T) {
	t.Parallel()
	catalogue := bookstore.Catalogue{
		1: {ID: 1, Title: "For the Love of Go"},
		2: {ID: 2, Title: "The Power of Go: Tools"},
	}
	want := []bookstore.Book{
		{ID: 1, Title: "For the Love of Go"},
		{ID: 2, Title: "The Power of Go: Tools"},
	}
	got := catalogue.GetAllBooks()
	sort.Slice(got, func(i, j int) bool {
		return got[i].ID < got[j].ID
	})
	assert.Equal(t, want, got)
}

func TestGetBook(t *testing.T) {
	t.Parallel()
	catalogue := bookstore.Catalogue{
		1: {ID: 1, Title: "For the Love of Go"},
		2: {ID: 2, Title: "The Power of Go: Tools"},
	}
	want := bookstore.Book{Title: "The Power of Go: Tools", ID: 2}
	got, err := catalogue.GetBook(2)
	if err != nil {
		t.Errorf("got err %v", err)
	}
	assert.Equal(t, want, got)
}

func TestGetBookMissing(t *testing.T) {
	t.Parallel()
	catalogue := bookstore.Catalogue{}
	_, err := catalogue.GetBook(999)
	if err == nil {
		t.Fatal("want error for non-existent ID, got nil")
	}
}

func TestNetPriceCents(t *testing.T) {
	t.Parallel()
	b := bookstore.Book{
		Title:           "For the Love of Go",
		PriceCents:      4000,
		DiscountPercent: 25,
	}
	want := 3000
	got := b.NetPriceCents()
	if want != got {
		t.Errorf("want %d, got %d", want, got)
	}
}

func TestSetCategory(t *testing.T) {
	b := bookstore.Book{
		Title: "For the Love of Go",
	}
	want := bookstore.CategoryAutobiography
	err := b.SetCategory(want)
	if err != nil {
		t.Fatal(err)
	}

	got := b.Category()
	if got != want {
		t.Errorf("want %v, got %v", want, got)
	}
}

func TestSetCategoryError(t *testing.T) {
	b := bookstore.Book{
		Title: "For the Love of Go",
	}
	err := b.SetCategory(bookstore.Category(-1))
	if err == nil {
		t.Fatal("want error for invalid category, got nil")
	}
}

func TestSetPriceCents(t *testing.T) {
	t.Parallel()
	b := bookstore.Book{
		Title:      "For the Love of Go",
		PriceCents: 4000,
	}
	want := 3000
	err := b.SetPriceCents(want)
	if err != nil {
		t.Fatal(err)
	}

	got := b.PriceCents
	if want != got {
		t.Errorf("want updated price %d, got %d", want, got)
	}
}

func TestSetPriceError(t *testing.T) {
	t.Parallel()
	b := bookstore.Book{
		Title:      "For the Love of Go",
		PriceCents: 4000,
	}
	err := b.SetPriceCents(-1)
	if err == nil {
		t.Fatal("want error setting invalid price -1, got nil")
	}
}
