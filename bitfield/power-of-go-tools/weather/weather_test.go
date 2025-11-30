package weather

import (
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildURL(t *testing.T) {
	client := NewClient("swordfish")

	u, err := url.Parse("https://api.example.com/v2")
	require.NoError(t, err)
	client.baseURL = u

	cases := []struct {
		name   string
		path   string
		params map[string]string
		want   string
	}{
		{
			name:   "no params",
			path:   "/weather/current.json",
			params: nil,
			want:   "https://api.example.com/v2/weather/current.json?appid=swordfish&units=metric",
		},
		{
			name:   "add params, remove leading slash",
			path:   "weather/current.json",
			params: map[string]string{"lat": "12.34", "lon": "56.78"},
			want:   "https://api.example.com/v2/weather/current.json?appid=swordfish&lat=12.34&lon=56.78&units=metric",
		},
		{
			name:   "directory traversal",
			path:   "../v3/weather/current.json",
			params: nil,
			want:   "https://api.example.com/v3/weather/current.json?appid=swordfish&units=metric",
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			u, err := client.buildURL(tt.path, tt.params)
			require.NoError(t, err)
			assert.Equal(t, tt.want, u)
		})
	}
}

func TestErrorResponse_BuildsCustomMessage(t *testing.T) {
	t.Parallel()
	data := loadData(t, "error.json")
	err := NewResponseError(400, data)
	want := "API error 400: Invalid date format (in sunrise, sunset)"
	assert.ErrorContains(t, err, want)
}

func TestErrorResponse_UseGenericMessage(t *testing.T) {
	err := NewResponseError(402, []byte("give me money, sucker"))
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
