package tui

import (
	"fmt"
	"os"
	"slices"

	tea "charm.land/bubbletea/v2"

	"banking/categorise"
	"banking/common"
)

// Run categorises transactions, prints the summary tree, and optionally
// launches the interactive editor for unknown transactions.
func Run(transactions []*common.Transaction, prefixes []categorise.Prefix, prefixesPath string, verbose, termWidth int, edit bool) error {
	matcher := categorise.NewMatcher(prefixes)

	var expenses, income categorise.Summary
	var unknowns []*common.Transaction
	expensesByCategory := make(map[string][]*common.Transaction)
	incomeByCategory := make(map[string][]*common.Transaction)
	for _, t := range transactions {
		cat := matcher.Match(t.Details)
		if t.Amount < 0 {
			expenses.Add(cat, t.Details, t.Amount)
			expensesByCategory[cat] = append(expensesByCategory[cat], t)
		} else {
			income.Add(cat, t.Details, t.Amount)
			incomeByCategory[cat] = append(incomeByCategory[cat], t)
		}
		if cat == categorise.Unknown {
			unknowns = append(unknowns, t)
		}
	}

	expenses.Sort()
	income.Sort()

	maxDepth := verbose + 1
	showTransactions := verbose >= 2
	if showTransactions {
		maxDepth = 1<<31 - 1 // unlimited
	}

	// Measure both trees to find a common left width.
	ep := TreePrinter{W: os.Stdout, MaxDepth: maxDepth, ShowTx: showTransactions, ByCategory: expensesByCategory}
	ip := TreePrinter{W: os.Stdout, MaxDepth: maxDepth, ShowTx: showTransactions, ByCategory: incomeByCategory}
	leftWidth := ep.Measure(expenses.Groups)
	if w := ip.Measure(income.Groups); w > leftWidth {
		leftWidth = w
	}
	if leftWidth+18 > termWidth {
		leftWidth = termWidth - 18
	}
	ep.LeftWidth = leftWidth
	ip.LeftWidth = leftWidth

	if len(expenses.Groups) > 0 {
		fmt.Println("Expenses")
		ep.Print(expenses.Groups)
	}
	if len(income.Groups) > 0 {
		if len(expenses.Groups) > 0 {
			fmt.Println()
		}
		fmt.Println("Income")
		ip.Print(income.Groups)
	}

	if edit && len(unknowns) > 0 {
		tree := BuildCategoryTree(prefixes)
		m := NewEditorModel(unknowns, prefixes, tree)
		result, err := tea.NewProgram(m).Run()
		if err != nil {
			return err
		}
		newPrefixes := result.(EditorModel).Added
		if len(newPrefixes) > 0 {
			all := slices.Concat(prefixes, newPrefixes)
			if err := categorise.SavePrefixes(prefixesPath, all); err != nil {
				return err
			}
			fmt.Printf("Added %d new prefix(es) to %s\n", len(newPrefixes), prefixesPath)
		}
	}

	return nil
}

// RunCategoryEditor launches the interactive category editor.
func RunCategoryEditor(prefixes []categorise.Prefix, prefixesPath string) error {
	m := NewCategoriesModel(prefixes)
	result, err := tea.NewProgram(m).Run()
	if err != nil {
		return err
	}

	cm := result.(CategoriesModel)
	if cm.Changed {
		if err := categorise.SavePrefixes(prefixesPath, cm.Prefixes); err != nil {
			return err
		}
		fmt.Printf("Saved changes to %s\n", prefixesPath)
	}

	return nil
}
