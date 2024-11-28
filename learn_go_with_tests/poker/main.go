package main

import (
	"log"
	"net/http"
)

func main() {
	storage := NewInMemoryStorage()
	server := NewPlayerServer(storage)
	log.Println("starting server on :8080")
	log.Fatal(http.ListenAndServe(":8080", server))
}
