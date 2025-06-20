package main

import (
	"fmt"
	"log"
	"net/http"
)

type PlayerStore interface {
	Score(name string) (int, error)
	SetScore(name string, score int) error
	RecordWin(name string) error
}

// PlayerServer returns player's score
type PlayerServer struct {
	store PlayerStore
	http.Handler
}

func NewPlayerServer(store PlayerStore) *PlayerServer {
	p := &PlayerServer{store: store}
	router := http.NewServeMux()
	router.HandleFunc("/league", p.leagueHandler)
	router.HandleFunc("GET /players/{name}", p.getScore)
	router.HandleFunc("POST /players/{name}", p.processWin)
	p.Handler = router
	return p
}

func (p *PlayerServer) leagueHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (p *PlayerServer) getScore(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	score, err := p.store.Score(name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	if score == 0 {
		w.WriteHeader(http.StatusNotFound)
	}
	fmt.Fprint(w, score)
	log.Printf("GET score for %s is %d", name, score)
}

func (p *PlayerServer) processWin(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	w.WriteHeader(http.StatusAccepted)
	err := p.store.RecordWin(name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	log.Printf("POST score for %s incremented", name)
}
