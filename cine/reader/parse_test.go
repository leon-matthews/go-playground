package reader

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	t.Run("optString", func(t *testing.T) {
		assert.Equal(t, "", optionalString(`\N`))
		assert.Equal(t, "US", optionalString("US"))
		assert.Equal(t, "", optionalString(""))
	})

	t.Run("optInt", func(t *testing.T) {
		n, err := optionalInt(`\N`)
		require.NoError(t, err)
		assert.Equal(t, missing, n)

		n, err = optionalInt("1987")
		require.NoError(t, err)
		assert.Equal(t, 1987, n)

		_, err = optionalInt("12x")
		assert.Error(t, err)
	})

	t.Run("reqInt", func(t *testing.T) {
		n, err := requiredInt("42")
		require.NoError(t, err)
		assert.Equal(t, 42, n)

		_, err = requiredInt(`\N`)
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

		// Raw UTF-8 has no backslash, so it stays on the fast path
		names, err = parseCharacters(`["Renée"]`)
		require.NoError(t, err)
		assert.Equal(t, []string{"Renée"}, names)

		// Escapes take the encoding/json fallback and are unescaped
		names, err = parseCharacters(`["\"Doc\" Holliday"]`)
		require.NoError(t, err)
		assert.Equal(t, []string{`"Doc" Holliday`}, names)

		names, err = parseCharacters(`["a\\b"]`)
		require.NoError(t, err)
		assert.Equal(t, []string{`a\b`}, names)

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

// jsonUnmarshal is the plain json-based semantics parseCharacters must
// reproduce: \N is nil, otherwise a JSON string array. It reports only whether
// decoding failed, since the wrapped error message is allowed to differ.
func jsonUnmarshal(s string) (names []string, failed bool) {
	if s == nullMarker {
		return nil, false
	}
	if err := json.Unmarshal([]byte(s), &names); err != nil {
		return nil, true
	}
	return names, false
}

// TestParseCharactersMatchesJSON locks the fast path to encoding/json across a
// table of well-formed, escaped, and malformed inputs.
func TestParseCharactersMatchesJSON(t *testing.T) {
	inputs := []string{
		`\N`,
		`["Self"]`,
		`["Man, seated","Bond"]`,
		`["A","B","C"]`,
		`[""]`,
		`[]`,
		`["\"Doc\" Holliday"]`,
		`["a\\b"]`,
		`[ "spaced" ]`,
		`["a"`,
		`[1,2]`,
		`["a"b"]`,
		`not json`,
	}
	for _, in := range inputs {
		t.Run(in, func(t *testing.T) {
			got, err := parseCharacters(in)
			want, failed := jsonUnmarshal(in)
			if failed {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, want, got)
		})
	}
}

// FuzzParseCharacters asserts the same equivalence on random input, so the fast
// path can never silently diverge from encoding/json.
func FuzzParseCharacters(f *testing.F) {
	seeds := []string{
		`\N`, `["Self"]`, `["Man, seated","Bond"]`, `[""]`, `[]`,
		`["\"Doc\" Holliday"]`, `["a\\b"]`, `["a"b"]`, `[1,2]`, `not json`,
	}
	for _, s := range seeds {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, s string) {
		got, err := parseCharacters(s)
		want, failed := jsonUnmarshal(s)
		if failed {
			assert.Error(t, err)
			return
		}
		require.NoError(t, err)
		assert.Equal(t, want, got)
	})
}
