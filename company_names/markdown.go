package main

import (
	"fmt"
	"strings"
)

// Table wraps data needed to build a table in Markdown format
type Table struct {
	Header     []string
	Rows       [][]string
	numColumns int   // Length of every row, including the header
	widths     []int // Length of longest string in each column
}

// NewTable builds a table from the given headers
func NewTable(headers []string) *Table {
	table := &Table{
		Header:     headers,
		Rows:       make([][]string, 0),
		numColumns: len(headers),
		widths:     make([]int, len(headers)),
	}
	table.updateWidths(headers)
	return table
}

// AppendRow adds the given slice of string to the end of the table
func (t *Table) AppendRow(row []string) error {
	if len(row) != t.numColumns {
		return fmt.Errorf("expected %d columns, got %d", t.numColumns, len(row))
	}
	t.Rows = append(t.Rows, row)
	return nil
}

// NumRows counts the number of data rows in table
func (t *Table) NumRows() int {
	return len(t.Rows)
}

// Print builds the table as a Markdown string
func (t *Table) Print() string {
	b := new(strings.Builder)
	fmt.Fprintf(b, "%v\n", t.Header)
	for _, row := range t.Rows {
		fmt.Fprintf(b, "%v\n", row)
	}
	return b.String()
}

func (t *Table) updateWidths(row []string) {
	// Update widths
	for i := 0; i < len(t.widths); i++ {
		t.widths[i] = max(t.widths[i], len(row[i]))
	}
}
