// Package anz reads ANZ Excel credit card statements
package anz

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/xuri/excelize/v2"

	"banking/common"
	"banking/statements"
)

const sheetName = "Transactions"

var headers = []string{"Transaction Date", "Processed Date", "Card", "Details", "Amount", "Conversion Charge", "Foreign Currency Amount"}

// Format is the ANZ statement format.
var Format anzFormat

type anzFormat struct{}

func init() {
	statements.Register(&Format)
}

func (anzFormat) Name() string { return "anz" }

// Detect checks whether the data contains an Excel file with a "Transactions"
// sheet and the expected header row.
func (anzFormat) Detect(data []byte) error {
	f, err := excelize.OpenReader(bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("not an Excel file")
	}
	defer f.Close()

	rows, err := f.GetRows(sheetName)
	if err != nil {
		return fmt.Errorf("sheet %q not found", sheetName)
	}
	if len(rows) == 0 {
		return fmt.Errorf("sheet %q is empty", sheetName)
	}
	if len(rows[0]) < len(headers) {
		return fmt.Errorf("expected at least %d columns, got %d", len(headers), len(rows[0]))
	}
	for i, h := range headers {
		if rows[0][i] != h {
			return fmt.Errorf("expected column %q, got %q", h, rows[0][i])
		}
	}
	return nil
}

// Read produces Transaction values from statement data.
func (anzFormat) Read(data []byte) ([]*common.Transaction, error) {
	f, err := excelize.OpenReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("opening file: %w", err)
	}
	defer f.Close()

	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, fmt.Errorf("get rows from sheet %q: %w", sheetName, err)
	}

	transactions := make([]*common.Transaction, 0, len(rows))
	for i, row := range rows {
		// Skip header
		if i == 0 {
			continue
		}
		t, err := parseRow(row)
		if err != nil {
			return nil, fmt.Errorf("new transaction from row %d: %w", i, err)
		}
		transactions = append(transactions, t)
	}

	return transactions, nil
}

// parseRow builds a Transaction from a single spreadsheet row
func parseRow(row []string) (*common.Transaction, error) {
	// excelize drops trailing empty cells; pad to expected length.
	if len(row) < len(headers) {
		padded := make([]string, len(headers))
		copy(padded, row)
		row = padded
	}
	var rowErr error
	date, err := common.ParseDate(common.DateFormat, row[0])
	if err != nil {
		rowErr = errors.Join(rowErr, err)
	}
	processed, err := common.ParseDate(common.DateFormat, row[1])
	if err != nil {
		rowErr = errors.Join(rowErr, err)
	}
	card := row[2]
	details := common.CleanString(row[3])
	amount, err := common.ParseAmount(row[4])
	if err != nil {
		rowErr = errors.Join(rowErr, err)
	}

	if rowErr != nil {
		return nil, fmt.Errorf("parse row: %w", rowErr)
	}

	t := common.Transaction{
		Date:      date,
		Processed: processed,
		Account:   card,
		Details:   details,
		Amount:    amount,
	}
	return &t, nil
}
