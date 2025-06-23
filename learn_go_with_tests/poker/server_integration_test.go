package poker

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Integration test between PlayerServer and implementations of PlayerStore
func TestRecordingWinsAndRetrievingThem(t *testing.T) {
	path, cleanup := createTempFile(t, "poker*.db")
	defer cleanup()
	boltdb := NewPlayerStoreBolt(path)

	testcases := []struct {
		name  string
		store PlayerStore
	}{
		{
			name:  "in memory",
			store: NewPlayerStoreMemory(),
		},
		{
			name:  "boltdb",
			store: boltdb,
		},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			var response *httptest.ResponseRecorder
			server := NewPlayerServer(tt.store)
			player := "leon"

			// Initial score should be zero
			response = httpGetScore(t, server, player)
			assert.Equal(t, http.StatusNotFound, response.Code)
			assert.Equal(t, "0", response.Body.String())

			// Record three wins
			httpPostWin(t, server, player)
			httpPostWin(t, server, player)
			httpPostWin(t, server, player)

			// Get score
			response = httpGetScore(t, server, player)
			assert.Equal(t, http.StatusOK, response.Code)
			assert.Equal(t, "3", response.Body.String())

			// Get league table
			response = httpGetLeague(t, server)
			assert.Equal(t, http.StatusOK, response.Code)
			assert.Equal(t, "application/json", response.Result().Header.Get("content-type"))
			var got []Player
			err := json.NewDecoder(response.Body).Decode(&got)
			require.NoError(t, err, "Could not decode JSON")
			want := []Player{
				{"leon", 3},
			}
			assert.Equal(t, want, got)
		})
	}
}

// httpGetScore performs an HTTP GET to fetch the score for the given player
func httpGetScore(t *testing.T, server *PlayerServer, name string) *httptest.ResponseRecorder {
	request, err := http.NewRequest(http.MethodGet, fmt.Sprintf("/players/%s", name), nil)
	require.NoError(t, err, "could not create request")
	response := httptest.NewRecorder()
	server.ServeHTTP(response, request)
	return response
}

// httpPostWin performs an HTTP POST to record a win for the given player name
func httpPostWin(t *testing.T, server *PlayerServer, name string) *httptest.ResponseRecorder {
	t.Helper()
	request, err := http.NewRequest(http.MethodPost, fmt.Sprintf("/players/%s", name), nil)
	require.NoError(t, err, "could not create request")
	response := httptest.NewRecorder()
	server.ServeHTTP(response, request)
	return response
}

// httpGetLeague fetches JSON data for player league table
func httpGetLeague(t *testing.T, server *PlayerServer) *httptest.ResponseRecorder {
	request, err := http.NewRequest(http.MethodGet, "/league", nil)
	require.NoError(t, err, "could not create request")
	response := httptest.NewRecorder()
	server.ServeHTTP(response, request)
	return response
}

// Create empty file and return its path
// Defer a call the returned cleanup function to delete file
func createTempFile(t *testing.T, pattern string) (string, func()) {
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
