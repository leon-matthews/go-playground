package main

import (
	"log"
	"os"

	pflag "github.com/spf13/pflag"
	"golang.org/x/term"

	"banking/categorise"
	"banking/common"
	"banking/statements/anz"
	"banking/tui"
)

func main() {
	prefixesPath := pflag.StringP("prefixes", "p", "", "path to prefixes CSV file")
	edit := pflag.BoolP("edit", "e", false, "interactively categorise unknown transactions")
	cats := pflag.BoolP("categories", "c", false, "edit category tree")
	verbose := pflag.CountP("verbose", "v", "increase category detail level")
	pflag.Parse()

	if *prefixesPath == "" {
		log.Fatal("Usage: banking --prefixes PREFIXES [--edit] [--categories] [STATEMENT ...]")
	}

	prefixes, err := categorise.LoadPrefixes(*prefixesPath)
	if err != nil {
		log.Fatal(err)
	}

	if *cats {
		if err := tui.RunCategoryEditor(prefixes, *prefixesPath); err != nil {
			log.Fatal(err)
		}
		return
	}

	if pflag.NArg() == 0 {
		log.Fatal("Usage: banking --prefixes PREFIXES [--edit] STATEMENT [STATEMENT ...]")
	}

	var transactions []*common.Transaction
	for _, path := range pflag.Args() {
		tt, err := anz.Read(path)
		if err != nil {
			log.Fatal(err)
		}
		transactions = append(transactions, tt...)
	}

	termWidth := 80
	if w, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil {
		termWidth = w
	}

	if err := tui.Run(transactions, prefixes, *prefixesPath, *verbose, termWidth, *edit); err != nil {
		log.Fatal(err)
	}
}
