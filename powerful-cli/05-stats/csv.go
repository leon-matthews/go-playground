package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
)

// statsFunc defines a generic statistical function
type statsFunc func(data []float64) float64

func mean(data []float64) float64 {
	return sum(data) / float64(len(data))
}

func sum(data []float64) float64 {
	var sum float64
	for _, v := range data {
		sum += v
	}
	return sum
}

// csv2float reads and converts floats from one column from the given CSV file
// The first row is skipped. columnIndex starts at one.
func csv2float(input io.Reader, columnIndex int) ([]float64, error) {
	// Adjust from one-based to zero-based
	columnIndex--

	// Read in all CSV data
	r := csv.NewReader(input)
	data, err := r.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("reading csv data: %w", err)
	}

	var column []float64
	for rowIndex, row := range data {
		// Skip first row
		if rowIndex == 0 {
			continue
		}

		// Checking number of columns in CSV file
		if len(row) <= columnIndex {
			// File does not have that many columns
			return nil, fmt.Errorf("%w: file has only %d columns", ErrInvalidColumn, len(row))
		}

		// Try to convert data read into a float number
		v, err := strconv.ParseFloat(row[columnIndex], 64)
		if err != nil {
			return nil, fmt.Errorf("row: %v: %w: %s", rowIndex, ErrNotNumber, err)
		}
		column = append(column, v)

	}

	return column, nil
}
