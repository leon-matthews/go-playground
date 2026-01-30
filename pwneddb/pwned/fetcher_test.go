package pwned_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"pwneddb/pwned"
)

func TestBuildURL(t *testing.T) {
	p := "abcde"
	prefix, err := pwned.NewPrefix(p)
	require.NoError(t, err)
	url := pwned.BuildURL(prefix)
	assert.Equal(t, "https://api.pwnedpasswords.com/range/abcde", url)
}

func TestNewPrefix(t *testing.T) {
	p, err := pwned.NewPrefix("abcde")
	require.NoError(t, err)
	assert.Equal(t, pwned.Prefix("abcde"), p)
}

func TestNewPrefixError(t *testing.T) {
	tests := map[string]struct {
		prefix  string
		wantErr string
	}{
		"empty": {
			prefix:  "",
			wantErr: `prefix must contain 5 characters: ""`,
		},
		"too short": {
			prefix:  "abcd",
			wantErr: `prefix must contain 5 characters: "abcd"`,
		},
		"too long": {
			prefix:  "abcdef",
			wantErr: `prefix must contain 5 characters: "abcdef"`,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := pwned.NewPrefix(tt.prefix)
			assert.Equal(t, pwned.Prefix(""), got)
			assert.ErrorContains(t, err, tt.wantErr)
		})
	}
}
