// Package anz reads ANZ Excel credit card statements
package anz

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
)

const (
	dateFormat = "2 Jan 2006"
	sheetName  = "Transactions"
)

var (
	amountRegexp  = regexp.MustCompile(`[^0-9.\-]+`)
	detailsRegexp = regexp.MustCompile(`\s+`)
)

// Read produces Transaction values from a statement export file
func Read(path string) ([]*Transaction, error) {
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

	transactions := make([]*Transaction, 0, len(rows))
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

func parseRow(row []string) (*Transaction, error) {
	var rowErr error
	date, err := parseDate(row[0])
	if err != nil {
		rowErr = errors.Join(rowErr, err)
	}
	processed, err := parseDate(row[1])
	if err != nil {
		rowErr = errors.Join(rowErr, err)
	}
	card := row[2]
	details := cleanString(row[3])
	amount, err := parseAmount(row[4])
	if err != nil {
		rowErr = errors.Join(rowErr, err)
	}

	fmt.Printf("[%T]%+[1]v\n", rowErr)
	if rowErr != nil {
		return nil, fmt.Errorf("parse row: %w", rowErr)
	}

	t := Transaction{
		Date:      date,
		Processed: processed,
		Account:   card,
		Details:   details,
		Amount:    amount,
	}
	return &t, nil
}

// cleanString removes repeated spaces and trims ends from given string
func cleanString(s string) string {
	clean := detailsRegexp.ReplaceAllString(s, " ")
	return strings.TrimSpace(clean)
}

// parseAmount reads a floating point value from the format "$-166.99"
func parseAmount(amount string) (float64, error) {
	cleaned := amountRegexp.ReplaceAllString(amount, "")
	f, err := strconv.ParseFloat(cleaned, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid amount: %q", amount)
	}
	return f, nil
}

// parseDate creates a timestamp from a date in the form '14 Jun 2025'
func parseDate(date string) (time.Time, error) {
	t, err := time.Parse("02 Jan 2006", date)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid date: %q", date)
	}
	return t, nil
}
