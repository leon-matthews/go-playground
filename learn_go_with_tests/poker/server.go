package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

const jsonContentType = "application/json"

type Player struct {
	Name string
	Wins int
}

type PlayerServer struct {
	storage PlayerStorage
	http.Handler
}

func NewPlayerServer(storage PlayerStorage) *PlayerServer {
	s := new(PlayerServer)
	s.storage = storage

	router := http.NewServeMux()
	router.HandleFunc("/league", s.leagueHandler)
	router.HandleFunc("/players/", s.playersHandler)

	s.Handler = router
	return s
}

func (s *PlayerServer) leagueHandler(w http.ResponseWriter, r *http.Request) {
	// Pants first
	w.Header().Set("content-type", jsonContentType)
	w.WriteHeader(http.StatusOK)

	// Then shoes
	json.NewEncoder(w).Encode(s.storage.GetLeague())
}

func (s *PlayerServer) playersHandler(w http.ResponseWriter, r *http.Request) {
	player := strings.TrimPrefix(r.URL.Path, "/players/")
	switch r.Method {
	case http.MethodGet:
		s.showScore(w, player)
	case http.MethodPost:
		s.processWin(w, player)
	}
}

func (s *PlayerServer) showScore(w http.ResponseWriter, player string) {
	score := s.storage.GetPlayerScore(player)
	if score == 0 {
		w.WriteHeader(http.StatusNotFound)
	}
	fmt.Fprint(w, score)
}

func (s *PlayerServer) processWin(w http.ResponseWriter, player string) {
	s.storage.RecordWin(player)
	w.WriteHeader(http.StatusAccepted)
}
