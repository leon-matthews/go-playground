package imdb

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	t.Run("optString", func(t *testing.T) {
		assert.Equal(t, "", optString(`\N`))
		assert.Equal(t, "US", optString("US"))
		assert.Equal(t, "", optString(""))
	})

	t.Run("optInt", func(t *testing.T) {
		n, err := optInt(`\N`)
		require.NoError(t, err)
		assert.Nil(t, n)

		n, err = optInt("1987")
		require.NoError(t, err)
		require.NotNil(t, n)
		assert.Equal(t, 1987, *n)

		_, err = optInt("12x")
		assert.Error(t, err)
	})

	t.Run("reqInt", func(t *testing.T) {
		n, err := reqInt("42")
		require.NoError(t, err)
		assert.Equal(t, 42, n)

		_, err = reqInt(`\N`)
		assert.Error(t, err)
	})

	t.Run("parseFloat", func(t *testing.T) {
		f, err := parseFloat("5.7")
		require.NoError(t, err)
		assert.InDelta(t, 5.7, f, 1e-9)

		_, err = parseFloat("high")
		assert.Error(t, err)
	})

	t.Run("parseBool", func(t *testing.T) {
		b, err := parseBool("0")
		require.NoError(t, err)
		assert.False(t, b)

		b, err = parseBool("1")
		require.NoError(t, err)
		assert.True(t, b)

		_, err = parseBool(`\N`)
		assert.Error(t, err)

		_, err = parseBool("2")
		assert.Error(t, err)
	})

	t.Run("splitList", func(t *testing.T) {
		assert.Nil(t, splitList(`\N`))
		assert.Nil(t, splitList(""))
		assert.Equal(t, []string{"a"}, splitList("a"))
		assert.Equal(t, []string{"a", "b", "c"}, splitList("a,b,c"))
	})

	t.Run("parseCharacters", func(t *testing.T) {
		names, err := parseCharacters(`\N`)
		require.NoError(t, err)
		assert.Nil(t, names)

		names, err = parseCharacters(`["Self"]`)
		require.NoError(t, err)
		assert.Equal(t, []string{"Self"}, names)

		// Commas inside a JSON element stay within that element
		names, err = parseCharacters(`["Man, seated","Bond"]`)
		require.NoError(t, err)
		assert.Equal(t, []string{"Man, seated", "Bond"}, names)

		_, err = parseCharacters(`[oops`)
		assert.Error(t, err)
	})
}

func TestSplitTabs(t *testing.T) {
	cases := []struct {
		name string
		line string
		want []string
	}{
		{"three fields", "a\tb\tc", []string{"a", "b", "c"}},
		{"single field", "solo", []string{"solo"}},
		{"empty line", "", []string{""}},
		{"empty middle", "a\t\tc", []string{"a", "", "c"}},
		{"trailing tab", "a\tb\t", []string{"a", "b", ""}},
	}

	// One dst is reused across cases, exercising the buffer reuse (and shrink)
	var dst []string
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dst = splitTabs(tc.line, dst)
			assert.Equal(t, tc.want, dst)
		})
	}
}
