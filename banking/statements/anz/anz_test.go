package anz

import (
	"statements/common"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParseRow(t *testing.T) {
	expected := &common.Transaction{
		Date:      time.Date(2025, time.October, 21, 0, 0, 0, 0, time.UTC),
		Processed: time.Date(2025, time.October, 21, 0, 0, 0, 0, time.UTC),
		Account:   "4055-xxxx-1234",
		Details:   "Bob's Burgers",
		Amount:    -75.8,
	}

	t.Run("valid", func(t *testing.T) {
		row := []string{"21 Oct 2025", "21 Oct 2025", "4055-xxxx-1234", "Bob's Burgers", "$-75.80"}
		transaction, err := parseRow(row)
		assert.NoError(t, err)
		assert.Equal(t, expected, transaction)
	})

	t.Run("errors", func(t *testing.T) {
		row := []string{"", "", "", "", ""}
		wantErr := `parse row: invalid date: ""
invalid date: ""
invalid amount: ""`
		transaction, err := parseRow(row)
		assert.Nil(t, transaction)
		assert.ErrorContains(t, err, wantErr)
	})
}
