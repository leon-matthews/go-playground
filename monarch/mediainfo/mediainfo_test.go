package mediainfo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractVersion(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		have := "MediaInfo Command line,\nMediaInfoLib - v24.12"
		want := "v24.12"
		got, err := extractVersion(have)
		assert.NoError(t, err)
		assert.Equal(t, want, got)
	})

	t.Run("empty", func(t *testing.T) {
		have := ""
		_, err := extractVersion(have)
		assert.ErrorContains(t, err, "version not found: \"\"")
	})

	t.Run("nonsense", func(t *testing.T) {
		have := "Banana Splits"
		_, err := extractVersion(have)
		assert.ErrorContains(t, err, "version not found: \"Banana Splits\"")
	})
}
