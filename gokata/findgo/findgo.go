// Package findgo recursively searches for *.go files in a filesystem.
// Taken from The Power of Go: Tools, ch. 6
package findgo

import (
	"io/fs"
	"path/filepath"
)

func Files(fsys fs.FS) (paths []string) {
	fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if filepath.Ext(path) == ".go" {
			paths = append(paths, path)
		}
		return nil
	})
	return paths
}
