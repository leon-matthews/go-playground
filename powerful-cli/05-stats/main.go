package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"sync"
)

func main() {
	// Verify and parse arguments
	op := flag.String("op", "sum", "Operation to be executed")
	column := flag.Int("col", 1, "CSV column on which to execute operation")

	flag.Parse()

	if err := run(flag.Args(), *op, *column, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	r := NewRunner(flag.Args(), *op, *column, os.Stdout)
	if err := r.run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

type runner struct {
	filenames []string
	column    int
	funcName  string
	output    io.Writer
}

func NewRunner(args []string, funcName string, column int, out io.Writer) *runner {
	return &runner{
		filenames: args,
		column:    column,
		funcName:  funcName,
		output:    out,
	}
}

// process attempts to extract floats from a column in a file using [csv2Float].
// If successful, its results are sent to resultStream, otherwise the first error
// is sent to errorStream
func (r *runner) process(filename string, column int, resultStream chan<- []float64, errorStream chan<- error) {
	// Open the file for reading
	f, err := os.Open(filename)
	if err != nil {
		errorStream <- err
		return
	}

	// Parse the CSV into a slice of float64 numbers
	data, err := csv2Float(f, r.column)
	if err != nil {
		errorStream <- err
		return
	}

	if err := f.Close(); err != nil {
		errorStream <- err
		return
	}

	// Send results out
	resultStream <- data
}

// run calculates and prints out the result
func (r *runner) run() error {
	// Validate input
	err := r.validate()
	if err != nil {
		return err
	}

	// Operation?
	operation, err := r.operation(r.funcName)
	if err != nil {
		return err
	}

	// Consolidate data from same column across all files
	resultStream := make(chan []float64)
	errorStream := make(chan error)
	doneStream := make(chan struct{})

	// Start one goroutine per file
	wg := sync.WaitGroup{}
	for fname := range r.filenameGenerator() {
		wg.Go(func() {
			r.process(fname, r.column, resultStream, errorStream)
		})
	}

	// Wait for all goroutines to finish
	go func() {
		wg.Wait()
		close(doneStream)
	}()

	// Collect results
	consolidated := make([]float64, 0)
	for {
		select {
		case result := <-resultStream:
			// Gather results together
			consolidated = append(consolidated, result...)
		case err := <-errorStream:
			// Abort on first error
			return err
		case <-doneStream:
			// Finished? Run operation and exit
			_, err := fmt.Fprintln(r.output, operation(consolidated))
			return err
		}
	}
}

// filenameGenerator builds a generator over filenames
func (r *runner) filenameGenerator() <-chan string {
	filenameStream := make(chan string)
	go func() {
		defer close(filenameStream)
		for _, filename := range r.filenames {
			filenameStream <- filename
		}
	}()
	return filenameStream
}

// operation select function to perform
func (r *runner) operation(name string) (statsFunc, error) {
	var opFunc statsFunc
	switch name {
	case "sum":
		opFunc = sum
	case "mean":
		opFunc = mean
	default:
		return nil, fmt.Errorf("%w: %s", ErrInvalidOperation, opFunc)
	}

	return opFunc, nil
}

// validate returns first problem found
func (r *runner) validate() error {
	if len(r.filenames) == 0 {
		return ErrNoFiles
	}

	if r.column < 0 {
		return fmt.Errorf("%w: %d", ErrInvalidColumn, r.column)
	}

	return nil
}

func run(filenames []string, op string, column int, out io.Writer) error {
	// Validate input
	if len(filenames) == 0 {
		return ErrNoFiles
	}
	if column < 0 {
		return fmt.Errorf("%w: %d", ErrInvalidColumn, column)
	}

	// Select operation to perform
	var opFunc statsFunc
	switch op {
	case "sum":
		opFunc = sum
	case "mean":
		opFunc = mean
	default:
		return fmt.Errorf("%w: %s", ErrInvalidOperation, op)
	}

	// Goroutine
	filenameStream := make(chan string)
	go func() {
		defer close(filenameStream)
		for _, filename := range filenames {
			filenameStream <- filename
		}
	}()

	// Consolidate data from same column across all files
	resultStream := make(chan []float64)
	errorStream := make(chan error)
	doneStream := make(chan struct{})

	wg := sync.WaitGroup{}
	for fname := range filenameStream {
		wg.Go(func() {
			// Open the file for reading
			f, err := os.Open(fname)
			if err != nil {
				errorStream <- err
			}
			// Parse the CSV into a slice of float64 numbers
			data, err := csv2Float(f, column)
			if err != nil {
				errorStream <- err
			}
			if err := f.Close(); err != nil {
				errorStream <- err
			}

			// Send results out
			resultStream <- data
		})
	}

	// Wait for all goroutines to finish
	go func() {
		wg.Wait()
		close(doneStream)
	}()

	// Collect results
	consolidated := make([]float64, 0)
	for {
		select {
		case result := <-resultStream:
			// Gather results together
			consolidated = append(consolidated, result...)
		case err := <-errorStream:
			// Abort on first error
			return err
		case <-doneStream:
			// Finished? Run operation and exit
			_, err := fmt.Fprintln(out, opFunc(consolidated))
			return err
		}
	}
}
