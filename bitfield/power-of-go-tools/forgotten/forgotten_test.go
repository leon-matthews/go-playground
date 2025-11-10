package forgotten_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"forgotten"
)

func readTestFile(t *testing.T, filename string) string {
	t.Helper()
	path := filepath.Join("testdata", filename)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func TestParseDiffNumStat(t *testing.T) {
	t.Parallel()

	t.Run("one line", func(t *testing.T) {
		text := readTestFile(t, "one-line.txt")
		want := forgotten.Status{
			NumFiles: 1,
		}
		got, err := forgotten.ParseDiffNumStat(text)
		assert.NoError(t, err)
		assert.Equal(t, want, got)
	})

	t.Run("empty", func(t *testing.T) {
		text := readTestFile(t, "empty.txt")
		got, err := forgotten.ParseDiffNumStat(text)
		assert.NoError(t, err)
		want := forgotten.Status{}
		assert.Equal(t, want, got)
	})
}
