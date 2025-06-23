package main

import (
	"log"
	"net/http"

	"learn_go_with_tests/poker"
)

const filename = "poker.db"

func main() {
	store := poker.NewPlayerStoreBolt(filename)
	server := poker.NewPlayerServer(store)
	address := ":8000"
	log.Printf("Starting server on %s\n", address)
	log.Fatal(http.ListenAndServe(address, server))
}
