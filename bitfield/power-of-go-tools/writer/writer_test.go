package writer_test

import (
	"os"
	"path/filepath"
	"testing"

	"writer"

	"github.com/google/go-cmp/cmp"
)

func TestWriteToFile_WritesGivenDataToFile(t *testing.T) {
	// Action
	t.Parallel()
	path := filepath.Join(t.TempDir(), "write_test.txt")
	want := []byte{1, 2, 3}
	err := writer.WriteToFile(path, want)
	if err != nil {
		t.Fatal(err)
	}

	// Exists with correct mode
	stat, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	perm := stat.Mode().Perm()
	if perm != 0o600 {
		t.Errorf("want file mode 0o600, got 0o%o", perm)
	}

	// Contents correct?
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(want, got) {
		t.Fatal(cmp.Diff(want, got))
	}
}

func TestWriteToFile_ReturnsErrorForUnwritableFile(t *testing.T) {
	t.Parallel()
	path := "bogusdir/write_test.txt"
	err := writer.WriteToFile(path, []byte{})
	if err == nil {
		t.Fatal("want error when file not writable")
	}
}

func TestWriteToFile_ClobbersExistingFile(t *testing.T) {
	// Create pre-existing file
	t.Parallel()
	path := filepath.Join(t.TempDir(), "clobber_test.txt")
	err := os.WriteFile(path, []byte{4, 5, 6}, 0o600)
	if err != nil {
		t.Fatal(err)
	}

	// Clobberin' time!
	want := []byte{1, 2, 3}
	err = writer.WriteToFile(path, want)
	if err != nil {
		t.Fatal(err)
	}

	// Contents correct?
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(want, got) {
		t.Fatal(cmp.Diff(want, got))
	}
}

func TestWriteToFile_ChangesPermsOnExistingFile(t *testing.T) {
	// Pre-create empty file with open perms
	t.Parallel()
	path := filepath.Join(t.TempDir(), "/perms_test.txt")
	err := os.WriteFile(path, []byte{}, 0o644)
	if err != nil {
		t.Fatal(err)
	}

	// Action
	err = writer.WriteToFile(path, []byte{})
	if err != nil {
		t.Fatal(err)
	}

	// Perms correct?
	stat, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	perm := stat.Mode().Perm()
	if perm != 0o600 {
		t.Errorf("want file mode 0600, got 0%o", perm)
	}
}
