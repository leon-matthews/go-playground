package poker

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func NewStorageMock(league []Player) *StorageMock {
	return &StorageMock{
		NewInMemoryStorage(),
		make([]string, 0),
		league,
	}
}

// StorageMock provides in-memory player store
type StorageMock struct {
	*InMemoryStorage
	winCalls []string
	league   []Player
}

func (s *StorageMock) GetLeague() League {
	return s.league
}

func (s *StorageMock) RecordWin(name string) {
	s.InMemoryStorage.RecordWin(name)
	s.winCalls = append(s.winCalls, name)
}

func TestGetPlayerScore(t *testing.T) {
	storage := NewStorageMock(nil)
	storage.Reset(map[string]int{
		"alyson": 20,
		"leon":   10,
	})
	server := NewPlayerServer(storage)

	t.Run("get Alyson's score", func(t *testing.T) {
		request := getScoreRequest("alyson")
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		assertStatus(t, http.StatusOK, response.Code)
		assertResponseBody(t, "20", response.Body.String())
	})

	t.Run("get Leon's score", func(t *testing.T) {
		request := getScoreRequest("leon")
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		assertStatus(t, http.StatusOK, response.Code)
		assertResponseBody(t, "10", response.Body.String())
	})

	t.Run("return 404 if player not found", func(t *testing.T) {
		request := getScoreRequest("bogus")
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		assertStatus(t, http.StatusNotFound, response.Code)
	})
}

func TestLeague(t *testing.T) {
	league := []Player{
		{"Alyson", 20},
		{"Leon", 10},
	}
	storage := NewStorageMock(league)
	server := NewPlayerServer(storage)

	t.Run("return JSON data from league", func(t *testing.T) {
		request := getLeagueRequest()
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		assertStatus(t, http.StatusOK, response.Code)
		if response.Result().Header.Get("Content-Type") != "application/json" {
			t.Errorf("response did not have content-type of application/json, got %v", response.Result().Header)
		}

		got := getLeagueFromResponse(t, response.Body)
		assert.Equal(t, league, got)
	})
}

func TestPostPlayerScore(t *testing.T) {
	storage := NewStorageMock(nil)
	storage.Reset(map[string]int{
		"alyson": 20,
		"leon":   10,
	})
	server := NewPlayerServer(storage)

	t.Run("ensure wins are recorded in storage", func(t *testing.T) {
		player := "eric"
		request := postWinRequest(player)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		assertStatus(t, http.StatusAccepted, response.Code)
		assert.Equal(t, 1, len(storage.winCalls), "expected 1 call to win, but got %d", len(storage.winCalls))
		assert.Equal(t, player, storage.winCalls[0])
	})
}

func assertContentType(t testing.TB, response *httptest.ResponseRecorder, want string) {
	t.Helper()
	assert.Equal(t, want, response.Header().Get("Content-Type"))
}

func assertResponseBody(t testing.TB, want, got string) {
	t.Helper()
	if got != want {
		t.Errorf("response body is wrong, got %q want %q", got, want)
	}
}

func assertStatus(t testing.TB, want, got int) {
	t.Helper()
	if got != want {
		t.Errorf("did not get correct status, got %d, want %d", got, want)
	}
}

func getLeagueFromResponse(t testing.TB, body io.Reader) []Player {
	var league []Player
	err := json.NewDecoder(body).Decode(&league)
	if err != nil {
		t.Fatalf("Unable to parse response from server %q into slice of Player, '%v'", body, err)
	}
	return league
}

func getLeagueRequest() *http.Request {
	request, _ := http.NewRequest(http.MethodGet, "/league", nil)
	return request
}

func getScoreRequest(name string) *http.Request {
	request, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/players/%s", name), nil)
	return request
}

func postWinRequest(name string) *http.Request {
	request, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("/players/%s", name), nil)
	return request
}
