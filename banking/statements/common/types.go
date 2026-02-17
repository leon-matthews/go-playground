package common

import (
	"fmt"
	"io"
	"time"
)

const dateFormat = "2 Jan 2006"

// A statement contains zero or more Transaction
type Statement interface {
	Read(r io.Reader) ([]*Transaction, error)
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
	return fmt.Sprintf("Date: %s\nProcessed: %s\nAccount: %s\nDetails: %q\nAmount: %.2f\n", t.Date.Format(dateFormat), t.Processed.Format(dateFormat), t.Account, t.Details, t.Amount)
}

// Tabbed builds tab-separated strings ready for printing by [text/tabwriter]
func (t *Transaction) Tabbed() string {
	return fmt.Sprintf("%s\t%s\t%s\t$%.2f\t", t.Date.Format(dateFormat), t.Account, t.Details, t.Amount)
}

const (
	CategoryFood = 100
	CategoryFoodTakeaways = 101
	CategoryFoodRestaurant = 102
	CategoryFoodCafe = 103

	CategoryTransport = 200
	CategoryTransportFuel = 201
	CategoryTransportMaintenance = 202
	CategoryTransportPublic = 203

	CategoryCharity = 300
	CategoryCharityLocal = 301
	CategoryCharityInternational = 302

	CategoryHealth = 400

	CategoryInsurance = 500

	CategoryEducation = 600

	CategoryPhone = 700
}
