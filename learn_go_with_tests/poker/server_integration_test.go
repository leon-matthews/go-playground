package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Integration test between PlayerServer and a PlayerStore
func TestRecordingWinsAndRetrievingThem(t *testing.T) {
	path, cleanup := createTempFile(t, "poker*.db")
	defer cleanup()

	cases := []struct {
		name  string
		store PlayerStore
	}{
		{
			name:  "in memory",
			store: NewPlayerStoreMemory(),
		},
		{
			name:  "boltdb",
			store: NewPlayerStoreBolt(path),
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {})
		server := PlayerServer{tt.store}
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
