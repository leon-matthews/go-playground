package trie_test

import (
	"fmt"

	"local.dev/trie"
)

// ExampleTrie_MatchLongestPrefix routes a phone number to its most specific dialing prefix.
func ExampleTrie_MatchLongestPrefix() {
	// A dial plan keyed by international prefix; the longest match is the most specific.
	regions := map[string]string{
		"1":   "North America",
		"44":  "United Kingdom",
		"64":  "New Zealand",
		"644": "Wellington",
	}
	dial := trie.NewTrie()
	for prefix := range regions {
		dial.Insert(prefix)
	}

	number := "6444719000"
	prefix, _ := dial.MatchLongestPrefix(number)
	fmt.Printf("%s -> %s (%s)\n", number, prefix, regions[prefix])
	// Output:
	// 6444719000 -> 644 (Wellington)
}

// ExampleTrie_HasPrefixMatch guards any request path that falls under a protected route.
func ExampleTrie_HasPrefixMatch() {
	protected := trie.NewTrie()
	protected.Insert("/admin")
	protected.Insert("/internal")

	// An exact match (here "/internal") counts as a prefix match too.
	for _, path := range []string{"/admin/users", "/public/docs", "/internal"} {
		fmt.Printf("%s: %v\n", path, protected.HasPrefixMatch(path))
	}
	// Output:
	// /admin/users: true
	// /public/docs: false
	// /internal: true
}

// ExampleRadixTrie_MatchLongestPrefix shows the RadixTrie as a drop-in with identical semantics.
func ExampleRadixTrie_MatchLongestPrefix() {
	// The RadixTrie compresses the long shared "/api/v1/" prefix into a single edge.
	routes := trie.NewRadixTrie()
	routes.Insert("/api/v1/users")
	routes.Insert("/api/v1/orders")

	route, ok := routes.MatchLongestPrefix("/api/v1/users/42")
	fmt.Println(route, ok)
	// Output: /api/v1/users true
}

// ExampleTrie_KeysWithPrefix lists autocomplete suggestions for a typed prefix.
func ExampleTrie_KeysWithPrefix() {
	words := trie.NewTrie()
	for _, w := range []string{"car", "card", "care", "cat", "dog"} {
		words.Insert(w)
	}

	fmt.Println(words.KeysWithPrefix("car"))
	fmt.Println(words.KeysWithPrefix("ca"))
	// Output:
	// [car card care]
	// [car card care cat]
}
