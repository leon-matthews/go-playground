package forgotten

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigFromYAML_CorrectlyParsesYAMLData(t *testing.T) {
	t.Parallel()
	want := Config{
		Global: GlobalConfig{
			ScrapeInterval:     15 * time.Second,
			EvaluationInterval: 30 * time.Second,
			ScrapeTimeout:      10 * time.Second,
			ExternalLabels: map[string]string{
				"monitor": "codelab",
				"foo":     "bar",
			},
		},
	}
	got, err := ConfigFromYAML("testdata/config.yaml")
	require.NoError(t, err)
	assert.Equal(t, want, got)
}
