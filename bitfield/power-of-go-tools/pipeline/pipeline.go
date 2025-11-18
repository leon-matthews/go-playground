// Package pipeline implements shell-like operations using a 'fluent' API
package pipeline

import (
	"io"
	"os"
	"strings"
)

// Pipeline is the sole function parameter and return type for chained operations
type Pipeline struct {
	Input  io.Reader
	Output io.Writer
	Error  error
}

// New returns a pointer to a new, default Pipeline using stdin & stdout
func New() *Pipeline {
	p := &Pipeline{
		Input:  os.Stdin,
		Output: os.Stdout,
	}
	return p
}

// FromString creates a new Pipeline using the given string as its input
func FromString(s string) *Pipeline {
	p := New()
	p.Input = strings.NewReader(s)
	return p
}

// Stdout simply copies pipeline input to output.
// It is a pipeline endpoint as it doesn't return a Pipeline.
func (p *Pipeline) Stdout() {
	if p.Error != nil {
		return
	}
	io.Copy(p.Output, p.Input)
}
