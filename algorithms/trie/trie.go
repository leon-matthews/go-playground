// Package trie provides prefix matching of inputs against a stored set of byte patterns.
package trie

import "slices"

// Trie is a prefix tree with one node per byte: every stored pattern is a root-to-node path.
type Trie struct {
	// children maps a present byte to its child node; a missing key means there is no such child.
	children map[byte]*Trie
	isEnd    bool
	pattern  string
}

// NewTrie returns an empty Trie, ready for Insert.
func NewTrie() *Trie {
	return &Trie{children: make(map[byte]*Trie)}
}

// Insert adds pattern to the trie; inserting the empty string is a no-op.
func (t *Trie) Insert(pattern string) {
	// "" is a prefix of every input, so it is not a valid stored pattern.
	if pattern == "" {
		return
	}
	node := t
	// Walk one byte at a time, creating any missing node along the path.
	for i := 0; i < len(pattern); i++ {
		b := pattern[i]
		child, ok := node.children[b]
		if !ok {
			child = NewTrie()
			node.children[b] = child
		}
		node = child
	}
	// The node where pattern runs out is its terminal node.
	node.isEnd = true
	node.pattern = pattern
}

// MatchLongestPrefix returns the longest stored pattern prefixing input, and whether one was found.
func (t *Trie) MatchLongestPrefix(input string) (string, bool) {
	node := t
	last := "" // longest stored pattern that has fully matched a prefix of input so far
	for i := 0; i < len(input); i++ {
		b := input[i]
		// A single map lookup selects the next child; a missing key ends the path here.
		child, ok := node.children[b]
		if !ok {
			break
		}
		node = child
		if node.isEnd {
			last = node.pattern // a complete pattern ends here; the deepest one wins
		}
	}
	return last, last != ""
}

// HasPrefixMatch reports whether any stored pattern is a prefix of input.
func (t *Trie) HasPrefixMatch(input string) bool {
	node := t
	for i := 0; i < len(input); i++ {
		b := input[i]
		child, ok := node.children[b]
		if !ok {
			break
		}
		node = child
		if node.isEnd {
			return true // first complete pattern is enough — no need to find the longest
		}
	}
	return false
}

// KeysWithPrefix returns every stored pattern beginning with prefix, in lexicographic order.
// An empty prefix returns all stored patterns; an unmatched prefix returns nil.
func (t *Trie) KeysWithPrefix(prefix string) []string {
	node := t
	// Navigate to the node where prefix ends.
	for i := 0; i < len(prefix); i++ {
		child, ok := node.children[prefix[i]]
		if !ok {
			return nil // no stored pattern continues this prefix
		}
		node = child
	}
	// Collect every stored pattern in the subtree rooted there.
	var keys []string
	node.collectKeys(&keys)
	slices.Sort(keys)
	return keys
}

// collectKeys appends every stored pattern in this subtree to keys, in no particular order.
func (t *Trie) collectKeys(keys *[]string) {
	if t.isEnd {
		*keys = append(*keys, t.pattern)
	}
	for _, child := range t.children {
		child.collectKeys(keys)
	}
}
