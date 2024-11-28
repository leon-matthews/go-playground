package main

import (
	"log"
	"net/http"
)

type InMemoryStorage struct{}

func (s InMemoryStorage) GetPlayerScore(name string) int {
	return 123
}

func (s InMemoryStorage) RecordWin(name string) {}

func main() {
	server := &PlayerServer{InMemoryStorage{}}
	log.Println("starting server on :8080")
	log.Fatal(http.ListenAndServe(":8080", server))
}
