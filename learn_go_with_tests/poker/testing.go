package poker

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// scheduledAlert is used by [AlerterMock] to check client code
type scheduledAlert struct {
	at     time.Duration
	amount int
}

func (s scheduledAlert) String() string {
	return fmt.Sprintf("%d chips at %v", s.amount, s.at)
}

// AlerterMock implements Alerter by just recorded calls to its Schedule() method
type AlerterMock struct {
	Alerts []scheduledAlert
}

// Schedule implements Alerter.Schedule by just recording the callers arguments
func (m *AlerterMock) Schedule(at time.Duration, amount int) {
	alert := scheduledAlert{at: at, amount: amount}
	m.Alerts = append(m.Alerts, alert)
}

// PlayerStoreMock records the calls to RecordWin for testing
type PlayerStoreMock struct {
	*PlayerStoreMemory
	winCalls []string
	league   League
}

// NewPlayerStoreMock initialises as new mock player store
func NewPlayerStoreMock() *PlayerStoreMock {
	return &PlayerStoreMock{
		NewPlayerStoreMemory(),
		make([]string, 0),
		make(League, 0),
	}
}

// League implements PlayerStore.League by simply returning contents of league property
func (s *PlayerStoreMock) League() (League, error) {
	return s.league, nil
}

// RecordWin increments the score for the given player
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

// CreateTempFile returns the path to a new, empty file.
// Call the returned cleanup function to delete the file.
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

// HTTPGetLeague fetches JSON data for player league table
func HTTPGetLeague(t *testing.T, server *PlayerServer) *httptest.ResponseRecorder {
	request, err := http.NewRequest(http.MethodGet, "/league", nil)
	require.NoError(t, err, "could not create request")
	response := httptest.NewRecorder()
	server.ServeHTTP(response, request)
	return response
}

// HTTPGetScore performs an HTTP GET to fetch the score for the given player
func HTTPGetScore(t *testing.T, server *PlayerServer, name string) *httptest.ResponseRecorder {
	request, err := http.NewRequest(http.MethodGet, fmt.Sprintf("/players/%s", name), nil)
	require.NoError(t, err, "could not create request")
	response := httptest.NewRecorder()
	server.ServeHTTP(response, request)
	return response
}

// HTTPPostWin performs an HTTP POST to record a win for the given player name
func HTTPPostWin(t *testing.T, server *PlayerServer, name string) *httptest.ResponseRecorder {
	t.Helper()
	request, err := http.NewRequest(http.MethodPost, fmt.Sprintf("/players/%s", name), nil)
	require.NoError(t, err, "could not create request")
	response := httptest.NewRecorder()
	server.ServeHTTP(response, request)
	return response
}
