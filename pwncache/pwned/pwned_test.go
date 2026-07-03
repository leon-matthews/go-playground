package pwned_test

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"

	"pwncache/pwned"
)

func TestPrefixes(t *testing.T) {
	t.Run("first five", func(t *testing.T) {
		p := make([]pwned.Prefix, 0, 5)
		for s := range pwned.Prefixes() {
			p = append(p, s)
			if len(p) >= 5 {
				break
			}
		}
		assert.Equal(t, 5, len(p))
		expected := []pwned.Prefix{"00000", "00001", "00002", "00003", "00004"}
		assert.Equal(t, expected, p)
	})

	t.Run("last five", func(t *testing.T) {
		// Is creating a million strings in a unit test excessive?
		s := slices.Collect(pwned.Prefixes())
		const expectedLength = 1_048_576
		assert.Equal(t, expectedLength, len(s))
		assert.Equal(t, s[0], pwned.Prefix("00000"))
		expected := []pwned.Prefix{"ffffb", "ffffc", "ffffd", "ffffe", "fffff"}
		assert.Equal(t, expected, s[expectedLength-len(expected):])
	})
}

func TestPrefixIndex(t *testing.T) {
	t.Run("maps a prefix to its integer value", func(t *testing.T) {
		cases := map[pwned.Prefix]int{
			"00000": 0,
			"00001": 1,
			"0000f": 15,
			"00010": 16,
			"fffff": pwned.PrefixCount - 1,
		}
		for prefix, want := range cases {
			got, err := prefix.Index()
			assert.NoError(t, err)
			assert.Equal(t, want, got)
		}
	})

	t.Run("round-trips the start of Prefixes", func(t *testing.T) {
		want := 0
		for prefix := range pwned.Prefixes() {
			got, err := prefix.Index()
			assert.NoError(t, err)
			assert.Equal(t, want, got)
			if want++; want >= 5 {
				break
			}
		}
	})

	t.Run("rejects invalid prefixes", func(t *testing.T) {
		for _, bad := range []pwned.Prefix{"", "abcd", "abcdef", "ghijk"} {
			_, err := bad.Index()
			assert.Error(t, err)
		}
	})
}

func BenchmarkPrefixes(b *testing.B) {
	const expectedLength = 1_048_576
	for b.Loop() {
		count := 0
		for range pwned.Prefixes() {
			count++
		}
		assert.Equal(b, expectedLength, count)
	}
}
