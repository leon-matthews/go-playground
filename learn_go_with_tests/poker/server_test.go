package poker

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLeague(t *testing.T) {
	store := NewPlayerStoreMock()
	store.league = League{
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

		var got League
		err := json.NewDecoder(response.Body).Decode(&got)
		if err != nil {
			t.Fatalf("Unable to parse JSON: %v", err)
		}

		// Same values, but sorted by number of wins
		want := League{
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
