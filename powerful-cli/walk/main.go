package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
)

type config struct {
	ext  string // extension to filter out
	size int    // minimum file size
	list bool   // just list files
}

func main() {
	// Parse options
	root := flag.String("root", ".", "Directory to start scanning")
	list := flag.Bool("list", false, "List files only")
	ext := flag.String("ext", "", "File extension to filter out")
	size := flag.Int("size", 0, "Minimum file size")
	flag.Parse()
	options := config{
		ext:  *ext,
		size: *size,
		list: *list,
	}

	if err := run(*root, os.Stdout, options); err != nil {
		log.Fatal(err)
	}
}

func run(root string, out io.Writer, options config) error {
	return fmt.Errorf("Just starting")
}
