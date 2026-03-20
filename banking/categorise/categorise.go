package categorise

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"
)

const Unknown = "UNKNOWN"

// Prefix maps a lowercase merchant prefix to its category.
type Prefix struct {
	Text     string
	Category string
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

// Matcher holds a sorted set of prefixes and supports longest-match lookup.
type Matcher struct {
	prefixes []Prefix
}

// NewMatcher creates a Matcher from the given prefixes, sorting them for
// binary search.
func NewMatcher(prefixes []Prefix) *Matcher {
	sorted := make([]Prefix, len(prefixes))
	copy(sorted, prefixes)
	slices.SortFunc(sorted, func(a, b Prefix) int {
		return strings.Compare(a.Text, b.Text)
	})
	return &Matcher{prefixes: sorted}
}

// Group is a node in the category hierarchy (e.g. "Food", "Groceries").
// It holds child groups and the transaction details that fall under it.
type Group struct {
	Name         string
	Children     []*Group
	Transactions []string
}

// Count returns the number of transactions in this group.
func (g *Group) Count() int {
	return len(g.Transactions)
}

// Summary is a top-level container holding root groups.
type Summary struct {
	Groups []*Group
}

// Add splits category on "/" and walks/creates groups along the path,
// appending the detail to every group along the path.
func (s *Summary) Add(category, detail string) {
	segments := strings.Split(category, "/")
	groups := &s.Groups
	for _, seg := range segments {
		g := findGroup(*groups, seg)
		if g == nil {
			g = &Group{Name: seg}
			*groups = append(*groups, g)
		}
		g.Transactions = append(g.Transactions, detail)
		groups = &g.Children
	}
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
