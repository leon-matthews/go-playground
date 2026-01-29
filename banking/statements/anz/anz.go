// Package anz reads ANZ Excel credit card statements
package anz

import (
	"errors"
	"fmt"

	"github.com/xuri/excelize/v2"

	"statements/common"
)

const (
	dateFormat = "2 Jan 2006"
	sheetName  = "Transactions"
)

// Read produces Transaction values from a statement export file
func Read(path string) ([]*common.Transaction, error) {
	f, err := excelize.OpenFile(path)
	if err != nil {
		return nil, fmt.Errorf("opening file: %v: %w", path, err)
	}
	defer f.Close()

	// Get all the rows in the right sheet
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
	var rowErr error
	date, err := common.ParseDate(dateFormat, row[0])
	if err != nil {
		rowErr = errors.Join(rowErr, err)
	}
	processed, err := common.ParseDate(dateFormat, row[1])
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
