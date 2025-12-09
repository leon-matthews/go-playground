package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"

	flag "github.com/spf13/pflag"
)

type config struct {
	root    string // Path to start searching
	ext     string // extension to filter out
	minSize int64  // minimum file minSize
	list    bool   // just list files
}

func parseArgs() config {
	options := config{}
	flag.StringVarP(&options.root, "root", "r", ".", "Directory to start scanning")
	flag.BoolVarP(&options.list, "list", "l", false, "List files only")
	flag.StringVarP(&options.ext, "ext", "e", "", "File extension to filter out")
	flag.Int64VarP(&options.minSize, "size", "s", 0, "Minimum file size")
	help := flag.BoolP("help", "h", false, "show this help")

	flag.Usage = func() {
		whoami := path.Base(os.Args[0])
		fmt.Fprintf(flag.CommandLine.Output(), "%s: print example help\n\n", whoami)
		fmt.Fprintf(flag.CommandLine.Output(), "usage: %s [-h]\n", whoami)
		flag.PrintDefaults()
	}
	flag.Parse()

	if *help {
		flag.CommandLine.SetOutput(os.Stdout)
		flag.Usage()
		os.Exit(0)
	}
	return options
}

func main() {
	options := parseArgs()

	// Run
	err := run(os.Stdout, options)
	if err != nil {
		log.Fatal(err)
	}
}

func run(out io.Writer, options config) error {
	err := filepath.Walk(options.root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if shouldSkip(path, options.ext, options.minSize, info) {
			return nil
		}

		if options.list {
			return listFile(path, out)
		}

		return listFile(path, out)
	})
	return err
}
