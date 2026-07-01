package trie

import "slices"

// edge is a labelled link to a child node; the label is the substring consumed along the link.
type edge struct {
	label string
	child *RadixTrie
}

// RadixTrie is a prefix tree whose edges carry whole substrings, compressing non-branching chains.
//
// Compressing chains into labelled edges makes it far leaner and faster than a plain [Trie], at
// the cost of a more intricate insert. Prefer it for large keyspaces or long keys with shared
// prefixes such as paths, URLs, or identifiers. See the package overview for the full comparison.
type RadixTrie struct {
	// children is keyed by each edge's first byte, which the radix invariant keeps unique per node.
	children map[byte]edge
	isEnd    bool
	pattern  string
}

// NewRadixTrie returns an empty RadixTrie, ready for Insert.
func NewRadixTrie() *RadixTrie {
	return &RadixTrie{children: make(map[byte]edge)}
}

// newLeaf builds a terminal node that stores the full pattern ending there.
func newLeaf(pattern string) *RadixTrie {
	leaf := NewRadixTrie()
	leaf.isEnd = true
	leaf.pattern = pattern
	return leaf
}

// Insert adds pattern to the trie; inserting the empty string is a no-op.
func (t *RadixTrie) Insert(pattern string) {
	// "" is a prefix of every input, so it is not a valid stored pattern.
	if pattern == "" {
		return
	}
	node := t
	remaining := pattern // the suffix of pattern still to be placed

	for remaining != "" {
		e, ok := node.children[remaining[0]]
		if !ok {
			// Nothing here starts with this byte - hang the rest of pattern as a new leaf.
			node.children[remaining[0]] = edge{label: remaining, child: newLeaf(pattern)}
			return
		}

		// e shares remaining[0], so the common prefix is always at least one byte long.
		prefix := commonPrefix(e.label, remaining)
		if prefix == e.label {
			// The whole edge matched - consume it and keep walking down.
			node = e.child
			remaining = remaining[len(e.label):]
			continue
		}

		// Input and edge agree on a prefix then diverge - split the edge at that point.
		mid := NewRadixTrie()
		// Re-hang the old edge's tail (everything after the shared prefix) beneath mid.
		mid.children[e.label[len(prefix)]] = edge{label: e.label[len(prefix):], child: e.child}
		// Point the parent at mid through an edge labelled with only the shared prefix.
		node.children[remaining[0]] = edge{label: prefix, child: mid}

		suffix := remaining[len(prefix):]
		if suffix == "" {
			// pattern ends exactly at the split point, so mid itself is an end node.
			mid.isEnd = true
			mid.pattern = pattern
		} else {
			// pattern continues past the split - its tail becomes mid's second child.
			mid.children[suffix[0]] = edge{label: suffix, child: newLeaf(pattern)}
		}
		return
	}

	// remaining drained by following existing edges (e.g. a duplicate or a shorter pattern).
	node.isEnd = true
	node.pattern = pattern
}

// MatchLongestPrefix returns the longest stored pattern prefixing input, and whether one was found.
func (t *RadixTrie) MatchLongestPrefix(input string) (string, bool) {
	node := t
	remaining := input
	last := "" // longest stored pattern that has fully matched a prefix of input so far

	for remaining != "" {
		// At most one edge can start with remaining[0], so the byte selects it in O(1).
		e, ok := node.children[remaining[0]]
		// Stop when no edge shares the byte, or its label isn't wholly present in remaining.
		if !ok || !hasPrefix(remaining, e.label) {
			break
		}
		node = e.child
		remaining = remaining[len(e.label):]
		if node.isEnd {
			last = node.pattern // a complete pattern ends here; the deepest one wins
		}
	}

	return last, last != ""
}

// HasPrefixMatch reports whether any stored pattern is a prefix of input.
func (t *RadixTrie) HasPrefixMatch(input string) bool {
	node := t
	remaining := input

	for remaining != "" {
		e, ok := node.children[remaining[0]]
		if !ok || !hasPrefix(remaining, e.label) {
			break
		}
		node = e.child
		remaining = remaining[len(e.label):]
		if node.isEnd {
			return true // first complete pattern is enough - no need to find the longest
		}
	}

	return false
}

// KeysWithPrefix returns every stored pattern beginning with prefix, in lexicographic order.
// An empty prefix returns all stored patterns; an unmatched prefix returns nil.
func (t *RadixTrie) KeysWithPrefix(prefix string) []string {
	node := t
	remaining := prefix

	for remaining != "" {
		e, ok := node.children[remaining[0]]
		if !ok {
			return nil // no edge continues this prefix
		}
		if len(e.label) >= len(remaining) {
			// Prefix ends within this edge, so it must be a prefix of the label.
			if !hasPrefix(e.label, remaining) {
				return nil
			}
			node = e.child
			remaining = ""
		} else {
			// The whole label must sit at the head of the remaining prefix.
			if !hasPrefix(remaining, e.label) {
				return nil
			}
			node = e.child
			remaining = remaining[len(e.label):]
		}
	}

	var keys []string
	node.collectKeys(&keys)
	slices.Sort(keys)
	return keys
}

// collectKeys appends every stored pattern in this subtree to keys, in no particular order.
func (t *RadixTrie) collectKeys(keys *[]string) {
	if t.isEnd {
		*keys = append(*keys, t.pattern)
	}
	for _, e := range t.children {
		e.child.collectKeys(keys)
	}
}

// commonPrefix returns the longest byte prefix shared by a and b.
func commonPrefix(a, b string) string {
	i := 0
	for i < len(a) && i < len(b) && a[i] == b[i] {
		i++
	}
	return a[:i]
}

// hasPrefix reports whether s begins with prefix (a local strings.HasPrefix, no import needed).
func hasPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}
