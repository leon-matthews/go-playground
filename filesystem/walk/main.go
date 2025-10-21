package main

import (
	"fmt"
	"io/fs"
	"iter"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/dustin/go-humanize"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Printf("Usage: %v PATH\n", os.Args[0])
		os.Exit(1)
	}

	root, err := filepath.Abs(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}

	err = filepathWalk(root)
	if err != nil {
		log.Fatal("filepath.Walk()", err)
	}

	err = filepathWalkDir(root)
	if err != nil {
		log.Fatal("filepath.WalkDir()", err)
	}

	// GitFolders
	start := time.Now()
	for p := range GitFolders(root) {
		log.Println(p)
	}
	log.Println("GitFolders took", time.Since(start))
}

// filepathWalk demonstrates the stdlib [filepath.Walk] function
func filepathWalk(root string) error {
	var numFiles, numFolders int64

	// visitor is of type [filepath.WalkFunc], as required by [filepath.Walk]
	// Returning [filepath.SkipDir] causes walk to skip the rest of the current folder,
	// while returning [filepath.SkipAll] causes all remaining files and folders to be skipped.
	visitor := func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			log.Printf("walkVisitor given a non-nil err for path: %q :%v\n", path, err)
			return err
		}
		if info.IsDir() {
			numFolders++
		} else {
			numFiles++

		}
		return nil
	}

	start := time.Now()
	err := filepath.Walk(root, visitor)
	log.Printf("filepath.Walk found %s files and %s folders in %v\n", humanize.Comma(numFiles), humanize.Comma(numFolders), time.Since(start))
	return err
}

// filepathWalkDir demonstrates the stdlib [filepath.WalkDir] function
func filepathWalkDir(root string) error {
	var numFiles, numFolders int64

	// visitor is of type [fs.WalkDirFunc], as used by [filepath.WalkDir]
	// Returning [io/fs.SkipDir] causes walk to skip the rest of the current folder,
	// while returning [io/fs.SkipAll] causes all remaining files and folders to be skipped.
	visitor := func(path string, info fs.DirEntry, err error) error {
		if err != nil {
			log.Printf("walkDirVisitor given a non-nil err for path: %q :%v\n", path, err)
			return err
		}
		if info.IsDir() {
			numFolders++
		} else {
			numFiles++
		}
		return nil
	}

	start := time.Now()
	err := filepath.WalkDir(root, visitor)
	log.Printf("filepath.WalkDir found %s files and %s folders in %v\n", humanize.Comma(numFiles), humanize.Comma(numFolders), time.Since(start))
	return err
}

// GitFolders iterates over all the git folders under the given root
func GitFolders(root string) iter.Seq[string] {
	return func(yield func(string) bool) {
		visitor := func(path string, info fs.DirEntry, err error) error {
			if info.IsDir() {
				// Yield (the parent) of git folders
				if info.Name() == ".git" {
					gitPath := filepath.Join(path, "..")
					if !yield(gitPath) {
						// Cancelled by caller
						return fs.SkipAll
					}
					return filepath.SkipDir
				}

				// Skip hidden folders
				if info.Name()[0] == '.' {
					return filepath.SkipDir
				}
			}
			return nil
		}
		_ = filepath.WalkDir(root, visitor)
	}
}
