package trie

import (
	"math/rand/v2"
	"testing"
)

const benchWordCount = 10_000

// makeWords deterministically builds count pseudo-random lowercase words of length 3..10.
func makeWords(count int) []string {
	rng := rand.New(rand.NewPCG(1, 2))
	words := make([]string, count)
	for i := range words {
		word := make([]byte, 3+rng.IntN(8))
		for j := range word {
			word[j] = byte('a' + rng.IntN(26))
		}
		words[i] = string(word)
	}
	return words
}

// BenchmarkInsert compares building each trie from the same word set.
func BenchmarkInsert(b *testing.B) {
	words := makeWords(benchWordCount)

	b.Run("Trie", func(b *testing.B) {
		b.ReportAllocs()
		var tr *Trie
		for b.Loop() {
			tr = NewTrie()
			for _, w := range words {
				tr.Insert(w)
			}
		}
		if !tr.HasPrefixMatch(words[0]) {
			b.Fatal("built Trie is missing a known word")
		}
	})

	b.Run("RadixTrie", func(b *testing.B) {
		b.ReportAllocs()
		var rt *RadixTrie
		for b.Loop() {
			rt = NewRadixTrie()
			for _, w := range words {
				rt.Insert(w)
			}
		}
		if !rt.HasPrefixMatch(words[0]) {
			b.Fatal("built RadixTrie is missing a known word")
		}
	})
}

// BenchmarkMatchLongestPrefix compares point lookups against a pre-built trie.
func BenchmarkMatchLongestPrefix(b *testing.B) {
	words := makeWords(benchWordCount)

	tr := NewTrie()
	rt := NewRadixTrie()
	for _, w := range words {
		tr.Insert(w)
		rt.Insert(w)
	}

	b.Run("Trie", func(b *testing.B) {
		var match string
		i := 0
		for b.Loop() {
			match, _ = tr.MatchLongestPrefix(words[i%len(words)])
			i++
		}
		if match == "" {
			b.Fatal("expected a match")
		}
	})

	b.Run("RadixTrie", func(b *testing.B) {
		var match string
		i := 0
		for b.Loop() {
			match, _ = rt.MatchLongestPrefix(words[i%len(words)])
			i++
		}
		if match == "" {
			b.Fatal("expected a match")
		}
	})
}

// BenchmarkKeysWithPrefix compares autocomplete enumeration against a pre-built trie.
func BenchmarkKeysWithPrefix(b *testing.B) {
	words := makeWords(benchWordCount)

	tr := NewTrie()
	rt := NewRadixTrie()
	for _, w := range words {
		tr.Insert(w)
		rt.Insert(w)
	}

	// Two-byte prefixes taken from the words, so every query matches at least one key.
	prefixes := make([]string, 0, 256)
	for i := 0; i < len(words) && len(prefixes) < 256; i += len(words) / 256 {
		prefixes = append(prefixes, words[i][:2])
	}

	b.Run("Trie", func(b *testing.B) {
		b.ReportAllocs()
		var keys []string
		i := 0
		for b.Loop() {
			keys = tr.KeysWithPrefix(prefixes[i%len(prefixes)])
			i++
		}
		if len(keys) == 0 {
			b.Fatal("expected at least one key")
		}
	})

	b.Run("RadixTrie", func(b *testing.B) {
		b.ReportAllocs()
		var keys []string
		i := 0
		for b.Loop() {
			keys = rt.KeysWithPrefix(prefixes[i%len(prefixes)])
			i++
		}
		if len(keys) == 0 {
			b.Fatal("expected at least one key")
		}
	})
}
