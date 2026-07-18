package books_test

import (
	"encoding/json"
	"io"
	"net"
	"net/http"
	"path/filepath"
	"testing"

	"local.dev/books"
)

func TestServerListsAllBooks(t *testing.T) {
	// Start server in new goroutine
	t.Parallel()
	catalogue := getTestCatalogue()
	catalogue.Path = filepath.Join(t.TempDir(), "catalogue.json")
	addr := randomLocalAddr(t)
	t.Log(addr)
	go func() {
		err := books.ListenAndServe(addr, catalogue)
		if err != nil {
			// We can't call `t.Fatal()` from a goroutine
			panic(err)
		}
	}()

	// Make request from server
	resp, err := http.Get("http://" + addr)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	// Check response header
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status %d", resp.StatusCode)
	}

	// Check response body
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	bookList := []books.Book{}
	err = json.Unmarshal(data, &bookList)
	if err != nil {
		t.Fatalf("%v in %q", err, data)
	}
	assertTestBooks(t, bookList)
}

// randomLocalAddr attempts to find an unused local port.
// Requests random unused port from OS, creating local listener on port zero,
// then immediately closes it, returning that closed port number.
func randomLocalAddr(t *testing.T) string {
	t.Helper()
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	return l.Addr().String()
}
