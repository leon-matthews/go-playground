package books_test

import (
	"cmp"
	"path/filepath"
	"slices"
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
	got := catalogue.AllBooks()
	assertTestBooks(t, got)
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

func TestNewCatalog_CreatesEmptyCatalog(t *testing.T) {
	t.Parallel()
	catalog := books.NewCatalogue()
	books := catalog.AllBooks()
	if len(books) > 0 {
		t.Errorf("want empty catalog, got %#v", books)
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

func TestSetCopies_IsRaceFree(t *testing.T) {
	t.Parallel()
	catalog := getTestCatalogue()

	// Set copies for "abc" in new goroutine
	go func() {
		for range 100 {
			_, err := catalog.SetCopies("abc", 0)
			if err != nil {
				panic(err)
			}
		}
	}()

	// Get copies for the same ID at the same time
	for range 100 {
		_, err := catalog.Copies("abc")
		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestSyncWritesCatalogDataToFile(t *testing.T) {
	t.Parallel()
	catalog := getTestCatalogue()
	catalog.Path = filepath.Join(t.TempDir(), "/catalogue.json")
	err := catalog.Sync()
	if err != nil {
		t.Fatal(err)
	}
	newCatalog, err := books.OpenCatalogue(catalog.Path)
	if err != nil {
		t.Fatal(err)
	}
	bookList := newCatalog.AllBooks()
	assertTestBooks(t, bookList)
}


func assertTestBooks(t *testing.T, got []books.Book) {
	t.Helper()
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
	slices.SortFunc(got, func(a, b books.Book) int {
		return cmp.Compare(a.Author, b.Author)
	})
	t.Logf("%#v", want)
	t.Logf("%#v", got)
	if !slices.Equal(want, got) {
		t.Fatalf("want %#v, got %#v", want, got)
	}
}

// getTestCatalogue creates a shiny new catalogue for our tests
func getTestCatalogue() *books.Catalogue {
	c := books.NewCatalogue()
	err := c.AddBook(books.Book{
		Title:  "In the Company of Cheerful Ladies",
		Author: "Alexander McCall Smith",
		Copies: 1,
		ID:     "abc",
	})
	if err != nil {
		panic(err)
	}
	err = c.AddBook(books.Book{
		Title:  "White Heat",
		Author: "Dominic Sandbrook",
		Copies: 2,
		ID:     "xyz",
	})
	if err != nil {
		panic(err)
	}
	return c
}
