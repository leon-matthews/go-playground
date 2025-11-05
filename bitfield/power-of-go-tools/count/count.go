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
		if len(paths) == 0 {
			return nil
		}

		// TODO Just the first file, for now...
		f, err := os.Open(paths[0])
		if err != nil {
			return err
		}
		c.input = f
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
	return lines
}

func Main() int {
	counter, err := NewCounter(
		WithInputFromArgs(os.Args[1:]),
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	fmt.Println(counter.Lines())
	return 0
}
