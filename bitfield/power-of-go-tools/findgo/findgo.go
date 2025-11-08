package findgo

import (
	"io/fs"
	"path/filepath"
)

// GoFiles recursively finds paths to all *.go files under tree
func GoFiles(tree fs.FS) (paths []string) {
	walker := func(p string, d fs.DirEntry, err error) error {
		if filepath.Ext(p) == ".go" {
			paths = append(paths, p)
		}
		return nil
	}
	fs.WalkDir(tree, ".", walker)
	return paths
}
