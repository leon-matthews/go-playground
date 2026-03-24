package common

import (
	"fmt"
	"time"
)

const DateFormat = "2 Jan 2006"

// Prefix maps a lowercase merchant prefix to its category.
type Prefix struct {
	Text     string `json:"prefix"`
	Category string `json:"category"`
}

// SourceConfig holds per-statement-source settings.
type SourceConfig struct {
	Delete []string `json:"delete"`
}

// Config is the top-level configuration for the banking app.
type Config struct {
	Prefixes []Prefix                `json:"prefixes"`
	Sources  map[string]SourceConfig `json:"sources"`
}

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
	return fmt.Sprintf("Date: %s\nProcessed: %s\nAccount: %s\nDetails: %q\nAmount: %.2f\n", t.Date.Format(DateFormat), t.Processed.Format(DateFormat), t.Account, t.Details, t.Amount)
}

// Tabbed builds tab-separated strings ready for printing by [text/tabwriter]
func (t *Transaction) Tabbed() string {
	return fmt.Sprintf("%s\t%s\t%s\t$%.2f\t", t.Date.Format(DateFormat), t.Account, t.Details, t.Amount)
}
