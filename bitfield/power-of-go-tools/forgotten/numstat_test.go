package forgotten

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
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
		want := NumStat{
			{1, 2, "bitfield/power-of-go-tools/forgotten/forgotten.go"},
		}
		got, err := ParseDiffNumStat(text)
		assert.NoError(t, err)
		assert.Equal(t, want, got)
	})

	t.Run("empty", func(t *testing.T) {
		text := readTestFile(t, "empty.txt")
		got, err := ParseDiffNumStat(text)
		assert.NoError(t, err)
		want := NumStat{}
		assert.Equal(t, want, got)
	})
}

func TestStatusToJSON(t *testing.T) {
	t.Parallel()
	f := NumStat{
		{1, 2, "bitfield/power-of-go-tools/forgotten/forgotten.go"},
		{3, 4, "bitfield/power-of-go-tools/forgotten/forgotten_test.go"},
	}
	wantBytes, err := os.ReadFile("testdata/numstat.json")
	if err != nil {
		t.Fatal(err)
	}
	want := string(wantBytes)
	got := f.ToJSON()
	assert.Equal(t, want, got)
}
