package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func BenchmarkRun(b *testing.B) {
	filenames, err := filepath.Glob("./testdata/benchmark/*.csv")
	require.NoError(b, err)

	for b.Loop() {
		if err := run(filenames, "mean", 2, io.Discard); err != nil {
			b.Error(err)
		}
	}
}

func TestRun(t *testing.T) {
	type testCase struct {
		column    int
		operation string
		files     []string
		want      string
		wantErr   error
	}

	testCases := map[string]testCase{
		"mean one file": {
			column:    3,
			operation: "mean",
			files:     []string{"testdata/example.csv"},
			want:      "227.6\n",
			wantErr:   nil,
		},
		"mean two files": {
			column:    3,
			operation: "mean",
			files:     []string{"testdata/example.csv", "testdata/example2.csv"},
			want:      "233.84\n",
			wantErr:   nil,
		},
		"error invalid path": {
			column:    2,
			operation: "mean",
			files:     []string{"testdata/example.csv", "testdata/no-such-file.csv"},
			want:      "",
			wantErr:   os.ErrNotExist,
		},
		"error no files": {
			column:    0,
			operation: "sum",
			files:     []string{},
			want:      "",
			wantErr:   ErrNoFiles,
		},
		"error invalid operation": {
			column:    2,
			operation: "magic",
			files:     []string{"testdata/example.csv"},
			want:      "",
			wantErr:   ErrInvalidOperation,
		},
	}

	for name, tt := range testCases {
		t.Run(name, func(t *testing.T) {
			var result bytes.Buffer
			err := run(tt.files, tt.operation, tt.column, &result)

			if tt.wantErr != nil {
				assert.Error(t, err, tt.wantErr)
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, result.String())
		})
	}
}
