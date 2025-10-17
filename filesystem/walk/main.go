package main

import (
	"fmt"
	"io/fs"
	"iter"
	"log"
	"os"
	"path/filepath"
	"time"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Printf("Usage: %v PATH", os.Args[0])
		os.Exit(1)
	}

	root, err := filepath.Abs(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}

	/*
		// filepath.Walk()
		start := time.Now()
		err = filepath.Walk(root, walkVisitor)
		if err != nil {
			log.Fatal("filepath.Walk()", err)
		}
		log.Println("filepath.Walk took", time.Since(start))

		// filepath.WalkDir()
		start = time.Now()
		err = filepath.WalkDir(root, walkDirVisitor)
		if err != nil {
			log.Fatal("filepath.WalkDir()", err)
		}
		log.Println("filepath.WalkDir took", time.Since(start))
	*/

	// GitFolders
	start := time.Now()
	for p := range GitFolders(root) {
		log.Println(p)
	}
	log.Println("GitFolders took", time.Since(start))
}

// GitFolders iterates over all the git folders under the given root
func GitFolders(root string) iter.Seq[string] {
	return func(yield func(string) bool) {
		visitor := func(path string, info fs.DirEntry, err error) error {
			fmt.Println(path)

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

// walkVisitor is of type [filepath.WalkFunc], as required by [filepath.Walk]
// Returning [filepath.SkipDir] causes walk to skip the rest of the current folder,
// while returning [filepath.SkipAll] causes all remaining files and folders to be skipped.
func walkVisitor(path string, info fs.FileInfo, err error) error {
	if err != nil {
		log.Printf("walkVisitor given a non-nil err for path: %q :%v", path, err)
		return err
	}
	if info.IsDir() {
		fmt.Println("visited folder:", info.Name())
	} else {
		fmt.Println("visited file:  ", info.Name())
	}
	return nil
}

// walkDirVisitor is of type [fs.WalkDirFunc], as used by [filepath.WalkDir]
// Returning [io/fs.SkipDir] causes walk to skip the rest of the current folder,
// while returning [io/fs.SkipAll] causes all remaining files and folders to be skipped.
func walkDirVisitor(path string, info fs.DirEntry, err error) error {
	if err != nil {
		log.Printf("walkDirVisitor given a non-nil err for path: %q :%v", path, err)
		return err
	}
	if info.IsDir() {
		fmt.Println("visited folder:", info.Name())
	} else {
		fmt.Println("visited file:  ", info.Name())
	}
	return nil
}
