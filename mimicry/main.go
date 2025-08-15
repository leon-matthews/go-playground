package main

import (
	"fmt"
	"log"
	"path/filepath"

	"github.com/spf13/pflag"

	"mimicry/files"
)

var root string

func main() {
	pflag.Parse()
	args := pflag.Args()
	if len(args) != 1 {
		log.Fatal("usage: mimicry FOLDER")
	}
	root, err := filepath.Abs(args[0])
	fmt.Printf("[%T]%+[1]v\n", root)
	if err != nil {
		log.Fatal(err)
	}

	count := 0
	for path := range files.FindRecursive(root) {
		hash, _ := files.Sha256(path)
		fmt.Println(path)
		fmt.Println(hash)
		fmt.Println()
		count++
	}
	fmt.Println("Finished", count)
}
