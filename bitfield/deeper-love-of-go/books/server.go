package books

import (
	"encoding/json"
	"net/http"
)

// ListenAndServe starts HTTP server
func ListenAndServe(url string, c *Catalogue) error {
	return http.ListenAndServe(url, listAllBooks(c))
}

// listAllBooks creates a closure around a handler, allowing access to the catalogue
func listAllBooks(c *Catalogue) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		b := c.AllBooks()
		err := json.NewEncoder(w).Encode(b)
		if err != nil {
			panic(err)
		}
	}
}
