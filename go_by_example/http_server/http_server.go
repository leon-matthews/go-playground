package main

import (
	"fmt"
	"net/http"
)

// hello is a minimal HTTP handler.
// /
// A handler is an object implementing the http.Handler interface:
//
//	type Handler interface {
//	    ServeHTTP(ResponseWriter, *Request)
//	}
func hello(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "Hello, world!")
}

func headers(w http.ResponseWriter, req *http.Request) {
	for name, headers := range req.Header {
		for _, h := range headers {
			fmt.Fprintf(w, "%v: %v\n", name, h)
		}
	}
}

const port = "8080"

func main() {
	http.HandleFunc("/headers", headers)
	http.HandleFunc("/hello", hello)

	fmt.Printf("Running server on port %s\n", port)
	err := http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
	if err != nil {
		panic(err)
	}
}
