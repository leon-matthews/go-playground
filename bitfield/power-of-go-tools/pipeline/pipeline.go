// Package pipeline implements shell-like operations using a 'fluent' API
package pipeline

import (
	"io"
	"os"
	"strings"
)

type Pipeline struct {
	Input  io.Reader
	Output io.Writer
	Error  error
}

func NewPipeline() *Pipeline {
	p := &Pipeline{
		Input:  os.Stdin,
		Output: os.Stdout,
	}
	return p
}

func FromString(s string) *Pipeline {
	p := NewPipeline()
	p.Input = strings.NewReader(s)
	return p
}

func (p *Pipeline) Stdout() {
	if p.Error != nil {
		return
	}
	io.Copy(p.Output, p.Input)
}
