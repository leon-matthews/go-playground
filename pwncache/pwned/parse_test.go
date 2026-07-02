package pwned_test

import (
	"encoding/hex"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"pwncache/pwned"
)

func TestParseHashList(t *testing.T) {
	t.Run("crlf endings", func(t *testing.T) {
		list := pwned.HashList(
			"003D68EB55068C33ACE09247EE4C639306B:3\r\n" +
				"012C192B2F16F82EA0EB9EF18D9D539B0DD:1247\r\n",
		)
		hashes, err := pwned.ParseHashList("cafe5", list)
		require.NoError(t, err)
		require.Len(t, hashes, 2)

		first, err := hex.DecodeString("cafe5003d68eb55068c33ace09247ee4c639306b")
		require.NoError(t, err)
		assert.Equal(t, first, hashes[0].SHA1)
		assert.Equal(t, int64(3), hashes[0].Count)
		assert.Len(t, hashes[1].SHA1, 20)
		assert.Equal(t, int64(1247), hashes[1].Count)
	})

	t.Run("bare newlines without trailing newline", func(t *testing.T) {
		list := pwned.HashList(
			"003D68EB55068C33ACE09247EE4C639306B:3\n" +
				"012C192B2F16F82EA0EB9EF18D9D539B0DD:1",
		)
		hashes, err := pwned.ParseHashList("cafe5", list)
		require.NoError(t, err)
		assert.Len(t, hashes, 2)
	})

	t.Run("empty list", func(t *testing.T) {
		hashes, err := pwned.ParseHashList("cafe5", "")
		require.NoError(t, err)
		assert.Empty(t, hashes)
	})

	t.Run("errors", func(t *testing.T) {
		tests := map[string]struct {
			list    pwned.HashList
			wantErr string
		}{
			"no separator": {
				list:    "003D68EB55068C33ACE09247EE4C639306B\r\n",
				wantErr: "hash list line 1: no colon separator",
			},
			"short suffix": {
				list:    "003D68EB:3\r\n",
				wantErr: "suffix must contain 35 characters",
			},
			"bad hex": {
				list:    "ZZZD68EB55068C33ACE09247EE4C639306B:3\r\n",
				wantErr: "hash list line 1",
			},
			"bad count": {
				list:    "003D68EB55068C33ACE09247EE4C639306B:many\r\n",
				wantErr: "hash list line 1",
			},
			"error reports line number": {
				list: "003D68EB55068C33ACE09247EE4C639306B:3\r\n" +
					"012C192B2F16F82EA0EB9EF18D9D539B0DD\r\n",
				wantErr: "hash list line 2: no colon separator",
			},
		}
		for name, tt := range tests {
			t.Run(name, func(t *testing.T) {
				hashes, err := pwned.ParseHashList("cafe5", tt.list)
				assert.Nil(t, hashes)
				assert.ErrorContains(t, err, tt.wantErr)
			})
		}
	})
}

func TestHashRange(t *testing.T) {
	t.Run("bounds", func(t *testing.T) {
		lower, upper, err := pwned.Prefix("cafe5").HashRange()
		require.NoError(t, err)
		assert.Equal(t, "cafe5"+strings.Repeat("0", 35), hex.EncodeToString(lower))
		assert.Equal(t, "cafe5"+strings.Repeat("f", 35), hex.EncodeToString(upper))
	})

	t.Run("first and last prefixes", func(t *testing.T) {
		lower, _, err := pwned.Prefix("00000").HashRange()
		require.NoError(t, err)
		assert.Equal(t, make([]byte, 20), lower)

		_, upper, err := pwned.Prefix("fffff").HashRange()
		require.NoError(t, err)
		assert.Equal(t, strings.Repeat("f", 40), hex.EncodeToString(upper))
	})

	t.Run("bad prefix", func(t *testing.T) {
		_, _, err := pwned.Prefix("zzzzz").HashRange()
		assert.ErrorContains(t, err, `prefix "zzzzz" is not hexadecimal`)
	})
}
