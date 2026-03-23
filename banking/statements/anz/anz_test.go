package anz

import (
	"os"
	"testing"
	"time"

	"banking/common"

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
		row := []string{"21 Oct 2025", "21 Oct 2025", "4055-xxxx-1234", "Bob's Burgers", "$-75.80", "", ""}
		transaction, err := parseRow(row)
		assert.NoError(t, err)
		assert.Equal(t, expected, transaction)
	})

	t.Run("trailing empty cells dropped", func(t *testing.T) {
		row := []string{"21 Oct 2025", "21 Oct 2025", "4055-xxxx-1234", "Bob's Burgers", "$-75.80"}
		transaction, err := parseRow(row)
		assert.NoError(t, err)
		assert.Equal(t, expected, transaction)
	})

	t.Run("errors", func(t *testing.T) {
		row := []string{"", "", "", "", "", "", ""}
		wantErr := `parse row: invalid date: ""
invalid date: ""
invalid amount: ""`
		transaction, err := parseRow(row)
		assert.Nil(t, transaction)
		assert.ErrorContains(t, err, wantErr)
	})
}

func TestRead(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		data, err := os.ReadFile("testdata/valid.xlsx")
		if err != nil {
			t.Fatal(err)
		}

		transactions, err := Format.Read(data)
		assert.NoError(t, err)
		assert.Len(t, transactions, 2)

		assert.Equal(t, time.Date(2025, time.October, 21, 0, 0, 0, 0, time.UTC), transactions[0].Date)
		assert.Equal(t, "4055-xxxx-1234", transactions[0].Account)
		assert.Equal(t, "Bob's Burgers", transactions[0].Details)
		assert.InDelta(t, -75.80, transactions[0].Amount, 0.001)

		assert.Equal(t, time.Date(2025, time.October, 22, 0, 0, 0, 0, time.UTC), transactions[1].Date)
		assert.Equal(t, time.Date(2025, time.October, 23, 0, 0, 0, 0, time.UTC), transactions[1].Processed)
		assert.Equal(t, "4055-xxxx-5678", transactions[1].Account)
		assert.Equal(t, "Coffee Shop", transactions[1].Details)
		assert.InDelta(t, -5.50, transactions[1].Amount, 0.001)
	})

	t.Run("invalid data", func(t *testing.T) {
		_, err := Format.Read([]byte("not excel"))
		assert.Error(t, err)
	})

	t.Run("wrong sheet", func(t *testing.T) {
		data, err := os.ReadFile("testdata/no_sheet.xlsx")
		if err != nil {
			t.Fatal(err)
		}
		_, err = Format.Read(data)
		assert.Error(t, err)
	})

	t.Run("bad row", func(t *testing.T) {
		data, err := os.ReadFile("testdata/bad_row.xlsx")
		if err != nil {
			t.Fatal(err)
		}
		_, err = Format.Read(data)
		assert.Error(t, err)
	})
}
