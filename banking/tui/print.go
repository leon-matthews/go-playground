package tui

import (
	"fmt"
	"io"
	"strings"

	"banking/categorise"
	"banking/common"
)

// TreePrinter prints categorised summaries as indented trees with aligned amounts.
type TreePrinter struct {
	W          io.Writer
	MaxDepth   int
	ShowTx     bool
	ByCategory map[string][]*common.Transaction
	LeftWidth  int
}

// Measure calculates the natural width of the left side of the tree output.
func (p *TreePrinter) Measure(groups []*categorise.Group) int {
	return p.measure(groups, 0, "")
}

func (p *TreePrinter) measure(groups []*categorise.Group, depth int, pathPrefix string) int {
	if depth >= p.MaxDepth {
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

		if p.ShowTx && len(g.Children) == 0 {
			txIndent := 2 * (depth + 1)
			for _, t := range p.ByCategory[path] {
				w := txIndent + len(t.Date.Format(common.DateFormat)) + 2 + len(t.Account) + 2 + len(t.Details)
				if w > maxWidth {
					maxWidth = w
				}
			}
		}

		if w := p.measure(g.Children, depth+1, path); w > maxWidth {
			maxWidth = w
		}
	}
	return maxWidth
}

// Print prints the categorised summary as an indented tree with aligned amounts.
func (p *TreePrinter) Print(groups []*categorise.Group) {
	p.print(groups, 0, "")
}

func (p *TreePrinter) print(groups []*categorise.Group, depth int, pathPrefix string) {
	if depth >= p.MaxDepth {
		return
	}
	indent := strings.Repeat("  ", depth)
	for _, g := range groups {
		path := g.Name
		if pathPrefix != "" {
			path = pathPrefix + "/" + g.Name
		}
		label := fmt.Sprintf("%s%s", indent, g.Name)
		fmt.Fprintf(p.W, "%-*s %10.2f (%d)\n", p.LeftWidth, label, g.Total, g.Count())

		if p.ShowTx && len(g.Children) == 0 {
			txIndent := strings.Repeat("  ", depth+1)
			for _, t := range p.ByCategory[path] {
				prefix := fmt.Sprintf("%s%s  %s  ", txIndent, t.Date.Format(common.DateFormat), t.Account)
				maxDetails := p.LeftWidth - len(prefix)
				details := t.Details
				if maxDetails > 3 && len(details) > maxDetails {
					details = details[:maxDetails-3] + "..."
				}
				left := prefix + details
				fmt.Fprintf(p.W, "%-*s %10.2f\n", p.LeftWidth, left, t.Amount)
			}
		}

		p.print(g.Children, depth+1, path)
	}
}
