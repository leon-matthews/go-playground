package main

import (
	"bufio"
	"fmt"
	"io"
	"iter"
	"log"
	"os"
	"strings"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatal("Usage: banking experiment.go")
	}
	path := os.Args[1]
	log.Printf("Read lines from %q", path)
	r, err := fileOrStdin(path)
	if err != nil {
		log.Fatal(err)
	}
	defer r.Close()

	for l := range skipComments(lines(r), "#") {
		fmt.Println(l)
	}
}

// fileOrStdin opens either the given file or uses stdin.
// If the path is the string "-", stdin is used.
// The caller is expected to close the returned value.
func fileOrStdin(path string) (io.ReadCloser, error) {
	// Stdin
	if path == "-" {
		return os.Stdin, nil
	}

	// File
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	return f, nil
}

// lineReader return an iterator over the lines in stdid, or a text file.
// Stdin is used if the path is the string "-"
// Comments are lines that start with "#", and are excluded from the result.
func lines(r io.Reader) iter.Seq[string] {
	scanner := bufio.NewScanner(r)
	return func(yield func(string) bool) {
		for scanner.Scan() {
			line := scanner.Text()
			if !yield(line) {
				return
			}
		}
	}
}

// skipComments ignores lines that start with the given prefix
func skipComments(lines iter.Seq[string], prefix string) iter.Seq[string] {
	return func(yield func(string) bool) {
		for l := range lines {
			if strings.HasPrefix(l, prefix) {
				continue
			}
			if !yield(l) {
				return
			}
		}
	}
}
