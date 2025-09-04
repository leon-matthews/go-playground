package mediainfo

import (
	"encoding/json"
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

func TestJsonInt(t *testing.T) {
	t.Run("marshal", func(t *testing.T) {
		i := jsonInt(10)
		b, err := json.Marshal(i)
		require.NoError(t, err)
		assert.Equal(t, []byte(`10`), b)
	})

	t.Run("unmarshal", func(t *testing.T) {
		var i jsonInt
		err := json.Unmarshal([]byte(`10`), &i)
		require.NoError(t, err)
		assert.Equal(t, jsonInt(10), i)
	})

	t.Run("unmarshal quoted", func(t *testing.T) {
		var i jsonInt
		err := json.Unmarshal([]byte(`"10"`), &i)
		require.NoError(t, err)
		assert.Equal(t, jsonInt(10), i)
	})

	t.Run("unmarshal error", func(t *testing.T) {
		var i jsonInt
		err := json.Unmarshal([]byte(`""`), &i)
		assert.ErrorContains(t, err, `strconv.Atoi: parsing "": invalid syntax`)
	})
}

func TestJsonFloat(t *testing.T) {
	t.Run("marshal", func(t *testing.T) {
		f := jsonFloat(1.2)
		b, err := json.Marshal(f)
		require.NoError(t, err)
		fmt.Printf("[%T]%+[1]v\n", string(b))
		assert.Equal(t, []byte(`1.2`), b)
	})

	t.Run("unmarshal", func(t *testing.T) {
		var f jsonFloat
		err := json.Unmarshal([]byte(`1.2`), &f)
		require.NoError(t, err)
		assert.Equal(t, jsonFloat(1.2), f)
	})

	t.Run("unmarshal quoted", func(t *testing.T) {
		var f jsonFloat
		err := json.Unmarshal([]byte(`"1.2"`), &f)
		require.NoError(t, err)
		assert.Equal(t, jsonFloat(1.2), f)
	})

	t.Run("unmarshal error", func(t *testing.T) {
		var f jsonFloat
		err := json.Unmarshal([]byte(`""`), &f)
		assert.ErrorContains(t, err, `strconv.ParseFloat: parsing "": invalid syntax`)
	})
}
