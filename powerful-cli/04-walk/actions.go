package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// shouldSkip returns true if given path should be ignored when walking
func shouldSkip(path, ext string, minSize int64, info os.FileInfo) bool {
	if info.IsDir() || info.Size() < minSize {
		return true
	}
	if ext != "" && filepath.Ext(path) != ext {
		return true
	}
	return false
}

// listFile prints path to out
func listFile(path string, out io.Writer) error {
	_, err := fmt.Fprintln(out, path)
	return err
}
