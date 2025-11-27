package weather_test

import (
	"os"
	"testing"
	"weather"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseError(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile("testdata/error.json")
	require.NoError(t, err)

	want := weather.ErrorResponse{
		Code:       404,
		Message:    "Invalid date format",
		Parameters: []string{"date"},
	}

	got, err := weather.ParseResponse(data)
	require.NoError(t, err)
	assert.Equal(t, want, got)
}
