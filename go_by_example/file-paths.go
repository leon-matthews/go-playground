package main

import (
	"fmt"
	"path/filepath"
	"strings"
)

func main() {
	// Join()
	p := filepath.Join("dir1", "dir2", "file.txt")
	fmt.Println(p)
	// dir1/dir2/file.txt

	p = filepath.Join(p, "../file2.txt")
	fmt.Println(p)
	// dir1/dir2/file2.txt

	// Base() & Dir()
	fmt.Println(filepath.Base(p))
	fmt.Println(filepath.Dir(p))
	// file2.txt
	// dir1/dir2

	// filepath.Ext() & strings.TrimSuffix()
	filename := "config.json"
	ext := filepath.Ext(filename)
	fmt.Println(ext)
	fmt.Println(strings.TrimSuffix(filename, ext))
	// .json
	// config

	// Abs()
	fmt.Println(filepath.IsAbs(p))
	p, err := filepath.Abs(p)
	if err != nil {
		panic(err)
	}
	fmt.Println(filepath.IsAbs(p))
	fmt.Println(p)
	// false
	// true
	// /home/leon/Projects/go-playground/go_by_example/dir1/dir2/file2.txt

	// Rel()
	// Find a relative path between base (first argument) and target (second)
	rel, err := filepath.Rel("a/b", "a/b/t/file")
	if err != nil {
		panic(err)
	}
	fmt.Println(rel)
	// t/file

	rel, err = filepath.Rel("a/b", "a/c/t/file")
	if err != nil {
		panic(err)
	}
	fmt.Println(rel)
	// ../c/t/file
}
