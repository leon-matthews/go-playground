package anz

import (
	"os"
	"testing"

	"github.com/xuri/excelize/v2"
)

func TestDetect(t *testing.T) {
	t.Run("valid ANZ file", func(t *testing.T) {
		data, err := os.ReadFile("testdata/valid.xlsx")
		if err != nil {
			t.Fatal(err)
		}

		if err := Format.Detect(data); err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})

	t.Run("wrong headers", func(t *testing.T) {
		data := excelBytes(t, func(f *excelize.File) {
			f.SetSheetName("Sheet1", "Transactions")
			f.SetSheetRow("Transactions", "A1", &[]string{"Wrong", "Headers", "Here"})
		})

		err := Format.Detect(data)
		if err == nil {
			t.Fatal("expected error for wrong headers")
		}
	})

	t.Run("too few columns", func(t *testing.T) {
		data := excelBytes(t, func(f *excelize.File) {
			f.SetSheetName("Sheet1", "Transactions")
			f.SetSheetRow("Transactions", "A1", &[]string{"Date"})
		})

		err := Format.Detect(data)
		if err == nil {
			t.Fatal("expected error for too few columns")
		}
	})

	t.Run("empty sheet", func(t *testing.T) {
		data := excelBytes(t, func(f *excelize.File) {
			f.SetSheetName("Sheet1", "Transactions")
		})

		err := Format.Detect(data)
		if err == nil {
			t.Fatal("expected error for empty sheet")
		}
		if want := `sheet "Transactions" is empty`; err.Error() != want {
			t.Errorf("error = %q, want %q", err.Error(), want)
		}
	})

	t.Run("not an Excel file", func(t *testing.T) {
		err := Format.Detect([]byte("this is not excel"))
		if err == nil {
			t.Fatal("expected error for non-Excel data")
		}
	})

	t.Run("ANZ Visa headers rejected", func(t *testing.T) {
		data := excelBytes(t, func(f *excelize.File) {
			f.SetSheetName("Sheet1", "Transactions")
			f.SetSheetRow("Transactions", "A1", &[]string{
				"Transaction Date", "Processed Date", "Card", "Details",
				"Amount", "Conversion Charge", "Foreign Currency Amount",
			})
		})

		err := Format.Detect(data)
		if err == nil {
			t.Fatal("expected error for ANZ Visa headers")
		}
	})
}

// excelBytes creates an in-memory Excel file and returns its bytes.
func excelBytes(t *testing.T, setup func(f *excelize.File)) []byte {
	t.Helper()
	f := excelize.NewFile()
	setup(f)
	path := t.TempDir() + "/test.xlsx"
	if err := f.SaveAs(path); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return data
}
