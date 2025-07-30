// Package excel reads data from ANZ credit card statements
package excel

import (
	"fmt"
	"regexp"
	"strconv"
)

func NewTransaction(row []string) (*Transaction, error) {
	amount, err := parseAmount(row[4])
	if err != nil {
		return nil, fmt.Errorf("creating transaction: %w", err)
	}
	t := Transaction{
		Date:      row[0],
		Processed: row[1],
		Card:      row[2],
		Details:   row[3],
		Amount:    amount,
	}
	return &t, nil
}

type Transaction struct {
	Date      string
	Processed string
	Card      string
	Details   string
	Amount    float64
}

func (t *Transaction) String() string {
	return fmt.Sprintf("Date: %s\nProcessed: %s\nCard: %s\nDetails: %q\nAmount: %.2f\n", t.Date, t.Processed, t.Card, t.Details, t.Amount)
}

var amountRegex = regexp.MustCompile("[^0-9\\.\\-]+")

func parseAmount(amount string) (float64, error) {
	cleaned := amountRegex.ReplaceAllString(amount, "")
	f, err := strconv.ParseFloat(cleaned, 64)
	if err != nil {
		return 0, fmt.Errorf("parse amount: %q", amount)
	}
	return f, nil
}
