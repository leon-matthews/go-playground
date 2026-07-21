package main

import (
	"fmt"
	"net/http"
	"strings"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /hello/{name...}", hello)
	mux.HandleFunc("GET /goodbye/{name...}", goodbye)
	http.ListenAndServe("localhost:8000", mux)
}

func hello(w http.ResponseWriter, r *http.Request) {
	name := getName(r)
	fmt.Fprintf(w, "Hello, %s!\n", name)
}

func goodbye(w http.ResponseWriter, r *http.Request) {
	name := getName(r)
	fmt.Fprintf(w, "Goodbye %s!\n", name)
}

// getName extracts a title-cased name from the URL's path.
// It requires that the wildcard `{name...}` be used in the [http.ServeMux]
// For example:
// http://localhost:8000/goodbye/leon/matthews/ => "Goodbye Leon Matthews!"
func getName(r *http.Request) string {
	name := r.PathValue("name")
	if name == "" {
		name = "world"
	}
	name = strings.Replace(name, "/", " ", -1)
	name = strings.TrimSpace(name)
	name = strings.Title(name)
	return name
}
