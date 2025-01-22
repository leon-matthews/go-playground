package main

import (
	"fmt"
	"net/http"
	"time"
)

func main() {
	http.HandleFunc("/hello", hello)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		err := fmt.Errorf("starting server: %w", err)
		fmt.Println(err)
	}
}

// hello is run once per incoming request
func hello(w http.ResponseWriter, req *http.Request) {
	fmt.Println("server: hello handler started")
	defer fmt.Println("server: hello handler ended")

	ctx := req.Context()
	select {
	case <-time.After(time.Second * 10):
		fmt.Fprintf(w, "Hello World")
	case <-ctx.Done():
		err := ctx.Err()
		fmt.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
