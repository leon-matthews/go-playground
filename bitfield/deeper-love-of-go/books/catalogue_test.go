package books_test

import (
	"cmp"
	"maps"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"local.dev/books"
)

func TestAddBook_ReturnsErrorIfIDExists(t *testing.T) {
	t.Parallel()
	catalogue := getTestCatalogue()
	_, ok := catalogue.GetBook("abc")
	if !ok {
		t.Fatal("book not present")
	}

	err := catalogue.AddBook(books.Book{
		ID:     "abc",
		Title:  "In the Company of Cheerful Ladies",
		Author: "Alexander McCall Smith",
		Copies: 1,
	})

	if err == nil {
		t.Fatal("want error for duplicate ID, got nil")
	}
}

func TestGetAllBooks_ReturnsAllBooks(t *testing.T) {
	t.Parallel()
	catalogue := getTestCatalogue()
	want := []books.Book{
		{
			ID:     "abc",
			Title:  "In the Company of Cheerful Ladies",
			Author: "Alexander McCall Smith",
			Copies: 1,
		},
		{
			ID:     "def",
			Title:  "White Heat",
			Author: "Dominic Sandbrook",
			Copies: 2,
		},
	}
	got := catalogue.AllBooks()
	slices.SortFunc(got, func(a, b books.Book) int {
		return strings.Compare(a.Author, b.Author)
	})
	if !slices.Equal(want, got) {
		t.Fatalf("want %#v, got %#v", want, got)
	}
}

func TestGetBook_FindsBookInCatalogByID(t *testing.T) {
	t.Parallel()
	catalogue := getTestCatalogue()
	want := books.Book{
		ID:     "abc",
		Title:  "In the Company of Cheerful Ladies",
		Author: "Alexander McCall Smith",
		Copies: 1,
	}
	got, found := catalogue.GetBook("abc")
	if !found {
		t.Fatal("book not found")
	}
	if want != got {
		t.Fatalf("want %#v, got %#v", want, got)
	}
}

func TestGetBook_ReturnsFalseWhenBookNotFound(t *testing.T) {
	t.Parallel()
	catalogue := getTestCatalogue()
	_, ok := catalogue.GetBook("no such ID")
	if ok {
		t.Fatal("want false for nonexistent ID, got true")
	}
}

func TestAddBook_AddsGivenBookToCatalog(t *testing.T) {
	// Pre-condition
	t.Parallel()
	catalogue := getTestCatalogue()
	_, ok := catalogue.GetBook("123")
	if ok {
		t.Fatal("book already present")
	}

	err := catalogue.AddBook(books.Book{
		ID:     "123",
		Title:  "The Prize of all the Oceans",
		Author: "Glyn Williams",
		Copies: 2,
	})
	if err != nil {
		t.Fatal(err)
	}

	// Post-condition
	_, ok = catalogue.GetBook("123")
	if !ok {
		t.Fatal("added book not found")
	}
}

func TestOpenCatalog_LoadsCatalogDataFromFile(t *testing.T) {
	t.Parallel()
	catalogue, err := books.OpenCatalogue("testdata/catalogue.json")
	if err != nil {
		t.Fatal(err)
	}
	want := []books.Book{
		{
			Title:  "In the Company of Cheerful Ladies",
			Author: "Alexander McCall Smith",
			Copies: 1,
			ID:     "abc",
		},
		{
			Title:  "White Heat",
			Author: "Dominic Sandbrook",
			Copies: 2,
			ID:     "xyz",
		},
	}
	got := catalogue.AllBooks()
	slices.SortFunc(got, func(a, b books.Book) int {
		return cmp.Compare(a.Author, b.Author)
	})
	if !slices.Equal(want, got) {
		t.Fatalf("want %#v, got %#v", want, got)
	}
}

func TestSetCopies_OnCatalogModifiesSpecifiedBook(t *testing.T) {
	t.Parallel()
	catalogue := getTestCatalogue()
	book, ok := catalogue.GetBook("abc")
	if !ok {
		t.Fatal("book not found")
	}
	if book.Copies != 1 {
		t.Fatalf("want 1 copy before change, got %d", book.Copies)
	}
	book, err := catalogue.SetCopies("abc", 2)
	if err != nil {
		t.Fatal(err)
	}
	if book.Copies != 2 {
		t.Fatalf("want 2 copies after change, got %d", book.Copies)
	}
}

func TestSync_UpdatesJSONFile(t *testing.T) {
	// Update copies for one book
	t.Parallel()
	catalogue := getTestCatalogue()
	want := 42
	book, err := catalogue.SetCopies("abc", 42)
	if err != nil {
		t.Fatal(err)
	}
	if book.Copies != want {
		t.Errorf("want %d, got %d", want, book.Copies)
	}

	// Save changes to file
	path := filepath.Join(t.TempDir(), "/catalogue.json")
	err = catalogue.Sync(path)
	if err != nil {
		t.Fatal(err)
	}

	// Read back and compare
	updated, err := books.OpenCatalogue(path)
	if err != nil {
		t.Fatal(err)
	}
	if !maps.Equal(catalogue, updated) {
		t.Errorf("want %+v, got %+v", catalogue, updated)
	}
}

// getTestCatalogue creates a shiny new catalogue for our tests
func getTestCatalogue() books.Catalogue {
	return map[string]books.Book{
		"abc": {
			Title:  "In the Company of Cheerful Ladies",
			Author: "Alexander McCall Smith",
			Copies: 1,
			ID:     "abc",
		},
		"def": {
			Title:  "White Heat",
			Author: "Dominic Sandbrook",
			Copies: 2,
			ID:     "def",
		},
	}
}
