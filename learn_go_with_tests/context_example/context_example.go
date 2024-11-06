package main

import (
	"fmt"
	"log"
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
		log.Printf("[%T] %+[1]v\n", ctx)

		fmt.Fprint(response, store.Fetch())
	}
}
