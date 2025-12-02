// MDP Markdown Preview
package main

import (
	"bytes"
	_ "embed"
	"flag"
	"fmt"
	"html/template"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/gomarkdown/markdown"
	"github.com/microcosm-cc/bluemonday"
)

//go:embed "template.html"
var htmlTemplate string

type content struct {
	Title string
	Body  template.HTML
}

func main() {
	filename := flag.String("file", "", "Markdown file to preview")
	skipPreview := flag.Bool("s", false, "Skip auto-preview")
	flag.Parse()

	if *filename == "" {
		flag.Usage()
		os.Exit(1)
	}

	if err := run(*filename, *skipPreview); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(filename string, skipPreview bool) error {
	// Read and parse Markdown file
	contents, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	body := parseContent(contents)
	c := content{
		Title: "Markdown Preview",
		Body:  template.HTML(body),
	}
	html, err := buildHTML(c)
	if err != nil {
		return err
	}

	// Write HTML to temporary file
	temp, err := os.CreateTemp("", "mdp-*.html")
	if err != nil {
		return err
	}
	if err := temp.Close(); err != nil {
		return err
	}
	outname := temp.Name()
	defer func() {
		time.Sleep(1 * time.Second)
		os.Remove(outname)
	}()

	err = os.WriteFile(outname, html, 0666)
	if err != nil {
		return err
	}

	// Preview HTML in system browser
	if skipPreview {
		return nil
	}
	return preview(outname)
}

// parseContent converts Markdown file into sanitized HTML
func parseContent(input []byte) []byte {
	// Convert to HTML using default options
	unsafe := markdown.ToHTML(input, nil, nil)

	// Sanitize HTML output using 'User Generated Content' policy
	body := bluemonday.UGCPolicy().SanitizeBytes(unsafe)
	return body
}

// preview opens HTML file in default system web browser
func preview(filename string) error {
	command := ""
	params := []string{}

	switch runtime.GOOS {
	case "linux":
		command = "xdg-open"
	case "windows":
		command = "cmd.exe"
	case "darwin":
		command = "open"
	default:
		return fmt.Errorf("OS not supported")
	}

	params = append(params, filename)
	commandPath, err := exec.LookPath(command)
	if err != nil {
		return err
	}
	return exec.Command(commandPath, params...).Run()
}

// buildHTML embeds given body output using HTML template
func buildHTML(c content) ([]byte, error) {
	t := template.Must(template.New("name").Parse(htmlTemplate))
	var html bytes.Buffer
	err := t.Execute(&html, c)
	if err != nil {
		return []byte{}, err
	}
	return html.Bytes(), nil
}
