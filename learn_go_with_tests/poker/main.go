package main

import (
	"log"
	"net/http"
)

func main() {
	store := NewPlayerStoreMemory()
	server := &PlayerServer{store: store}
	address := ":8000"
	log.Printf("Starting server on %s\n", address)
	log.Fatal(http.ListenAndServe(address, server))
}
