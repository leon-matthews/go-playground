package mediainfo

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtractInfo(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		raw, err := os.ReadFile("testdata/cow.json")
		require.NoError(t, err)

		got, err := extractInfo("cow.mp4", raw)
		require.NoError(t, err)

		expected := &Media{
			Name:          "cow.mp4",
			Size:          483210,
			Format:        "MPEG-4",
			Bitrate:       962570,
			Duration:      time.Duration(4.016 * float64(time.Second)),
			Height:        848,
			Width:         480,
			AudioBitrate:  132300,
			AudioChannels: 2,
			AudioFormat:   "AAC",
			VideoBitrate:  819440,
			VideoFormat:   "HEVC",
		}
		assert.Equal(t, expected, got)
	})

	t.Run("empty", func(t *testing.T) {
		raw, err := os.ReadFile("testdata/empty.json")
		require.NoError(t, err)

		_, err = extractInfo("", raw)
		assert.ErrorContains(t, err, "failed to unmarshal JSON: unexpected end of JSON input")
	})

	t.Run("invalid", func(t *testing.T) {
		raw, err := os.ReadFile("testdata/invalid.json")
		require.NoError(t, err)

		info, err := extractInfo("invalid.file", raw)
		fmt.Printf("[%T]%+[1]v\n", info)
		assert.ErrorContains(t, err, `expected one video track in "invalid.file", found 0`)
	})
}

func TestExtractVersion(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		have := []byte("mediaInfo Command line,\nMediaInfoLib - v24.12")
		want := "v24.12"
		got, err := extractVersion(have)
		assert.NoError(t, err)
		assert.Equal(t, want, got)
	})

	t.Run("empty", func(t *testing.T) {
		have := []byte("")
		_, err := extractVersion(have)
		assert.ErrorContains(t, err, "version not found: \"\"")
	})

	t.Run("invalid", func(t *testing.T) {
		have := []byte("Banana Splits")
		_, err := extractVersion(have)
		assert.ErrorContains(t, err, "version not found: \"Banana Splits\"")
	})
}
