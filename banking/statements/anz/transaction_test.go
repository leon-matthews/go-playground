package anz

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTransaction(t *testing.T) {
	transaction := Transaction{
		Date:      time.Date(2025, time.October, 21, 0, 0, 0, 0, time.UTC),
		Processed: time.Date(2025, time.October, 21, 0, 0, 0, 0, time.UTC),
		Account:   "4055-xxxx-1234",
		Details:   "Bob's Burgers",
		Amount:    -75.8,
	}

	t.Run("string", func(t *testing.T) {
		want := `Date: 21 Oct 2025
Processed: 21 Oct 2025
Account: 4055-xxxx-1234
Details: "Bob's Burgers"
Amount: -75.80
`
		got := transaction.String()
		fmt.Printf("[%T]%+[1]v\n", transaction)
		assert.Equal(t, want, got)
	})

	t.Run("tabbed", func(t *testing.T) {
		want := "21 Oct 2025\t4055-xxxx-1234\tBob's Burgers\t$-75.80\t"
		got := transaction.Tabbed()
		assert.Equal(t, want, got)
	})
}
