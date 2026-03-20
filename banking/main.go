package main

import (
	"fmt"
	"log"
	"maps"
	"os"
	"path"
	"slices"
	"text/tabwriter"

	"banking/statements/anz"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatalf("Usage: %v PATH", path.Base(os.Args[0]))
	}

	transactions, err := anz.Read(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}

	// Accumulate card totals
	cards := make(map[string]float64)
	for _, t := range transactions {
		cards[t.Account] += t.Amount
	}

	// Print in columns
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', tabwriter.AlignRight)
	for _, t := range transactions {
		fmt.Fprintln(w, t.Tabbed())
	}
	w.Flush()

	// Print totals
	names := slices.Sorted(maps.Keys(cards))
	for _, name := range names {
		fmt.Printf("%s $%.2f\n", name, cards[name])
	}
}
