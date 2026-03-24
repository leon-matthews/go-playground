package anz

import (
	"os"
	"testing"
	"time"

	"banking/common"

	"github.com/stretchr/testify/assert"
)

func TestParseRow(t *testing.T) {
	t.Run("all fields populated", func(t *testing.T) {
		row := []string{
			"18 Mar 2026", "18 Mar 2026", "Visa Purchase",
			"4835-****-****-9371", "Df", "Waterview Co", "",
			"- $5.75", "- $94,042.03", "", "", "",
		}
		tx, err := parseRow(row)
		assert.NoError(t, err)
		assert.Equal(t, &common.Transaction{
			Date:      time.Date(2026, time.March, 18, 0, 0, 0, 0, time.UTC),
			Processed: time.Date(2026, time.March, 18, 0, 0, 0, 0, time.UTC),
			Details:   "Waterview Co Df 4835-****-****-9371",
			Amount:    -5.75,
		}, tx)
	})

	t.Run("empty processed date falls back to transaction date", func(t *testing.T) {
		row := []string{
			"23 Mar 2026", "", "Visa Hold",
			"4835-****-****-9371", "Df", "Waterview Co", "",
			"- $5.75", "- $95,377.29", "", "", "",
		}
		tx, err := parseRow(row)
		assert.NoError(t, err)
		assert.Equal(t, time.Date(2026, time.March, 23, 0, 0, 0, 0, time.UTC), tx.Date)
		assert.Equal(t, tx.Date, tx.Processed)
	})

	t.Run("payment with reference and account", func(t *testing.T) {
		row := []string{
			"19 Mar 2026", "19 Mar 2026", "Payment",
			"Watercare Serv Ltd", "Watercare 24E", "Fairland", "5434857-01",
			"- $69.83", "- $94,420.56", "02-0192-0115055-02", "", "",
		}
		tx, err := parseRow(row)
		assert.NoError(t, err)
		assert.Equal(t, "Fairland Watercare 24E 5434857-01 Watercare Serv Ltd", tx.Details)
		assert.InDelta(t, -69.83, tx.Amount, 0.001)
	})

	t.Run("salary credit", func(t *testing.T) {
		row := []string{
			"04 Mar 2026", "04 Mar 2026", "Salary",
			"Auckland Transport", "Wage/Salary", "", "",
			"$3,985.32", "- $89,540.15", "", "", "",
		}
		tx, err := parseRow(row)
		assert.NoError(t, err)
		assert.Equal(t, "Wage/Salary Auckland Transport", tx.Details)
		assert.InDelta(t, 3985.32, tx.Amount, 0.001)
	})

	t.Run("trailing empty cells dropped", func(t *testing.T) {
		row := []string{
			"18 Mar 2026", "18 Mar 2026", "Salary",
			"Auckland Transport", "Wage/Salary", "", "",
			"$3,985.32",
		}
		tx, err := parseRow(row)
		assert.NoError(t, err)
		assert.InDelta(t, 3985.32, tx.Amount, 0.001)
	})

	t.Run("errors", func(t *testing.T) {
		row := []string{"", "", "", "", "", "", "", "", "", "", "", ""}
		tx, err := parseRow(row)
		assert.Nil(t, tx)
		assert.ErrorContains(t, err, `invalid date: ""`)
		assert.ErrorContains(t, err, `invalid amount: ""`)
	})
}

func TestBuildDetails(t *testing.T) {
	t.Run("skips empty fields", func(t *testing.T) {
		assert.Equal(t, "Auckland Transport Wage/Salary", buildDetails("Auckland Transport", "Wage/Salary", "", ""))
	})

	t.Run("all empty", func(t *testing.T) {
		assert.Equal(t, "", buildDetails("", "", "", ""))
	})

	t.Run("cleans whitespace", func(t *testing.T) {
		assert.Equal(t, "Anz S3D730 St Lukes Br", buildDetails("Anz   S3D730 St Lukes Br", "", "", ""))
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
		assert.Len(t, transactions, 12)

		// Row 0: Visa Hold — empty processed date falls back to transaction date
		tx := transactions[0]
		assert.Equal(t, time.Date(2026, time.March, 15, 0, 0, 0, 0, time.UTC), tx.Date)
		assert.Equal(t, tx.Date, tx.Processed)
		assert.Equal(t, "Corner Cafe Df 4111-****-****-1234", tx.Details)
		assert.InDelta(t, -6.50, tx.Amount, 0.001)
		assert.Empty(t, tx.Account)

		// Row 1: Visa Purchase
		tx = transactions[1]
		assert.Equal(t, time.Date(2026, time.March, 14, 0, 0, 0, 0, time.UTC), tx.Processed)
		assert.Equal(t, "Fresh Grocer Df 4111-****-****-1234", tx.Details)
		assert.InDelta(t, -87.30, tx.Amount, 0.001)

		// Row 3: Eft-Pos with all detail fields
		tx = transactions[3]
		assert.Equal(t, "C 4111******** 1234 260313142507 Hardware World Mt Eden", tx.Details)
		assert.InDelta(t, -24.99, tx.Amount, 0.001)

		// Row 5: Payment with reference
		tx = transactions[5]
		assert.Equal(t, "Ref 9988 Rates 2026 7712345-01 City Council", tx.Details)
		assert.InDelta(t, -310.00, tx.Amount, 0.001)

		// Row 6: Salary credit
		tx = transactions[6]
		assert.Equal(t, "Wage/Salary Acme Corp Ltd", tx.Details)
		assert.InDelta(t, 4250.00, tx.Amount, 0.001)

		// Row 10: ATM — different processed date
		tx = transactions[10]
		assert.Equal(t, time.Date(2026, time.March, 6, 0, 0, 0, 0, time.UTC), tx.Date)
		assert.Equal(t, time.Date(2026, time.March, 7, 0, 0, 0, 0, time.UTC), tx.Processed)
		assert.InDelta(t, -200.00, tx.Amount, 0.001)
	})

	t.Run("invalid data", func(t *testing.T) {
		_, err := Format.Read([]byte("not excel"))
		assert.Error(t, err)
	})
}
