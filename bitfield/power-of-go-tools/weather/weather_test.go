package weather_test

import (
	"os"
	"path/filepath"
	"testing"
	"weather"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestErrorResponse_BuildsCustomMessage(t *testing.T) {
	t.Parallel()
	data := loadData(t, "error.json")
	err := weather.NewResponseError(400, data)
	want := "API error 400: Invalid date format (in sunrise, sunset)"
	assert.ErrorContains(t, err, want)
}

func TestErrorResponse_UseGenericMessage(t *testing.T) {
	err := weather.NewResponseError(402, []byte("give me money, sucker"))
	want := "API unknown error 402: give me money, sucker"
	assert.ErrorContains(t, err, want)
}

// loadData returns bytes from file under testdata folder
func loadData(t *testing.T, name string) []byte {
	t.Helper()
	path := filepath.Join("testdata", name)
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	return data
}
