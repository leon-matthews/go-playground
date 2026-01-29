package common

import (
	"fmt"
	"time"
)

const dateFormat = "2 Jan 2006"

// Transaction holds basic details on a single financial event
// A negative Amount indicates outgoing funds, eg. a purchase.
// The Date and Processed fields have no time information
type Transaction struct {
	Date      time.Time
	Processed time.Time
	Account   string
	Details   string
	Amount    float64
}

// String builds long-form string representation
func (t *Transaction) String() string {
	return fmt.Sprintf("Date: %s\nProcessed: %s\nAccount: %s\nDetails: %q\nAmount: %.2f\n", t.Date.Format(dateFormat), t.Processed.Format(dateFormat), t.Account, t.Details, t.Amount)
}

// Tabbed builds tab-separated strings ready for printing by [text/tabwriter]
func (t *Transaction) Tabbed() string {
	return fmt.Sprintf("%s\t%s\t%s\t$%.2f\t", t.Date.Format(dateFormat), t.Account, t.Details, t.Amount)
}
