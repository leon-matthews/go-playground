package categorise

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"

	"banking/common"
)

const Unknown = "UNKNOWN"

// Prefix maps a lowercase merchant prefix to its category.
type Prefix struct {
	Text     string
	Category string
}

// comparePrefix orders prefixes by text for binary search.
func comparePrefix(a, b Prefix) int {
	return strings.Compare(a.Text, b.Text)
}

// LoadPrefixes reads a CSV file of prefix,category pairs.
func LoadPrefixes(path string) ([]Prefix, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open prefixes: %w", err)
	}
	defer f.Close()

	r := csv.NewReader(f)
	var prefixes []Prefix
	for {
		record, err := r.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("read prefixes: %w", err)
		}
		if len(record) != 2 {
			return nil, fmt.Errorf("expected 2 fields, got %d: %v", len(record), record)
		}
		prefixes = append(prefixes, Prefix{
			Text:     strings.ToLower(strings.TrimSpace(record[0])),
			Category: strings.TrimSpace(record[1]),
		})
	}
	return prefixes, nil
}

// SavePrefixes sorts and writes all prefixes to a CSV file.
func SavePrefixes(path string, prefixes []Prefix) error {
	sorted := make([]Prefix, len(prefixes))
	copy(sorted, prefixes)
	slices.SortFunc(sorted, comparePrefix)

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create prefixes: %w", err)
	}
	defer f.Close()

	w := csv.NewWriter(f)
	for _, p := range sorted {
		if err := w.Write([]string{p.Text, p.Category}); err != nil {
			return fmt.Errorf("write prefix: %w", err)
		}
	}
	w.Flush()
	return w.Error()
}

// Matcher holds a sorted set of prefixes and supports longest-match lookup.
type Matcher struct {
	prefixes []Prefix
}

// NewMatcher creates a Matcher from the given prefixes, sorting them for
// binary search. If the prefixes are already sorted, the sort is skipped.
func NewMatcher(prefixes []Prefix) *Matcher {
	sorted := make([]Prefix, len(prefixes))
	copy(sorted, prefixes)
	if !slices.IsSortedFunc(sorted, comparePrefix) {
		slices.SortFunc(sorted, comparePrefix)
	}
	return &Matcher{prefixes: sorted}
}

// Group is a node in the category hierarchy (e.g. "Food", "Groceries").
// It holds child groups and the transactions that fall under it.
type Group struct {
	Name         string
	Total        float64
	Children     []*Group
	Transactions []*common.Transaction
}

// Count returns the number of transactions in this group.
func (g *Group) Count() int {
	return len(g.Transactions)
}

// Summary is a top-level container holding root groups and a flat index
// of transactions by their full category path.
type Summary struct {
	Groups     []*Group
	ByCategory map[string][]*common.Transaction
}

// Sort recursively sorts all groups alphabetically by name.
func (s *Summary) Sort() {
	sortGroups(s.Groups)
}

func sortGroups(groups []*Group) {
	slices.SortFunc(groups, func(a, b *Group) int {
		return strings.Compare(a.Name, b.Name)
	})
	for _, g := range groups {
		sortGroups(g.Children)
	}
}

// Add splits category on "/" and walks/creates groups along the path,
// appending the transaction and accumulating the amount at every level.
func (s *Summary) Add(category string, t *common.Transaction) {
	if s.ByCategory == nil {
		s.ByCategory = make(map[string][]*common.Transaction)
	}
	s.ByCategory[category] = append(s.ByCategory[category], t)

	segments := strings.Split(category, "/")
	groups := &s.Groups
	for _, seg := range segments {
		g := findGroup(*groups, seg)
		if g == nil {
			g = &Group{Name: seg}
			*groups = append(*groups, g)
		}
		g.Total += t.Amount
		g.Transactions = append(g.Transactions, t)
		groups = &g.Children
	}
}

// Summarise categorises transactions into expense and income summaries,
// and collects any transactions that could not be matched.
func Summarise(matcher *Matcher, transactions []*common.Transaction) (expenses, income Summary, unknowns []*common.Transaction) {
	for _, t := range transactions {
		cat := matcher.Match(t.Details)
		if t.Amount < 0 {
			expenses.Add(cat, t)
		} else {
			income.Add(cat, t)
		}
		if cat == Unknown {
			unknowns = append(unknowns, t)
		}
	}
	expenses.Sort()
	income.Sort()
	return
}

func findGroup(groups []*Group, name string) *Group {
	for _, g := range groups {
		if g.Name == name {
			return g
		}
	}
	return nil
}

// Match finds the longest prefix that matches the start of detail.
// Returns Unknown if no prefix matches.
func (m *Matcher) Match(detail string) string {
	lower := strings.ToLower(detail)
	// Find the insertion point for the detail in the sorted prefixes.
	// The longest matching prefix (if any) is at or before that point.
	i, found := slices.BinarySearchFunc(m.prefixes, lower, func(p Prefix, target string) int {
		return strings.Compare(p.Text, target)
	})
	if !found {
		i--
	}
	for ; i >= 0; i-- {
		if strings.HasPrefix(lower, m.prefixes[i].Text) {
			return m.prefixes[i].Category
		}
	}
	return Unknown
}
