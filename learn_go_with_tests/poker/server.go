package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
)

// PlayerStore
type PlayerStore interface {
	GetScore(name string) int
	RecordWin(name string)
}

// PlayerServer returns player's score
type PlayerServer struct {
	store PlayerStore
}

func (p *PlayerServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	player := strings.TrimPrefix(r.URL.Path, "/players/")

	switch r.Method {
	case http.MethodGet:
		p.getScore(w, player)
	case http.MethodPost:
		p.processWin(w, player)
	}
}

func (p *PlayerServer) getScore(w http.ResponseWriter, name string) {
	score := p.store.GetScore(name)
	if score == 0 {
		w.WriteHeader(http.StatusNotFound)
	}
	fmt.Fprint(w, score)
	log.Printf("GET score for %s is %d", name, score)
}

func (p *PlayerServer) processWin(w http.ResponseWriter, name string) {
	w.WriteHeader(http.StatusAccepted)
	p.store.RecordWin(name)
	log.Printf("POST score for %s incremented", name)
}
