package tui

import (
	"fmt"
	"strings"

	"banking/categorise"
	"banking/common"
)

// MeasureTree calculates the natural width of the left side of the tree output.
func MeasureTree(groups []*categorise.Group, depth, maxDepth int, showTx bool, byCategory map[string][]*common.Transaction, pathPrefix string) int {
	if depth >= maxDepth {
		return 0
	}
	maxWidth := 0
	indent := 2 * depth
	for _, g := range groups {
		path := g.Name
		if pathPrefix != "" {
			path = pathPrefix + "/" + g.Name
		}
		if w := indent + len(g.Name); w > maxWidth {
			maxWidth = w
		}

		if showTx && len(g.Children) == 0 {
			txIndent := 2 * (depth + 1)
			for _, t := range byCategory[path] {
				w := txIndent + len(t.Date.Format(common.DateFormat)) + 2 + len(t.Account) + 2 + len(t.Details)
				if w > maxWidth {
					maxWidth = w
				}
			}
		}

		if w := MeasureTree(g.Children, depth+1, maxDepth, showTx, byCategory, path); w > maxWidth {
			maxWidth = w
		}
	}
	return maxWidth
}

// PrintTree prints the categorised summary as an indented tree with aligned amounts.
func PrintTree(groups []*categorise.Group, depth, maxDepth int, showTx bool, byCategory map[string][]*common.Transaction, pathPrefix string, leftWidth int) {
	if depth >= maxDepth {
		return
	}
	indent := strings.Repeat("  ", depth)
	for _, g := range groups {
		path := g.Name
		if pathPrefix != "" {
			path = pathPrefix + "/" + g.Name
		}
		label := fmt.Sprintf("%s%s", indent, g.Name)
		fmt.Printf("%-*s %10.2f (%d)\n", leftWidth, label, g.Total, g.Count())

		if showTx && len(g.Children) == 0 {
			txIndent := strings.Repeat("  ", depth+1)
			for _, t := range byCategory[path] {
				prefix := fmt.Sprintf("%s%s  %s  ", txIndent, t.Date.Format(common.DateFormat), t.Account)
				maxDetails := leftWidth - len(prefix)
				details := t.Details
				if maxDetails > 3 && len(details) > maxDetails {
					details = details[:maxDetails-3] + "..."
				}
				left := prefix + details
				fmt.Printf("%-*s %10.2f\n", leftWidth, left, t.Amount)
			}
		}

		PrintTree(g.Children, depth+1, maxDepth, showTx, byCategory, path, leftWidth)
	}
}
