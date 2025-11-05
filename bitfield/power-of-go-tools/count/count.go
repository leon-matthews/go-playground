package count

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
)

type counter struct {
	input  io.Reader
	output io.Writer
	files  []io.Reader
}

// NewCreate returns a new counter with zero or more options, using the 'functional options' pattern.
// Use options to customise built object, eg. NewCounter(WithInput(r))
func NewCounter(opts ...option) (*counter, error) {
	// Defaults
	c := &counter{
		input:  os.Stdin,
		output: os.Stdout,
	}

	// Use option constructors to override defaults
	for _, opt := range opts {
		err := opt(c)
		if err != nil {
			return nil, err
		}
	}
	return c, nil
}

type option func(*counter) error

// WithInput sets the reader of the counter
func WithInput(input io.Reader) option {
	return func(c *counter) error {
		if input == nil {
			return errors.New("nil input reader")
		}
		c.input = input
		return nil
	}
}

// WithInputFromArgs opens the given paths, setting the counter's input to them.
func WithInputFromArgs(paths []string) option {
	return func(c *counter) error {
		if len(paths) < 1 {
			return nil
		}

		c.files = make([]io.Reader, len(paths))
		for i, path := range paths {
			f, err := os.Open(path)
			if err != nil {
				return err
			}
			c.files[i] = f
		}

		c.input = io.MultiReader(c.files...)
		return nil
	}
}

// WithOutput sets the output writer of the counter
func WithOutput(output io.Writer) option {
	return func(c *counter) error {
		if output == nil {
			return errors.New("nil output writer")
		}
		c.output = output
		return nil
	}
}

func (c *counter) Lines() int {
	lines := 0
	input := bufio.NewScanner(c.input)
	for input.Scan() {
		lines++
	}
	c.close()
	return lines
}

func (c *counter) Words() int {
	s := bufio.NewScanner(c.input)
	s.Split(bufio.ScanWords)
	words := 0
	for s.Scan() {
		fmt.Printf("[%T]%+[1]v\n", s.Text())
		words++
	}
	c.close()
	return words
}

// close files that are closable
func (c *counter) close() {
	for _, f := range c.files {
		if closer, ok := f.(io.Closer); ok {
			closer.Close()
		}
	}
}

func MainLines() {
	counter, err := NewCounter(
		WithInputFromArgs(os.Args[1:]),
	)

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Println(counter.Lines())
}

func MainWords() {
	counter, err := NewCounter(
		WithInputFromArgs(os.Args[1:]),
	)

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Println(counter.Words())
}
