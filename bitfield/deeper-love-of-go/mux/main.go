package main

import (
	"fmt"
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /hello", hello)
	mux.HandleFunc("GET /goodbye", goodbye)
	http.ListenAndServe("localhost:8000", mux)
}

func hello(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello, world!")
}

func goodbye(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Goodbye cruel world!")
}
