package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
)

func main() {
	fmt.Println("context example")
}

// Store holds content
type Store interface {
	Fetch(ctx context.Context) (string, error)
}

// Server fetches content from store and send to client
func Server(store Store) http.HandlerFunc {
	return func(response http.ResponseWriter, request *http.Request) {
		data, err := store.Fetch(request.Context())
		if err != nil {
			log.Println("error fetching string from store:", err)
			return
		}
		fmt.Fprint(response, data)
	}
}
