package count

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
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

func WithArgs(args []string) option {
	return func(c *counter) error {
		f, err := os.Open(args[0])
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

func Main() {
	counter, err := NewCounter()
	if err != nil {
		log.Fatalf("error creating counter: %v", err)
	}
	fmt.Println(counter.Lines())
}
