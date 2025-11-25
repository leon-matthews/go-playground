package weather_test

import (
	"os"
	"testing"
	"weather"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseResponse_CorrectlyParsesJSONData(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile("testdata/current.json")
	if err != nil {
		t.Fatal(err)
	}
	want := weather.Current{
		Summary: "Clouds",
	}
	got, err := weather.ParseResponse(data)
	require.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestParseResponse_ReturnsErrorGivenEmptyData(t *testing.T) {
	t.Parallel()
	_, err := weather.ParseResponse([]byte{})
	assert.Error(t, err, "want error parsing empty response, got nil")
}

func TestParseResponse_ReturnsErrorGivenInvalidJSON(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile("testdata/current_invalid.json")
	if err != nil {
		t.Fatal(err)
	}
	_, err = weather.ParseResponse(data)
	assert.ErrorContains(t, err, "want at least one Weather element")
}
