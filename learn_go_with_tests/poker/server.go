// Package poker maintain scoring for poker players
package poker

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"slices"
	"strings"
)

// Player keeps track of number of wins for each named player
type Player struct {
	Name string
	Wins int
}

// A League is a collection of players
type League []Player

// Sort orders players by wins, highest score first
// If scores are equal, names are sorted alphabetically
func (l *League) Sort() {
	slices.SortFunc([]Player(*l), func(a, b Player) int {
		if a.Wins == b.Wins {
			return strings.Compare(a.Name, b.Name)
		}
		return b.Wins - a.Wins
	})
}

// A PlayerStore is used to keep players' scores.
type PlayerStore interface {
	// League fetches table of names and scores for all players
	League() (League, error)

	// RecordWin increments the score of the named player
	RecordWin(name string) error

	// Score fetches the current score for the named player
	Score(name string) (int, error)

	// SetScore increments the score of the named player
	SetScore(name string, score int) error
}

// PlayerServer returns player's score
type PlayerServer struct {
	store PlayerStore
	http.Handler
}

// NewPlayerServer creates a server with the given storage engine and sets up routes
func NewPlayerServer(store PlayerStore) *PlayerServer {
	p := &PlayerServer{store: store}
	router := http.NewServeMux()
	router.HandleFunc("/league", p.leagueHandler)
	router.HandleFunc("GET /players/{name}", p.getScore)
	router.HandleFunc("POST /players/{name}", p.processWin)
	p.Handler = router
	return p
}

func (p *PlayerServer) leagueHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("content-type", "application/json")
	league, err := p.store.League()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	league.Sort()
	json.NewEncoder(w).Encode(league)
}

func (p *PlayerServer) getScore(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	score, err := p.store.Score(name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
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
		return
	}
	log.Printf("POST score for %s incremented", name)
}
