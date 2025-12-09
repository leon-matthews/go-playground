package main

import (
    _ "expvar"
	"fmt"
	"maps"
	"net/http"
	"slices"
)

const port = "8080"

func main() {
	// Simple function must first be cast to a type that implements [Handler]...
	helloHandler := http.HandlerFunc(hello)
	http.Handle("/hello", helloHandler)

	// ...which can be achieved via a helper
	http.HandleFunc("/headers", headers)

	// Importing "expvar" creates the endpoint: /debug/vars
	// Can monitor memory use live with, eg. 'expvarmon'

    // Run server!
	fmt.Printf("Running server on port %s\n", port)
	err := http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
	if err != nil {
		panic(err)
	}
}

// Print hello directly to writer
func hello(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "Hello, world!")
}

// headers prints request's headers out in alphabetical order
func headers(w http.ResponseWriter, req *http.Request) {
	keys := slices.Collect(maps.Keys(req.Header))
	slices.Sort(keys)
	for _, key := range keys {
		value := req.Header.Get(key)
		fmt.Fprintf(w, "%v: %v\n", key, value)
	}
}
