package main

import (
	"path/filepath"
	"fmt"
	"log"


	"github.com/spf13/pflag"

	"mimicry/internal/files"
)

var root string

func main() {
	pflag.Parse()
	args := pflag.Args()
	if len(args) != 1 {
		log.Fatal("usage: mimicry FOLDER")
	}
	root, err := filepath.Abs(args[0])
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
