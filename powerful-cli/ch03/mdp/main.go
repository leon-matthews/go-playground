// MDP Markdown Preview
package main

import (
	"bytes"
	_ "embed"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/microcosm-cc/bluemonday"
	"github.com/russross/blackfriday/v2"
)

//go:embed "template.html"
var htmlTemplate string

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
}

func run(filename string) error {
	contents, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	body := parseContent(contents)
	html, err := buildHTML(body)
	if err != nil {
		return err
	}
	outname := fmt.Sprintf("%s.html", filepath.Base(filename))
	err = os.WriteFile(outname, html, 0666)
	return err
}

// parseContent converts Markdown file into sanitized HTML
func parseContent(input []byte) []byte {
	// Convert to HTML using default options
	unsafe := blackfriday.Run(input)

	// Sanitize HTML output using 'User Generated Content' policy
	body := bluemonday.UGCPolicy().SanitizeBytes(unsafe)
	return body
}

// buildHTML embeds given body output using HTML template
func buildHTML(body []byte) ([]byte, error) {
	// Use text/template here, as we don't need to escape our generated HTML
	t := template.Must(template.New("name").Parse(htmlTemplate))
	var html bytes.Buffer
	err := t.Execute(&html, string(body))
	if err != nil {
		return []byte{}, err
	}
	return html.Bytes(), nil
}
