package main

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRun(t *testing.T) {
	t.Parallel()

	type testCase struct {
		root     string
		cfg      config
		expected string
	}

	testCases := map[string]testCase{
		"lists all files if no filters": {
			cfg: config{
				ext:     "",
				root:    "testdata",
				minSize: 0,
				list:    true,
			},
			expected: "testdata/dir.log\ntestdata/dir2/script.sh\n",
		},
		"lists only files that match extension": {
			cfg: config{
				ext:     ".log",
				root:    "testdata",
				minSize: 0,
				list:    true,
			},
			expected: "testdata/dir.log\n",
		},
		"lists file that match both extension and minimum size": {
			cfg: config{
				ext:     ".log",
				root:    "testdata",
				minSize: 10,
				list:    true,
			},
			expected: "testdata/dir.log\n",
		},
		"ignores small files even if extension matches": {
			cfg: config{
				ext:     ".log",
				root:    "testdata",
				minSize: 20,
				list:    true,
			},
			expected: "",
		},
		"ignores files if extension doesn't match": {
			cfg: config{
				ext:     ".gz",
				root:    "testdata",
				minSize: 0,
				list:    true,
			},
			expected: "",
		},
	}

	for name, tt := range testCases {
		t.Run(name, func(t *testing.T) {
			var buffer bytes.Buffer
			err := run(&buffer, tt.cfg)
			require.NoError(t, err)
			got := buffer.String()
			assert.Equal(t, tt.expected, got)
		})
	}
}
