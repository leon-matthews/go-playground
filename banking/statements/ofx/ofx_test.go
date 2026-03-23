package ofx

import (
	"os"
	"testing"
	"time"
)

func TestDetect(t *testing.T) {
	t.Run("valid OFX file", func(t *testing.T) {
		data, err := os.ReadFile("testdata/valid.ofx")
		if err != nil {
			t.Fatal(err)
		}
		if err := Format.Detect(data); err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})

	t.Run("OFX header only", func(t *testing.T) {
		if err := Format.Detect([]byte("OFXHEADER:100\n")); err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})

	t.Run("OFX tag only", func(t *testing.T) {
		if err := Format.Detect([]byte("<OFX>\n</OFX>")); err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})

	t.Run("not OFX", func(t *testing.T) {
		err := Format.Detect([]byte("this is not OFX"))
		if err == nil {
			t.Fatal("expected error for non-OFX data")
		}
	})

	t.Run("Excel file", func(t *testing.T) {
		// First bytes of a zip/xlsx file
		err := Format.Detect([]byte{0x50, 0x4B, 0x03, 0x04})
		if err == nil {
			t.Fatal("expected error for Excel data")
		}
	})
}

func TestRead(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		data, err := os.ReadFile("testdata/valid.ofx")
		if err != nil {
			t.Fatal(err)
		}

		transactions, err := Format.Read(data)
		if err != nil {
			t.Fatal(err)
		}

		if len(transactions) != 3 {
			t.Fatalf("got %d transactions, want 3", len(transactions))
		}

		// First transaction: debit
		tx := transactions[0]
		if tx.Account != "1234-xxxx-5678" {
			t.Errorf("account = %q, want 1234-xxxx-5678", tx.Account)
		}
		wantDate := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
		if !tx.Date.Equal(wantDate) {
			t.Errorf("date = %v, want %v", tx.Date, wantDate)
		}
		if tx.Details != "Woolworths Auckland Nz" {
			t.Errorf("details = %q, want Woolworths Auckland Nz", tx.Details)
		}
		if tx.Amount != -42.50 {
			t.Errorf("amount = %.2f, want -42.50", tx.Amount)
		}

		// Second transaction: has memo
		tx = transactions[1]
		want := "Cafe Mocha Auckland Nz FXAmnt=4.00 FXCurr=AUD FXRate=0.6897"
		if tx.Details != want {
			t.Errorf("details = %q, want %q", tx.Details, want)
		}

		// Third transaction: credit
		tx = transactions[2]
		if tx.Amount != 500.00 {
			t.Errorf("amount = %.2f, want 500.00", tx.Amount)
		}
		if tx.Details != "Online Payment Thank You" {
			t.Errorf("details = %q, want Online Payment Thank You", tx.Details)
		}
	})

	t.Run("invalid data", func(t *testing.T) {
		_, err := Format.Read([]byte("not valid ofx"))
		if err == nil {
			t.Fatal("expected error for invalid data")
		}
	})
}
