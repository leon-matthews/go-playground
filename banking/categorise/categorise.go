package categorise

import (
	"slices"
	"strings"

	"banking/common"
)

const Unknown = "UNKNOWN"

// Matcher holds a sorted set of prefixes and supports longest-match lookup.
type Matcher struct {
	prefixes []common.Prefix
}

// NewMatcher creates a Matcher from the given prefixes, sorting them for
// binary search. If the prefixes are already sorted, the sort is skipped.
func NewMatcher(prefixes []common.Prefix) *Matcher {
	sorted := make([]common.Prefix, len(prefixes))
	copy(sorted, prefixes)
	if !slices.IsSortedFunc(sorted, common.ComparePrefix) {
		slices.SortFunc(sorted, common.ComparePrefix)
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
	i, found := slices.BinarySearchFunc(m.prefixes, lower, func(p common.Prefix, target string) int {
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
