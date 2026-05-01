// Package main implements a File Duplicate Scanner.
package main

import (
	"crypto/sha256"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
)

// FileInfo track per-file statistics
type FileInfo struct {
	Path      string
	Size      int64
	Extension string
	Hash      string
}

// ExtensionStats tracks total per extension
type ExtensionStats struct {
	Count int
	Size  int64
}

// collectFiles builds a slice of absolute paths to all the files under root
func collectFiles(root string) ([]string, error) {
	var paths []string
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}

	err = filepath.WalkDir(absRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			// Missing root?
			if path == absRoot {
				return err
			}

			// skip files we can't read
			return nil
		}

		if !d.IsDir() {
			paths = append(paths, path)
		}
		return nil
	})
	log.Println(paths)
	return paths, err
}

// hashFile calculates a SHA-256 hash for the file with the given path
func hashFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

func main() {
	fmt.Println("Hello")
}
