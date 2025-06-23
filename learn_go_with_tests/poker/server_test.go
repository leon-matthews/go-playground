package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// PlayerStoreMock records the calls to RecordWin for testing
type PlayerStoreMock struct {
	*PlayerStoreMemory
	winCalls []string
	league   []Player
}

func NewPlayerStoreMock() *PlayerStoreMock {
	return &PlayerStoreMock{
		NewPlayerStoreMemory(),
		make([]string, 0),
		make([]Player, 0),
	}
}

func (s *PlayerStoreMock) League() ([]Player, error) {
	return s.league, nil
}

func (s *PlayerStoreMock) RecordWin(name string) error {
	s.winCalls = append(s.winCalls, name)
	s.PlayerStoreMemory.RecordWin(name)
	return nil
}

func TestLeague(t *testing.T) {
	store := NewPlayerStoreMock()
	store.league = []Player{
		{"alyson", 30},
		{"blake", 44},
		{"leon", 12},
		{"stella", 27},
	}
	server := NewPlayerServer(store)

	t.Run("GET /league", func(t *testing.T) {
		request, _ := http.NewRequest(http.MethodGet, "/league", nil)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)
		assert.Equal(t, http.StatusOK, response.Code)
		assert.Equal(t, "application/json", response.Result().Header.Get("content-type"))

		var got []Player
		err := json.NewDecoder(response.Body).Decode(&got)
		if err != nil {
			t.Fatalf("Unable to parse JSON: %v", err)
		}

		// Same values, but sorted by number of wins
		want := []Player{
			{"blake", 44},
			{"alyson", 30},
			{"stella", 27},
			{"leon", 12},
		}
		assert.Equal(t, want, got)
	})
}

func TestScores(t *testing.T) {
	store := NewPlayerStoreMock()
	store.SetScore("alyson", 10)
	store.SetScore("leon", 20)
	server := NewPlayerServer(store)

	t.Run("Get Leon's score", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, "/players/leon", nil)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		assert.Equal(t, http.StatusOK, response.Code)
		assert.Equal(t, "20", response.Body.String())
	})

	t.Run("Get Alyson's score", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, "/players/alyson", nil)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		assert.Equal(t, http.StatusOK, response.Code)
		assert.Equal(t, "10", response.Body.String())
	})

	t.Run("return 404 if player not found", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, "/players/eric", nil)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		assert.Equal(t, http.StatusNotFound, response.Code)
	})

	t.Run("Update Eric's score", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodPost, "/players/eric", nil)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		assert.Equal(t, http.StatusAccepted, response.Code)
		assert.Equal(t, []string{"eric"}, store.winCalls)
	})
}
