package main

import (
	"log"
	"net/http"

	"learn_go_with_tests/poker"
)

const (
	address  = ":8000"
	filename = "poker.db"
)

func main() {
	store, err := poker.NewPlayerStoreBolt(filename)
	if err != nil {
		log.Fatalf("Error %v", err)
	}
	server := poker.NewPlayerServer(store)
	log.Printf("Starting server on %s\n", address)
	log.Fatal(http.ListenAndServe(address, server))
}
