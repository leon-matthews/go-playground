package poker

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func (s *PlayerStoreMock) League() (League, error) {
	return s.league, nil
}

func (s *PlayerStoreMock) RecordWin(name string) error {
	s.winCalls = append(s.winCalls, name)
	s.PlayerStoreMemory.RecordWin(name)
	return nil
}

// AssertPlayerWin checks that RecordWin() was called exactly once and with given name
func AssertPlayerWin(t *testing.T, store *PlayerStoreMock, name string) {
	t.Helper()
	if len(store.winCalls) != 1 {
		t.Fatal("expected a win call but didn't get any")
	}
	assert.Equal(t, name, store.winCalls[0])
}

// Create empty file and return its path
// Defer a call the returned cleanup function to delete file
func CreateTempFile(t *testing.T, pattern string) (string, func()) {
	t.Helper()
	f, err := os.CreateTemp("", pattern)
	if err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	cleanup := func() {
		os.Remove(f.Name())
	}
	return f.Name(), cleanup
}

// HttpGetLeague fetches JSON data for player league table
func HttpGetLeague(t *testing.T, server *PlayerServer) *httptest.ResponseRecorder {
	request, err := http.NewRequest(http.MethodGet, "/league", nil)
	require.NoError(t, err, "could not create request")
	response := httptest.NewRecorder()
	server.ServeHTTP(response, request)
	return response
}

// HttpGetScore performs an HTTP GET to fetch the score for the given player
func HttpGetScore(t *testing.T, server *PlayerServer, name string) *httptest.ResponseRecorder {
	request, err := http.NewRequest(http.MethodGet, fmt.Sprintf("/players/%s", name), nil)
	require.NoError(t, err, "could not create request")
	response := httptest.NewRecorder()
	server.ServeHTTP(response, request)
	return response
}

// HttpPostWin performs an HTTP POST to record a win for the given player name
func HttpPostWin(t *testing.T, server *PlayerServer, name string) *httptest.ResponseRecorder {
	t.Helper()
	request, err := http.NewRequest(http.MethodPost, fmt.Sprintf("/players/%s", name), nil)
	require.NoError(t, err, "could not create request")
	response := httptest.NewRecorder()
	server.ServeHTTP(response, request)
	return response
}
