package poker

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Integration test between PlayerServer and implementations of PlayerStore
func TestRecordingWinsAndRetrievingThem(t *testing.T) {
	path, cleanup := CreateTempFile(t, "poker*.db")
	defer cleanup()
	boltdb, err := NewPlayerStoreBolt(path)
	require.NoError(t, err)

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
			response = HTTPGetScore(t, server, player)
			assert.Equal(t, http.StatusNotFound, response.Code)
			assert.Equal(t, "0", response.Body.String())

			// Record three wins
			HTTPPostWin(t, server, player)
			HTTPPostWin(t, server, player)
			HTTPPostWin(t, server, player)

			// Get score
			response = HTTPGetScore(t, server, player)
			assert.Equal(t, http.StatusOK, response.Code)
			assert.Equal(t, "3", response.Body.String())

			// Get league table
			response = HTTPGetLeague(t, server)
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
