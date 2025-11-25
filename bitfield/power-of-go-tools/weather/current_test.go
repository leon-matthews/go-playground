package weather_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"weather"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHTTPGet_SuccessfullyGetsFromLocalServer(t *testing.T) {
	t.Parallel()
	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "testdata/current.json")
	}))
	defer ts.Close()

	fmt.Printf("[%T]%+[1]v\n", ts.URL)
	client := ts.Client()
	resp, err := client.Get(ts.URL)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

}

func TestParseResponse_CorrectlyParsesJSONData(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile("testdata/current.json")
	if err != nil {
		t.Fatal(err)
	}
	want := weather.Current{
		Summary:     "Clouds",
		Temperature: 21.74,
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
