// MDP Markdown Preview
package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

const (
	header = `<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<title>Markdown Preview Tool</title>
</head>
<body>`
	footer = `
</body>
</html>`
)

func main() {
	filename := flag.String("file", "", "Markdown file to preview")
	flag.Parse()

	if *filename == "" {
		flag.Usage()
		os.Exit(1)
	}

	if err := run(*filename); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Println(*filename)
}

func run(filename string) error {
	contents, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	html, err := parseContent(contents)
	if err != nil {
		return err
	}

	outname := fmt.Sprintf("%s.html", filepath.Base(filename))
	return saveHTML(outname, html)
}

func parseContent(contents []byte) (string, error) {
	var buffer bytes.Buffer
	buffer.WriteString(header)
	buffer.Write(contents)
	buffer.WriteString(footer)
	return buffer.String(), nil
}

func saveHTML(filename string, html string) error {
	fmt.Printf("=== %s ===", filename)
	fmt.Println(html)
	return nil
}
