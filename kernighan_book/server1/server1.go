// Minimal echo server
package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	// All requests start with "/"
	http.HandleFunc("/", handler)
	log.Fatal(http.ListenAndServe("localhost:8000", nil))
}

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "URL.Path = %q\n", r.URL.Path)
}
