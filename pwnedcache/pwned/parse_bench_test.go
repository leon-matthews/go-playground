package pwned_test

import (
	"fmt"
	"strings"
	"testing"

	"pwnedcache/pwned"
)

// BenchmarkParseHashList parses a realistically sized list so allocs/op shows
// whether per-line allocation has been eliminated: the figure should stay flat
// regardless of line count, not grow with it.
func BenchmarkParseHashList(b *testing.B) {
	const lines = 800
	var sb strings.Builder
	for i := range lines {
		fmt.Fprintf(&sb, "%035X:%d\r\n", i, i+1)
	}
	list := pwned.HashList(sb.String())

	b.ReportAllocs()
	for b.Loop() {
		hashes, err := pwned.ParseHashList("cafe5", list)
		if err != nil {
			b.Fatal(err)
		}
		if len(hashes) != lines {
			b.Fatalf("got %d hashes, want %d", len(hashes), lines)
		}
	}
}
