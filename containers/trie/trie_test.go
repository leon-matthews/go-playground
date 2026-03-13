package main

import "testing"

// --- Insert & MatchPrefix (shortest) ---

func TestMatchPrefix_SinglePattern(t *testing.T) {
	trie := NewTrie()
	trie.Insert("foo")

	got, ok := trie.MatchPrefix("foobar")
	if !ok || got != "foo" {
		t.Errorf("expected (foo, true), got (%q, %v)", got, ok)
	}
}

func TestMatchPrefix_ExactMatch(t *testing.T) {
	trie := NewTrie()
	trie.Insert("foo")

	got, ok := trie.MatchPrefix("foo")
	if !ok || got != "foo" {
		t.Errorf("expected (foo, true), got (%q, %v)", got, ok)
	}
}

func TestMatchPrefix_NoMatch(t *testing.T) {
	trie := NewTrie()
	trie.Insert("foo")

	got, ok := trie.MatchPrefix("bar")
	if ok || got != "" {
		t.Errorf("expected ('', false), got (%q, %v)", got, ok)
	}
}

func TestMatchPrefix_PartialInputNoMatch(t *testing.T) {
	trie := NewTrie()
	trie.Insert("foobar")

	// input is a prefix of the pattern, but pattern never ends
	got, ok := trie.MatchPrefix("foo")
	if ok || got != "" {
		t.Errorf("expected ('', false), got (%q, %v)", got, ok)
	}
}

func TestMatchPrefix_ReturnsShortest(t *testing.T) {
	trie := NewTrie()
	trie.Insert("fo")
	trie.Insert("foo")
	trie.Insert("foobar")

	// Should return the shortest match "fo", not "foo" or "foobar"
	got, ok := trie.MatchPrefix("foobar")
	if !ok || got != "fo" {
		t.Errorf("expected (fo, true), got (%q, %v)", got, ok)
	}
}

func TestMatchPrefix_EmptyTrie(t *testing.T) {
	trie := NewTrie()

	got, ok := trie.MatchPrefix("anything")
	if ok || got != "" {
		t.Errorf("expected ('', false), got (%q, %v)", got, ok)
	}
}

func TestMatchPrefix_EmptyInput(t *testing.T) {
	trie := NewTrie()
	trie.Insert("foo")

	got, ok := trie.MatchPrefix("")
	if ok || got != "" {
		t.Errorf("expected ('', false), got (%q, %v)", got, ok)
	}
}

func TestMatchPrefix_EmptyPatternInserted(t *testing.T) {
	trie := NewTrie()
	trie.Insert("")

	// The root is immediately an end node, but the loop never runs
	// so last stays "" and the function returns ("", false).
	// This documents the current behaviour.
	got, ok := trie.MatchPrefix("anything")
	if ok || got != "" {
		t.Errorf("expected ('', false), got (%q, %v)", got, ok)
	}
}

// --- MatchLongestPrefix ---

func TestMatchLongestPrefix_ReturnsLongest(t *testing.T) {
	trie := NewTrie()
	trie.Insert("fo")
	trie.Insert("foo")
	trie.Insert("foobar")

	got, ok := trie.MatchLongestPrefix("foobarbaz")
	if !ok || got != "foobar" {
		t.Errorf("expected (foobar, true), got (%q, %v)", got, ok)
	}
}

func TestMatchLongestPrefix_SinglePattern(t *testing.T) {
	trie := NewTrie()
	trie.Insert("foo")

	got, ok := trie.MatchLongestPrefix("foobar")
	if !ok || got != "foo" {
		t.Errorf("expected (foo, true), got (%q, %v)", got, ok)
	}
}

func TestMatchLongestPrefix_NoMatch(t *testing.T) {
	trie := NewTrie()
	trie.Insert("foo")

	got, ok := trie.MatchLongestPrefix("bar")
	if ok || got != "" {
		t.Errorf("expected ('', false), got (%q, %v)", got, ok)
	}
}

func TestMatchLongestPrefix_ExactMatch(t *testing.T) {
	trie := NewTrie()
	trie.Insert("foo")
	trie.Insert("foobar")

	got, ok := trie.MatchLongestPrefix("foobar")
	if !ok || got != "foobar" {
		t.Errorf("expected (foobar, true), got (%q, %v)", got, ok)
	}
}

func TestMatchLongestPrefix_InputShorterThanLongestPattern(t *testing.T) {
	trie := NewTrie()
	trie.Insert("fo")
	trie.Insert("foobar")

	// "foo" can only reach "fo" as a complete match
	got, ok := trie.MatchLongestPrefix("foo")
	if !ok || got != "fo" {
		t.Errorf("expected (fo, true), got (%q, %v)", got, ok)
	}
}

// --- Multiple patterns, disjoint ---

func TestMatchPrefix_DisjointPatterns(t *testing.T) {
	trie := NewTrie()
	trie.Insert("abc")
	trie.Insert("xyz")

	got, ok := trie.MatchPrefix("xyzzy")
	if !ok || got != "xyz" {
		t.Errorf("expected (xyz, true), got (%q, %v)", got, ok)
	}

	got, ok = trie.MatchPrefix("abcdef")
	if !ok || got != "abc" {
		t.Errorf("expected (abc, true), got (%q, %v)", got, ok)
	}
}

// --- Binary / high-byte safety ---

func TestMatchPrefix_BinaryBytes(t *testing.T) {
	trie := NewTrie()
	pattern := string([]byte{0x01, 0x02, 0x03})
	trie.Insert(pattern)

	input := string([]byte{0x01, 0x02, 0x03, 0x04})
	got, ok := trie.MatchPrefix(input)
	if !ok || got != pattern {
		t.Errorf("expected (%q, true), got (%q, %v)", pattern, got, ok)
	}
}

// --- Table-driven broader coverage ---

func TestMatchLongestPrefix_Table(t *testing.T) {
	patterns := []string{"a", "ab", "abc", "abcd"}

	tests := []struct {
		input    string
		wantPat  string
		wantBool bool
	}{
		{"abcde", "abcd", true},
		{"abcX", "abc", true},
		{"abX", "ab", true},
		{"aX", "a", true},
		{"X", "", false},
		{"", "", false},
	}

	trie := NewTrie()
	for _, p := range patterns {
		trie.Insert(p)
	}

	for _, tc := range tests {
		got, ok := trie.MatchLongestPrefix(tc.input)
		if ok != tc.wantBool || got != tc.wantPat {
			t.Errorf("MatchLongestPrefix(%q): expected (%q, %v), got (%q, %v)",
				tc.input, tc.wantPat, tc.wantBool, got, ok)
		}
	}
}
