package files

import (
	"io/fs"
	"path/filepath"
)

// FindRecursive sends
func FindRecursive(root string) <-chan string  {
	paths := make(chan string)

	go func() {
		defer close(paths)
		filepath.WalkDir(root, func(path string, info fs.DirEntry, err error) error {
			if !info.IsDir() {
				paths <- path
			}
			return nil
		})
	}()

    return paths
}
