package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

type PlayerStoreMock struct {
	scores   map[string]int
	winCalls []string
}

func (s *PlayerStoreMock) GetScore(name string) (int, error) {
	return s.scores[name], nil
}

func (s *PlayerStoreMock) RecordWin(name string) error {
	s.winCalls = append(s.winCalls, name)
	return nil
}

func TestPlayerServer(t *testing.T) {
	store := &PlayerStoreMock{
		scores: map[string]int{
			"alyson": 10,
			"leon":   20,
		},
	}
	server := &PlayerServer{
		store: store,
	}

	t.Run("returns Leon's score", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, "/players/leon", nil)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		assert.Equal(t, http.StatusOK, response.Code)
		assert.Equal(t, "20", response.Body.String())
	})

	t.Run("returns Alyson's score", func(t *testing.T) {
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

	t.Run("POST Eric's score", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodPost, "/players/eric", nil)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		assert.Equal(t, http.StatusAccepted, response.Code)
		assert.Equal(t, []string{"eric"}, store.winCalls)
	})
}
