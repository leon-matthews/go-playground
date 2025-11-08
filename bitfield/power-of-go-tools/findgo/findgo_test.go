package findgo_test

import (
	"archive/zip"
	"os"
	"testing"
	"testing/fstest"

	"github.com/google/go-cmp/cmp"

	"findgo"
)

func TestFilesCorrectlyListsFilesInTree(t *testing.T) {
	t.Parallel()
	want := []string{
		"file.go",
		"subfolder/subfolder.go",
		"subfolder2/another.go",
		"subfolder2/file.go",
	}
	tree, err := os.OpenRoot("testdata/tree")
	if err != nil {
		t.Fatal(err)
	}
	got := findgo.GoFiles(tree.FS())
	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got))
	}
}

func TestFilesCorrectlyListsFilesInMapFS(t *testing.T) {
	t.Parallel()
	fsys := fstest.MapFS{
		"file.go":                {},
		"subfolder/subfolder.go": {},
		"subfolder2/another.go":  {},
		"subfolder2/file.go":     {},
	}
	want := []string{
		"file.go",
		"subfolder/subfolder.go",
		"subfolder2/another.go",
		"subfolder2/file.go",
	}
	got := findgo.GoFiles(fsys)
	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got))
	}
}

func TestFilesCorrectlyListsFilesInZIPArchive(t *testing.T) {
	t.Parallel()
	fsys, err := zip.OpenReader("testdata/tree.zip")
	if err != nil {
		t.Fatal(err)
	}
	want := []string{
		"tree/file.go",
		"tree/subfolder/subfolder.go",
		"tree/subfolder2/another.go",
		"tree/subfolder2/file.go",
	}
	got := findgo.GoFiles(fsys)
	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got))
	}
}

func BenchmarkFilesOnDisk(b *testing.B) {
	fsys := os.DirFS("testdata/tree")
	b.ResetTimer()
	for b.Loop() {
		_ = findgo.GoFiles(fsys)
	}
}

func BenchmarkFilesInMemory(b *testing.B) {
	fsys := fstest.MapFS{
		"file.go":                {},
		"subfolder/subfolder.go": {},
		"subfolder2/another.go":  {},
		"subfolder2/file.go":     {},
	}
	b.ResetTimer()
	for b.Loop() {
		_ = findgo.GoFiles(fsys)
	}
}

func BenchmarkFilesInZip(b *testing.B) {
	fsys, err := zip.OpenReader("testdata/tree.zip")
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for b.Loop() {
		_ = findgo.GoFiles(fsys)
	}
}
