package main

import (
	"findgo"
	"flag"
	"fmt"
	"os"
)

const Usage = `Usage: findgo FOLDER`

func main() {
	flag.Parse()
	if len(flag.Args()) < 1 {
		fmt.Fprintln(os.Stderr, Usage)
		os.Exit(1)
	}
	root, err := os.OpenRoot(flag.Arg(0))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	
	paths := findgo.GoFiles(root.FS())
	for _, path := range paths {
		fmt.Println(path)
	}
}
