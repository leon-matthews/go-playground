package trie

import (
	"slices"
	"testing"
)

// --- Contract parity: HasPrefixMatch & Insert basics ---

func TestRadixHasPrefixMatch_SinglePattern(t *testing.T) {
	rt := NewRadixTrie()
	rt.Insert("foo")

	if !rt.HasPrefixMatch("foobar") {
		t.Error("expected true for input with a matching prefix")
	}
}

func TestRadixHasPrefixMatch_ExactMatch(t *testing.T) {
	rt := NewRadixTrie()
	rt.Insert("foo")

	if !rt.HasPrefixMatch("foo") {
		t.Error("expected true when input exactly equals a pattern")
	}
}

func TestRadixHasPrefixMatch_NoMatch(t *testing.T) {
	rt := NewRadixTrie()
	rt.Insert("foo")

	if rt.HasPrefixMatch("bar") {
		t.Error("expected false when no pattern is a prefix")
	}
}

func TestRadixHasPrefixMatch_PartialInputNoMatch(t *testing.T) {
	rt := NewRadixTrie()
	rt.Insert("foobar")

	// Input is a prefix of the pattern, but the pattern never ends.
	if rt.HasPrefixMatch("foo") {
		t.Error("expected false when the pattern never ends within the input")
	}
}

func TestRadixHasPrefixMatch_StopsAtShortest(t *testing.T) {
	rt := NewRadixTrie()
	rt.Insert("fo")
	rt.Insert("foo")
	rt.Insert("foobar")

	// "fo" already satisfies existence; the deeper patterns don't change it.
	if !rt.HasPrefixMatch("foobar") {
		t.Error("expected true once any prefix matches")
	}
}

func TestRadixHasPrefixMatch_EmptyTrie(t *testing.T) {
	rt := NewRadixTrie()

	if rt.HasPrefixMatch("anything") {
		t.Error("expected false from an empty trie")
	}
}

func TestRadixHasPrefixMatch_EmptyInput(t *testing.T) {
	rt := NewRadixTrie()
	rt.Insert("foo")

	if rt.HasPrefixMatch("") {
		t.Error("expected false for empty input")
	}
}

func TestRadixInsert_EmptyPatternRejected(t *testing.T) {
	rt := NewRadixTrie()
	rt.Insert("")

	// Insert("") is a no-op, so "" never becomes a stored pattern.
	if rt.HasPrefixMatch("anything") {
		t.Error("HasPrefixMatch: expected false after Insert(\"\")")
	}
	if got, ok := rt.MatchLongestPrefix("anything"); ok || got != "" {
		t.Errorf("MatchLongestPrefix: expected (\"\", false), got (%q, %v)", got, ok)
	}
}

func TestRadixMatchLongestPrefix_BinaryBytes(t *testing.T) {
	rt := NewRadixTrie()
	pattern := string([]byte{0x01, 0x02, 0x03})
	rt.Insert(pattern)

	input := string([]byte{0x01, 0x02, 0x03, 0x04})
	got, ok := rt.MatchLongestPrefix(input)
	if !ok || got != pattern {
		t.Errorf("expected (%q, true), got (%q, %v)", pattern, got, ok)
	}
}

// --- MatchLongestPrefix: each Insert situation plus the radix-only subtleties ---

func TestRadixMatchLongestPrefix_Table(t *testing.T) {
	tests := []struct {
		name     string
		patterns []string
		input    string
		wantPat  string
		wantOK   bool
	}{
		{"leaf only", []string{"car"}, "carpet", "car", true},
		{"descend then leaf", []string{"car", "card"}, "cards", "card", true},
		{"split with branch", []string{"car", "cat"}, "cat", "cat", true},
		{"split, shorter is branch end", []string{"card", "car"}, "cargo", "car", true},
		{"split parent is not a word", []string{"car", "cat"}, "ca", "", false},
		{"mid-edge divergence", []string{"card"}, "carrot", "", false},
		{"longest among nested", []string{"a", "ab", "abc"}, "abcd", "abc", true},
		{"exact match", []string{"foo", "foobar"}, "foobar", "foobar", true},
		{"no match", []string{"foo"}, "bar", "", false},
		{"empty input", []string{"foo"}, "", "", false},
		{"disjoint patterns", []string{"abc", "xyz"}, "xyzzy", "xyz", true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rt := NewRadixTrie()
			for _, p := range tc.patterns {
				rt.Insert(p)
			}

			got, ok := rt.MatchLongestPrefix(tc.input)
			if ok != tc.wantOK || got != tc.wantPat {
				t.Errorf("MatchLongestPrefix(%q) with %v: got (%q, %v), want (%q, %v)",
					tc.input, tc.patterns, got, ok, tc.wantPat, tc.wantOK)
			}
		})
	}
}

// --- Insert: duplicates, draining onto a branch, order independence ---

func TestRadixInsert_Duplicate(t *testing.T) {
	rt := NewRadixTrie()
	rt.Insert("foo")
	rt.Insert("foo") // idempotent: drains onto the existing end node

	if got, ok := rt.MatchLongestPrefix("foobar"); !ok || got != "foo" {
		t.Errorf("MatchLongestPrefix(\"foobar\"): got (%q, %v), want (\"foo\", true)", got, ok)
	}
}

func TestRadixInsert_DrainOntoExistingBranch(t *testing.T) {
	rt := NewRadixTrie()
	rt.Insert("test")
	rt.Insert("team") // splits at "te"; the "te" node is a pure branch, not a word
	rt.Insert("te")   // drains onto that branch, turning it into a stored word

	if got, ok := rt.MatchLongestPrefix("tea"); !ok || got != "te" {
		t.Errorf("MatchLongestPrefix(\"tea\"): got (%q, %v), want (\"te\", true)", got, ok)
	}
	if got, ok := rt.MatchLongestPrefix("tester"); !ok || got != "test" {
		t.Errorf("MatchLongestPrefix(\"tester\"): got (%q, %v), want (\"test\", true)", got, ok)
	}
}

func TestRadixInsert_OrderIndependent(t *testing.T) {
	patterns := []string{
		"romane", "romanus", "romulus", "rubens", "ruber", "rubicon", "rubicundus", "rom",
	}

	forward := NewRadixTrie()
	for _, p := range patterns {
		forward.Insert(p)
	}
	reverse := NewRadixTrie()
	for i := len(patterns) - 1; i >= 0; i-- {
		reverse.Insert(patterns[i])
	}

	forwardSeen := make(map[string]bool)
	reverseSeen := make(map[string]bool)
	validateRadix(t, forward, "", forwardSeen)
	validateRadix(t, reverse, "", reverseSeen)

	for _, p := range patterns {
		if !forwardSeen[p] {
			t.Errorf("forward insert order: pattern %q missing", p)
		}
		if !reverseSeen[p] {
			t.Errorf("reverse insert order: pattern %q missing", p)
		}
	}
	if len(forwardSeen) != len(patterns) || len(reverseSeen) != len(patterns) {
		t.Errorf("stored pattern counts differ: forward=%d reverse=%d want=%d",
			len(forwardSeen), len(reverseSeen), len(patterns))
	}
}

// --- Structure invariants ---

func TestRadixInsert_SplitStructure(t *testing.T) {
	patterns := []string{"car", "card", "care", "cat"}

	rt := NewRadixTrie()
	for _, p := range patterns {
		rt.Insert(p)
	}

	seen := make(map[string]bool)
	validateRadix(t, rt, "", seen)

	if len(seen) != len(patterns) {
		t.Errorf("stored %d patterns, want %d", len(seen), len(patterns))
	}
	for _, p := range patterns {
		if !seen[p] {
			t.Errorf("pattern %q missing after splits", p)
		}
	}
}

// validateRadix walks the trie asserting the radix invariants, recording stored patterns in seen.
func validateRadix(t *testing.T, node *RadixTrie, path string, seen map[string]bool) {
	t.Helper()
	if node.isEnd {
		if node.pattern != path {
			t.Errorf("end node pattern %q does not match its root path %q", node.pattern, path)
		}
		seen[node.pattern] = true
	}
	for key, e := range node.children {
		switch {
		case e.label == "":
			t.Errorf("empty edge label at path %q", path)
		case e.label[0] != key:
			t.Errorf("edge keyed %q has label %q (first byte %q)", key, e.label, e.label[0])
		default:
			validateRadix(t, e.child, path+e.label, seen)
		}
	}
}

// --- Helpers ---

func TestCommonPrefix(t *testing.T) {
	tests := []struct {
		a, b, want string
	}{
		{"", "", ""},
		{"abc", "", ""},
		{"", "abc", ""},
		{"abc", "abd", "ab"},
		{"abc", "abc", "abc"},
		{"abc", "abcd", "abc"},
		{"xyz", "abc", ""},
	}

	for _, tc := range tests {
		if got := commonPrefix(tc.a, tc.b); got != tc.want {
			t.Errorf("commonPrefix(%q, %q) = %q, want %q", tc.a, tc.b, got, tc.want)
		}
	}
}

// --- KeysWithPrefix ---

func TestRadixKeysWithPrefix(t *testing.T) {
	patterns := []string{"car", "card", "care", "cat", "dog"}

	tests := []struct {
		prefix string
		want   []string
	}{
		{"car", []string{"car", "card", "care"}},
		{"ca", []string{"car", "card", "care", "cat"}},
		{"c", []string{"car", "card", "care", "cat"}}, // prefix ends mid-edge
		{"card", []string{"card"}},
		{"care", []string{"care"}},
		{"d", []string{"dog"}},
		{"dog", []string{"dog"}},
		{"do", []string{"dog"}},                             // prefix ends mid-edge on the "dog" leaf
		{"", []string{"car", "card", "care", "cat", "dog"}}, // empty prefix matches all
		{"x", nil},
		{"cb", nil},  // diverges within the "ca" edge
		{"cxy", nil}, // diverges within the "ca" edge, prefix longer than the label
		{"cax", nil},
		{"cared", nil}, // prefix longer than any stored key
	}

	rt := NewRadixTrie()
	for _, p := range patterns {
		rt.Insert(p)
	}

	for _, tc := range tests {
		got := rt.KeysWithPrefix(tc.prefix)
		if !slices.Equal(got, tc.want) {
			t.Errorf("KeysWithPrefix(%q) = %v, want %v", tc.prefix, got, tc.want)
		}
	}
}
