package findgo_test

import (
	"testing"
	"testing/fstest"

	"github.com/gokatas/findgo"
	"github.com/google/go-cmp/cmp"
)

func TestFilesFindsGoFiles(t *testing.T) {
	t.Parallel()
	fsys := fstest.MapFS{
		"file.go":            {},
		"file.pl":            {},
		"dir/file.go":        {},
		"dir/file.pl":        {},
		"dir/another.go":     {},
		"dir/subdir/file.go": {},
	}
	want := []string{
		"dir/another.go",
		"dir/file.go",
		"dir/subdir/file.go",
		"file.go",
	}
	got := findgo.Files(fsys)
	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got))
	}
}
