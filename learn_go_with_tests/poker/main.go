package main

import (
	"log"
	"net/http"
)

const filename = "poker.db"

func main() {
	store := NewPlayerStoreBolt(filename)
	server := NewPlayerServer(store)
	address := ":8000"
	log.Printf("Starting server on %s\n", address)
	log.Fatal(http.ListenAndServe(address, server))
}
