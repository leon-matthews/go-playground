package main

import (
	"fmt"
	"net/http"
	"time"
)

const port = "8080"

// A Context carries deadlines, cancellation signals, and other request-scoped
// values across API boundaries and goroutines.
//
// Visit http://localhost:8080/hello, then hit stop to see handler get cancelled
func main() {
	http.HandleFunc("/hello", hello)
	fmt.Printf("Running server on port %s\n", port)
	err := http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
	if err != nil {
		panic(err)
	}
}

// hello is run once per incoming request
func hello(w http.ResponseWriter, req *http.Request) {
	fmt.Println("server: hello handler started")
	defer fmt.Println("server: hello handler ended")

	// Every request has a context created for it
	ctx := req.Context()

	select {
	case <-time.After(5 * time.Second):
		fmt.Fprintf(w, "Hello World")
	case <-ctx.Done():
		err := ctx.Err()
		fmt.Println("server:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
