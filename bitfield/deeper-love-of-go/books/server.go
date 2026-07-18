package books

import (
	"fmt"
	"net/http"
)

func ListenAndServe(url string, c *Catalogue) error {
	return http.ListenAndServe(url, http.HandlerFunc(hello))
}

func hello(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "[]")
}
