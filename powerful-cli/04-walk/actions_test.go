package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShouldSkip_(t *testing.T) {
	t.Parallel()

	type testCase struct {
		file     string
		ext      string
		minSize  int64
		expected bool
	}

	testCases := map[string]testCase{
		"allows file when no filters set": {
			file:     "testdata/dir.log",
			expected: false,
		},
		"allows file when extension matches": {
			ext:      ".log",
			file:     "testdata/dir.log",
			expected: false,
		},
		"skips file if extension does not match": {
			ext:      ".gz",
			file:     "testdata/dir.log",
			expected: true,
		},
		"allows file if extension match and file not too-small": {
			file:     "testdata/dir.log",
			ext:      ".log",
			minSize:  10,
			expected: false,
		},
		"skips tool-small files, even if extension matches": {
			file:     "testdata/dir.log",
			ext:      ".log",
			minSize:  20,
			expected: true,
		},
	}

	for name, tt := range testCases {
		t.Run(name, func(t *testing.T) {
			info, err := os.Stat(tt.file)
			if err != nil {
				t.Fatal(err)
			}
			got := shouldSkip(tt.file, tt.ext, tt.minSize, info)
			assert.Equal(t, got, tt.expected)
		})
	}
}
