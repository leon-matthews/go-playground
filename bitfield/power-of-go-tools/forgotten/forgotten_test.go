package forgotten_test

import (
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"

	"forgotten"
)

func TestParseDiffNumStat(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile("testdata/git-diff-numstat.txt")
	if err != nil {
		t.Fatal(err)
	}
	want := forgotten.Status{
		NumFiles: 1,
	}
	got, err := forgotten.ParseDiffNumStat(string(data))
	if err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got))
	}
}
