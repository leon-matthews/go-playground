package main

import (
	"fmt"
	"log"

	"github.com/xuri/excelize/v2"

	"banking/excel"
)

func main() {
	f, err := excelize.OpenFile("/home/leon/Documents/visa/4055-xxxx-xxxx-6191_Statement_2025-07-15.xlsx")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	// Get all the rows in the Sheet1.
	rows, err := f.GetRows("Transactions")
	if err != nil {
		fmt.Println(err)
		return
	}

	for i, row := range rows {
		// Skip header
		if i == 0 {
			continue
		}
		t, err := excel.NewTransaction(row)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%s\n", t)
	}
}
