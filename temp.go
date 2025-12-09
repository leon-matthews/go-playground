package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

const maxDepth = 1

func main() {
	curdir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	splitPath := strings.Split(curdir, string(filepath.Separator))
	if len(splitPath) < maxDepth {
		log.Fatal("too shallow path")
	}

	allowedPatterns := filepath.Join("..")
	fmt.Printf("[%T]%+[1]v\n", allowedPatterns)

	args := filepath.Join(strings.Join(splitPath[:len(splitPath)-maxDepth], string(filepath.Separator)), "somefile.txt")
	fmt.Printf("[%T]%+[1]v\n", args)

}
