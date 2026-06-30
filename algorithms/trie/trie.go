// Package trie provides prefix matching of inputs against a stored set of byte patterns.
package trie

// Trie is a prefix tree with one node per byte: every stored pattern is a root-to-node path.
type Trie struct {
	// children is a dense table indexed by byte value; a nil entry means there is no such child.
	children [256]*Trie
	isEnd    bool
	pattern  string
}

// NewTrie returns an empty Trie, ready for Insert.
func NewTrie() *Trie {
	return &Trie{}
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
		if node.children[b] == nil {
			node.children[b] = &Trie{}
		}
		node = node.children[b]
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
		// Indexing the dense table picks the next child in O(1); nil means the path ends here.
		if node.children[b] == nil {
			break
		}
		node = node.children[b]
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
		if node.children[b] == nil {
			break
		}
		node = node.children[b]
		if node.isEnd {
			return true // first complete pattern is enough — no need to find the longest
		}
	}
	return false
}
