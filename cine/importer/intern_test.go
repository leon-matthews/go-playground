package importer

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInterner(t *testing.T) {
	t.Run("ids assigned first-seen and stable", func(t *testing.T) {
		in := newInterner()
		assert.Equal(t, int64(0), in.id("movie"))
		assert.Equal(t, int64(1), in.id("short"))
		assert.Equal(t, int64(0), in.id("movie")) // already seen
		assert.Equal(t, int64(2), in.id("tvEpisode"))
	})

	t.Run("bit shifts one by the id", func(t *testing.T) {
		in := newInterner()
		assert.Equal(t, int64(1), in.bit("Action"))    // 1 << 0
		assert.Equal(t, int64(2), in.bit("Adventure")) // 1 << 1
		assert.Equal(t, int64(1), in.bit("Action"))    // already seen
	})

	t.Run("bit panics past the mask width", func(t *testing.T) {
		in := newInterner()
		for i := 0; i <= maxBit; i++ {
			in.bit(fmt.Sprintf("g%d", i)) // ids 0..maxBit are fine
		}
		assert.Panics(t, func() { in.bit("one too many") })
	})
}

func TestParseID(t *testing.T) {
	cases := []struct {
		in   string
		want int64
	}{
		{"tt0000001", 1},
		{"tt32857063", 32857063},
		{"nm0000001", 1},
	}
	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			got, err := parseID(tc.in)
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}

	t.Run("errors", func(t *testing.T) {
		_, err := parseID("tt")
		assert.Error(t, err)
		_, err = parseID("ttabc")
		assert.Error(t, err)
	})
}
