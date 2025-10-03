// Package excel reads data from ANZ credit card statements
package excel

import (
	"fmt"
	"regexp"
	"strconv"
	"time"
)

func NewTransaction(row []string) (*Transaction, error) {
	amount, err := parseAmount(row[4])
	date, err := parseDate(row[0])
	processed, err := parseDate(row[1])
	if err != nil {
		return nil, fmt.Errorf("creating transaction: %w", err)
	}
	t := Transaction{
		Date:      date,
		Processed: processed,
		Card:      row[2],
		Details:   row[3],
		Amount:    amount,
	}
	return &t, nil
}

type Transaction struct {
	Date      time.Time
	Processed time.Time
	Card      string
	Details   string
	Amount    float64
}

func (t *Transaction) String() string {
	return fmt.Sprintf("Date: %s\nProcessed: %s\nCard: %s\nDetails: %q\nAmount: %.2f\n", t.Date, t.Processed, t.Card, t.Details, t.Amount)
}

var amountRegex = regexp.MustCompile(`[^0-9.\-]+`)

// parseAmount reads a floating point value from the format "$-166.99"
func parseAmount(amount string) (float64, error) {
	cleaned := amountRegex.ReplaceAllString(amount, "")
	f, err := strconv.ParseFloat(cleaned, 64)
	if err != nil {
		return 0, fmt.Errorf("parse amount: %q", amount)
	}
	return f, nil
}

// parseDate creates a timestamp from a date in the form '14 Jun 2025'
func parseDate(date string) (time.Time, error) {
	t, err := time.Parse("02 Jan 2006", date)
	if err != nil {
		return time.Time{}, fmt.Errorf("parsing date: %w", err)
	}
	return t, nil
}
