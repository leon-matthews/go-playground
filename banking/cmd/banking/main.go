package main

import (
	"fmt"
	"log/slog"
	"os"
	"slices"

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
	debug := pflag.Bool("debug", false, "enable debug logging")
	pflag.Parse()

	level := slog.LevelWarn
	if *debug {
		level = slog.LevelDebug
	}
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level})))

	if *configPath == "" {
		fmt.Fprintln(os.Stderr, "Usage: banking --config CONFIG [--bank FORMAT] [--edit] [--categories] [STATEMENT ...]")
		os.Exit(1)
	}

	slog.Debug("loading config", "path", *configPath)
	cfg, err := common.LoadConfig(*configPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	knownFormats := statements.Names()
	for name := range cfg.Sources {
		if !slices.Contains(knownFormats, name) {
			fmt.Fprintf(os.Stderr, "config: unknown source %q (available: %v)\n", name, knownFormats)
			os.Exit(1)
		}
	}

	if *cats {
		if err := tui.RunCategoryEditor(cfg, *configPath); err != nil {
			slog.Error("category editor failed", "error", err)
			os.Exit(1)
		}
		return
	}

	if pflag.NArg() == 0 {
		fmt.Fprintln(os.Stderr, "Usage: banking --config CONFIG [--bank FORMAT] [--edit] STATEMENT [STATEMENT ...]")
		os.Exit(1)
	}

	var transactions []*common.Transaction
	for _, path := range pflag.Args() {
		slog.Debug("reading statement", "path", path, "bank", *bank)
		tt, err := readStatementFile(*bank, path, cfg)
		if err != nil {
			slog.Error("failed to read statement", "path", path, "error", err)
			os.Exit(1)
		}
		transactions = append(transactions, tt...)
	}

	slog.Info("processing complete", "transactions", len(transactions))

	termWidth := 80
	if w, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil {
		termWidth = w
	}

	if err := tui.Run(transactions, cfg, *configPath, *verbose, termWidth, *edit); err != nil {
		slog.Error("run failed", "error", err)
		os.Exit(1)
	}
}

func readStatementFile(bank, path string, cfg *common.Config) ([]*common.Transaction, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read statement %s: %w", path, err)
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
			return nil, fmt.Errorf("detect format of %s: %w", path, err)
		}
		format = f
	}

	slog.Debug("using format", "path", path, "format", format.Name())

	transactions, err := format.Read(data)
	if err != nil {
		return nil, fmt.Errorf("read %s as %s: %w", path, format.Name(), err)
	}

	slog.Debug("parsed transactions", "path", path, "count", len(transactions))

	if sc, ok := cfg.Sources[format.Name()]; ok && len(sc.Delete) > 0 {
		replacer := common.NewReplacer(sc.Delete)
		for _, t := range transactions {
			t.Details = common.CleanDetails(t.Details, replacer)
		}
	}

	return transactions, nil
}
