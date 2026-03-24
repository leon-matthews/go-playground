package main

import (
	"fmt"
	"log"
	"os"

	pflag "github.com/spf13/pflag"
	"golang.org/x/term"

	"banking/common"
	"banking/statements"
	_ "banking/statements/anz"      // register ANZ format
	_ "banking/statements/anz_visa" // register ANZ Visa format
	_ "banking/statements/ofx"      // register OFX format
	"banking/tui"
)

func main() {
	configPath := pflag.StringP("config", "f", "", "path to config JSON file")
	bank := pflag.StringP("bank", "b", "", "bank format (e.g. anz_visa); auto-detected if omitted")
	edit := pflag.BoolP("edit", "e", false, "interactively categorise unknown transactions")
	cats := pflag.BoolP("categories", "c", false, "edit category tree")
	verbose := pflag.CountP("verbose", "v", "increase category detail level")
	pflag.Parse()

	if *configPath == "" {
		log.Fatal("Usage: banking --config CONFIG [--bank FORMAT] [--edit] [--categories] [STATEMENT ...]")
	}

	cfg, err := common.LoadConfig(*configPath)
	if err != nil {
		log.Fatal(err)
	}

	if *cats {
		if err := tui.RunCategoryEditor(cfg, *configPath); err != nil {
			log.Fatal(err)
		}
		return
	}

	if pflag.NArg() == 0 {
		log.Fatal("Usage: banking --config CONFIG [--bank FORMAT] [--edit] STATEMENT [STATEMENT ...]")
	}

	var transactions []*common.Transaction
	for _, path := range pflag.Args() {
		tt, err := readStatementFile(*bank, path, cfg)
		if err != nil {
			log.Fatal(err)
		}
		transactions = append(transactions, tt...)
	}

	termWidth := 80
	if w, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil {
		termWidth = w
	}

	if err := tui.Run(transactions, cfg, *configPath, *verbose, termWidth, *edit); err != nil {
		log.Fatal(err)
	}
}

func readStatementFile(bank, path string, cfg *common.Config) ([]*common.Transaction, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var format statements.Format
	if bank != "" {
		f, ok := statements.Get(bank)
		if !ok {
			return nil, fmt.Errorf("unknown bank format %q (available: %v)", bank, statements.Names())
		}
		format = f
	} else {
		f, err := statements.Detect(data)
		if err != nil {
			return nil, err
		}
		format = f
	}

	transactions, err := format.Read(data)
	if err != nil {
		return nil, err
	}

	if sc, ok := cfg.Sources[format.Name()]; ok && len(sc.Delete) > 0 {
		replacer := common.NewReplacer(sc.Delete)
		for _, t := range transactions {
			t.Details = common.CleanDetails(t.Details, replacer)
		}
	}

	return transactions, nil
}
