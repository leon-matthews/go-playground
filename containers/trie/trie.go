package main

type Trie struct {
	children [256]*Trie
	isEnd    bool
	pattern  string
}

func NewTrie() *Trie {
	return &Trie{}
}

func (t *Trie) Insert(pattern string) {
	node := t
	for i := 0; i < len(pattern); i++ {
		b := pattern[i]
		if node.children[b] == nil {
			node.children[b] = &Trie{}
		}
		node = node.children[b]
	}
	node.isEnd = true
	node.pattern = pattern
}

func (t *Trie) MatchPrefix(input string) (string, bool) {
	return t.matchPrefix(input, false)
}

func (t *Trie) MatchLongestPrefix(input string) (string, bool) {
	return t.matchPrefix(input, true)
}

func (t *Trie) matchPrefix(input string, longest bool) (string, bool) {
	node := t
	last := ""
	for i := 0; i < len(input); i++ {
		b := input[i]
		if node.children[b] == nil {
			break
		}
		node = node.children[b]
		if node.isEnd {
			if !longest {
				return node.pattern, true
			}
			last = node.pattern
		}
	}
	return last, last != ""
}