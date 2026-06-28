package main

type RadixTrie struct {
	children map[string]*RadixTrie
	isEnd    bool
	pattern  string
}

func NewRadixTrie() *RadixTrie {
	return &RadixTrie{children: make(map[string]*RadixTrie)}
}

func (t *RadixTrie) Insert(pattern string) {
	node := t
	remaining := pattern

	for remaining != "" {
		matched := false
		for edge, child := range node.children {
			prefix := commonPrefix(edge, remaining)
			if prefix == "" {
				continue
			}

			if prefix == edge {
				// Entire edge label matched — descend and consume it
				node = child
				remaining = remaining[len(edge):]
				matched = true
				break
			}

			// Partial match — split the edge
			//
			// Before:  node --"edge"--> child
			// After:   node --"prefix"--> mid --"edge[len(prefix):]"--> child
			//                              └---"remaining[len(prefix):]"--> new leaf
			mid := &RadixTrie{children: make(map[string]*RadixTrie)}
			mid.children[edge[len(prefix):]] = child
			delete(node.children, edge)
			node.children[prefix] = mid

			suffix := remaining[len(prefix):]
			if suffix == "" {
				mid.isEnd = true
				mid.pattern = pattern
			} else {
				leaf := &RadixTrie{
					children: make(map[string]*RadixTrie),
					isEnd:    true,
					pattern:  pattern,
				}
				mid.children[suffix] = leaf
			}
			return
		}

		if !matched {
			// No edge matched at all — add a new leaf
			node.children[remaining] = &RadixTrie{
				children: make(map[string]*RadixTrie),
				isEnd:    true,
				pattern:  pattern,
			}
			return
		}
	}

	node.isEnd = true
	node.pattern = pattern
}

func (t *RadixTrie) MatchLongestPrefix(input string) (string, bool) {
	node := t
	remaining := input
	last := ""

	for remaining != "" {
		matched := false
		for edge, child := range node.children {
			if len(edge) > len(remaining) {
				// Edge longer than remaining input — check for partial match
				if hasPrefix(remaining, edge[:len(remaining)]) {
					// Input is exhausted inside this edge — no complete match here
					return last, last != ""
				}
				continue
			}
			if hasPrefix(remaining, edge) {
				node = child
				remaining = remaining[len(edge):]
				if node.isEnd {
					last = node.pattern
				}
				matched = true
				break
			}
		}
		if !matched {
			break
		}
	}

	return last, last != ""
}

func commonPrefix(a, b string) string {
	i := 0
	for i < len(a) && i < len(b) && a[i] == b[i] {
		i++
	}
	return a[:i]
}

func hasPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}