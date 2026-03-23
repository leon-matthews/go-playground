package main

import (
	"log"
	"os"

	pflag "github.com/spf13/pflag"
	"golang.org/x/term"

	"banking/categorise"
	"banking/common"
	"banking/statements"
	_ "banking/statements/anz" // register ANZ format
	_ "banking/statements/ofx" // register OFX format
	"banking/tui"
)

func main() {
	prefixesPath := pflag.StringP("prefixes", "p", "", "path to prefixes CSV file")
	bank := pflag.StringP("bank", "b", "", "bank format (e.g. anz); auto-detected if omitted")
	edit := pflag.BoolP("edit", "e", false, "interactively categorise unknown transactions")
	cats := pflag.BoolP("categories", "c", false, "edit category tree")
	verbose := pflag.CountP("verbose", "v", "increase category detail level")
	pflag.Parse()

	if *prefixesPath == "" {
		log.Fatal("Usage: banking --prefixes PREFIXES [--bank FORMAT] [--edit] [--categories] [STATEMENT ...]")
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
		log.Fatal("Usage: banking --prefixes PREFIXES [--bank FORMAT] [--edit] STATEMENT [STATEMENT ...]")
	}

	var transactions []*common.Transaction
	for _, path := range pflag.Args() {
		tt, err := readStatementFile(*bank, path)
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

func readStatementFile(bank, path string) ([]*common.Transaction, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var format statements.Format
	if bank != "" {
		f, ok := statements.Get(bank)
		if !ok {
			log.Fatalf("unknown bank format %q (available: %v)", bank, statements.Names())
		}
		format = f
	} else {
		f, err := statements.Detect(data)
		if err != nil {
			return nil, err
		}
		format = f
	}

	return format.Read(data)
}
