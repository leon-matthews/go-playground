package main

import (
	"fmt"
	"net/http"
)

func main() {
	fmt.Println("context example")
}

// Store holds content
type Store interface {
	Cancel()
	Fetch() string
}

// Server fetches content from store and send to client
func Server(store Store) http.HandlerFunc {
	return func(response http.ResponseWriter, request *http.Request) {
		ctx := request.Context()
		data := make(chan string, 1)

		go func() {
			data <- store.Fetch()
		}()

		select {
		case d := <-data:
			fmt.Fprintf(response, d)
		case <-ctx.Done():
			store.Cancel()
		}
	}
}
