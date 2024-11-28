package main

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRecordAndRetrieveWins(t *testing.T) {
	store := NewInMemoryStorage()
	server := NewPlayerServer(store)
	player := "Leon"

	// On a winning streak!
	server.ServeHTTP(httptest.NewRecorder(), postWinRequest(player))
	server.ServeHTTP(httptest.NewRecorder(), postWinRequest(player))
	server.ServeHTTP(httptest.NewRecorder(), postWinRequest(player))

	// Check score
	response := httptest.NewRecorder()
	server.ServeHTTP(response, getScoreRequest(player))
	assert.Equal(t, http.StatusOK, response.Code)
	assert.Equal(t, "3", response.Body.String())
}
