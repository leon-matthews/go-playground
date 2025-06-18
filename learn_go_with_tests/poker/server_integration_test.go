package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Integration test between PlayerServer and a PlayerStore
func TestRecordingWinsAndRetrievingThem(t *testing.T) {
	store := NewPlayerStoreMemory()
	server := PlayerServer{store}
	player := "leon"

	// Initial score should be zero
	var response *httptest.ResponseRecorder
	response = httptest.NewRecorder()
	server.ServeHTTP(response, newGetScoreRequest(player))
	assert.Equal(t, http.StatusNotFound, response.Code)
	assert.Equal(t, "0", response.Body.String())

	// Record three wins
	server.ServeHTTP(httptest.NewRecorder(), newPostWinRequest(t, player))
	server.ServeHTTP(httptest.NewRecorder(), newPostWinRequest(t, player))
	server.ServeHTTP(httptest.NewRecorder(), newPostWinRequest(t, player))

	// Fetch and check updated score
	response = httptest.NewRecorder()
	server.ServeHTTP(response, newGetScoreRequest(player))
	assert.Equal(t, http.StatusOK, response.Code)
	assert.Equal(t, "3", response.Body.String())
}

func newPostWinRequest(t *testing.T, name string) *http.Request {
	t.Helper()
	req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("/players/%s", name), nil)
	return req
}

func newGetScoreRequest(name string) *http.Request {
	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/players/%s", name), nil)
	return req
}
