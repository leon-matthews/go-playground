package main

import (
	"fmt"
	"log"
	"os"
	"text/tabwriter"

	"github.com/xuri/excelize/v2"

	"banking/excel"
)

func main() {
	if len(os.Args) != 2 {
		log.Printf("Usage: %v PATH\n", os.Args[0])
		os.Exit(1)
	}

	transactions, err := getTransactions(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}

	// Print in columns
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', tabwriter.AlignRight)
	for _, t := range transactions {
		fmt.Fprintln(w, t.Tabbed())
	}
	w.Flush()
}

const sheetName = "Transactions"

func getTransactions(path string) ([]*excel.Transaction, error) {
	f, err := excelize.OpenFile(path)
	if err != nil {
		return nil, fmt.Errorf("opening file: %v: %w", path, err)
	}
	defer f.Close()

	// Get all the rows in the right sheet
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, fmt.Errorf("get rows from sheet %q: %w", sheetName, err)
	}

	transactions := make([]*excel.Transaction, 0, len(rows))
	for i, row := range rows {
		// Skip header
		if i == 0 {
			continue
		}
		t, err := excel.NewTransaction(row)
		if err != nil {
			return nil, fmt.Errorf("new transaction from row %d: %w", i, err)
		}
		transactions = append(transactions, t)
	}

	return transactions, nil
}
