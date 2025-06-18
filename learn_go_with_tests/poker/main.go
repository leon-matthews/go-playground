package main

import (
	"log"
	"net/http"
)

type InMemoryStore struct{}

func (s *InMemoryStore) GetScore(name string) int {
	score := 123
	log.Printf("Score is always %d\n", score)
	return score
}

func (s *InMemoryStore) RecordWin(name string) {
	log.Println("Pretending to record Win")
}

func main() {
	store := &InMemoryStore{}
	server := &PlayerServer{store: store}
	address := ":8000"
	log.Printf("Starting server on %s\n", address)
	log.Fatal(http.ListenAndServe(address, server))
}
