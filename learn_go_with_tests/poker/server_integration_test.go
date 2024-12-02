package poker

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRecordAndRetrieveWins(t *testing.T) {
	database, cleanDatabase := createTempFile(t, "[]")
	defer cleanDatabase()
	store, err := NewFileSystemStorage(database)
	assert.NoError(t, err)
	server := NewPlayerServer(store)
	player := "Leon"

	// On a winning streak!
	server.ServeHTTP(httptest.NewRecorder(), postWinRequest(player))
	server.ServeHTTP(httptest.NewRecorder(), postWinRequest(player))
	server.ServeHTTP(httptest.NewRecorder(), postWinRequest(player))

	t.Run("get score", func(t *testing.T) {
		response := httptest.NewRecorder()

		server.ServeHTTP(response, getScoreRequest(player))

		assert.Equal(t, http.StatusOK, response.Code)
		assert.Equal(t, "3", response.Body.String())
	})

	t.Run("get score not found", func(t *testing.T) {
		response := httptest.NewRecorder()

		server.ServeHTTP(response, getScoreRequest("bogus"))

		assert.Equal(t, http.StatusNotFound, response.Code)
		assert.Equal(t, "0", response.Body.String())
	})

	t.Run("get league", func(t *testing.T) {
		response := httptest.NewRecorder()

		server.ServeHTTP(response, getLeagueRequest())

		assert.Equal(t, http.StatusOK, response.Code)
		got := getLeagueFromResponse(t, response.Body)
		want := []Player{
			{"Leon", 3},
		}
		assert.Equal(t, want, got)
	})
}
