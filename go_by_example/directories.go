package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

func check(e error) {
    if e != nil {
        panic(e)
    }
}

func main() {
	// Create a folder in $TMP, then schedule its deletion
	dir, err := os.MkdirTemp("", "go-by-example-directories-*")
	check(err)
	defer os.RemoveAll(dir)	// rm -fr !!!

	// Utility closures
	j := func(relpath string) string {
		return filepath.Join(dir, relpath)
	}

	createEmptyFile := func(relpath string) {
		d := []byte("")
		path := filepath.Join(dir, relpath)
		err := os.WriteFile(path, d, 0666)
		check(err)
	}

	// Populate folder
	createEmptyFile("file1.txt")
	err = os.MkdirAll(j("subdir/parent/child"), 0777)
	check(err)
	createEmptyFile("subdir/parent/file2.pdf")
    createEmptyFile("subdir/parent/file3.json")
    createEmptyFile("subdir/parent/child/file4.py")

    // os.ReadDir() returns a slice of os.DirEntry objects
    ls, err := os.ReadDir(dir)
    check(err)
	for _, entry := range ls {
        fmt.Println(" ", entry.Name(), entry.IsDir())
    }

    // os.Chdir()
    err = os.Chdir(j("subdir/parent"))
    check(err)
    ls, err = os.ReadDir(".")
    check(err)
    for _, entry := range ls {
        fmt.Println(" ", entry.Name(), entry.IsDir())
    }

    // filepath.WalkDir() visits file system tree recursively
    fmt.Println("Visiting", dir)
    err = filepath.WalkDir(dir, visit)
    check(err)
}

// visit is called for every file and folder by [filepath.Walkdir]
func visit(path string, info fs.DirEntry, err error) error {
	if err != nil {
		return err
	}
	fmt.Print(" ", path)
	if info.IsDir() {
		fmt.Println("/")
	} else {
		fmt.Println()
	}
	return nil
}
