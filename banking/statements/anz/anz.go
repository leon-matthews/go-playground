// Package anz reads ANZ Excel bank account statements.
package anz

import (
	"bytes"
	"errors"
	"fmt"
	"strings"

	"github.com/xuri/excelize/v2"

	"banking/common"
	"banking/statements"
)

const sheetName = "Transactions"

var headers = []string{
	"Transaction Date", "Processed Date", "Type", "Details",
	"Particulars", "Code", "Reference", "Amount", "Balance",
	"To/From Account Number", "Conversion Charge", "Foreign Currency Amount",
}

type anzFormat struct{}

// Format is the ANZ bank account statement format.
var Format anzFormat

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

// parseRow builds a Transaction from a single spreadsheet row.
func parseRow(row []string) (*common.Transaction, error) {
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

	processed := date
	if row[1] != "" {
		p, err := common.ParseDate(common.DateFormat, row[1])
		if err != nil {
			rowErr = errors.Join(rowErr, err)
		} else {
			processed = p
		}
	}

	amount, err := common.ParseAmount(row[7])
	if err != nil {
		rowErr = errors.Join(rowErr, err)
	}

	if rowErr != nil {
		return nil, fmt.Errorf("parse row: %w", rowErr)
	}

	details := buildDetails(row[5], row[4], row[6], row[3])

	return &common.Transaction{
		Date:      date,
		Processed: processed,
		Details:   details,
		Amount:    amount,
	}, nil
}

// buildDetails joins non-empty detail fields with spaces and cleans the result.
func buildDetails(parts ...string) string {
	var non []string
	for _, p := range parts {
		if p != "" {
			non = append(non, p)
		}
	}
	return common.CleanString(strings.Join(non, " "))
}
