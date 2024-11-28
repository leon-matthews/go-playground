package main

import (
	"log"
	"net/http"
)

func main() {
	server := &PlayerServer{NewInMemoryStorage()}
	log.Println("starting server on :8080")
	log.Fatal(http.ListenAndServe(":8080", server))
}
