// Package pipeline implements shell-like operations using a 'fluent' API
package pipeline

import (
	"io"
	"os"
	"strings"
)

// Pipeline is the sole function parameter and return type for chained operations
type Pipeline struct {
	Reader io.Reader
	Writer io.Writer
	Error  error
}

// New returns a pointer to a new, default Pipeline using stdin & stdout
func New() *Pipeline {
	// Reader does not default to os.Stdin by design: forgetting to set it
	// during development would cause pipelines to stall waiting for user input.
	p := &Pipeline{
		Writer: os.Stdout,
	}
	return p
}

// FromString creates a new Pipeline using the given string as its input
func FromString(s string) *Pipeline {
	p := New()
	p.Reader = strings.NewReader(s)
	return p
}

// FromFile creates a new Pipeline reading from file
func FromFile(file string) *Pipeline {
	f, err := os.Open(file)
	if err != nil {
		return &Pipeline{Error: err}
	}
	p := New()
	p.Reader = f
	return p
}

// Stdout simply copies pipeline input to output.
// It is a pipeline endpoint as it doesn't return a Pipeline.
func (p *Pipeline) Stdout() {
	if p.Error != nil {
		return
	}
	io.Copy(p.Writer, p.Reader)
}

// String is a pipeline endpoint that builds string from input
func (p *Pipeline) String() (string, error) {
	if p.Error != nil {
		return "", p.Error
	}

	data, err := io.ReadAll(p.Reader)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
