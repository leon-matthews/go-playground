package main

import (
	"os"
	"slices"
	"testing"

	"golang.org/x/net/html"
)

// loadTestHTML reads the sample document the benchmarks run against.
func loadTestHTML(tb testing.TB) string {
	tb.Helper()
	body, err := os.ReadFile("testdata/google.html")
	if err != nil {
		tb.Fatal(err)
	}
	return string(body)
}

func BenchmarkTokenizeLinks(b *testing.B) {
	body := loadTestHTML(b)
	b.ReportAllocs()
	for b.Loop() {
		if _, err := tokenizeLinks(body); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParseLinks(b *testing.B) {
	body := loadTestHTML(b)
	b.ReportAllocs()
	for b.Loop() {
		if _, _, err := parseLinks(body); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkTdewolffLinks(b *testing.B) {
	body := loadTestHTML(b)
	b.ReportAllocs()
	for b.Loop() {
		if _, err := tdewolffLinks(body); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkIndexLinks(b *testing.B) {
	body := loadTestHTML(b)
	b.ReportAllocs()
	for b.Loop() {
		if _, err := indexLinks(body); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRegexpLinks(b *testing.B) {
	body := loadTestHTML(b)
	b.ReportAllocs()
	for b.Loop() {
		if _, err := regexpLinks(body); err != nil {
			b.Fatal(err)
		}
	}
}

// TestLinksAgree checks that all three approaches find the same set of links.
func TestLinksAgree(t *testing.T) {
	body := loadTestHTML(t)

	tokenized, err := tokenizeLinks(body)
	if err != nil {
		t.Fatalf("tokenizeLinks: %v", err)
	}
	parsed, _, err := parseLinks(body)
	if err != nil {
		t.Fatalf("parseLinks: %v", err)
	}
	lexed, err := tdewolffLinks(body)
	if err != nil {
		t.Fatalf("tdewolffLinks: %v", err)
	}

	tokenized = normalize(tokenized)
	for _, other := range []struct {
		name  string
		links []string
	}{
		{"parse", normalize(parsed)},
		{"tdewolff", normalize(lexed)},
	} {
		if !slices.Equal(tokenized, other.links) {
			t.Errorf("tokenize vs %s differ: tokenize found %d, %s found %d\n only in tokenize: %v\n only in %s:    %v",
				other.name, len(tokenized), other.name, len(other.links),
				difference(tokenized, other.links), other.name, difference(other.links, tokenized))
		}
	}
}

// roughLinkTolerance caps how many links a best-effort extractor may miss or invent.
const roughLinkTolerance = 3

// TestRoughLinksWithinTolerance checks the approximate extractors stay close to canonical.
func TestRoughLinksWithinTolerance(t *testing.T) {
	body := loadTestHTML(t)
	canonical, err := tokenizeLinks(body)
	if err != nil {
		t.Fatalf("tokenizeLinks: %v", err)
	}
	canonical = normalize(canonical)

	for _, tc := range []struct {
		name string
		fn   func(string) ([]string, error)
	}{
		{"index", indexLinks},
		{"regexp", regexpLinks},
	} {
		got, err := tc.fn(body)
		if err != nil {
			t.Fatalf("%s: %v", tc.name, err)
		}
		// Decode entities so encoding differences are not counted; indexLinks does not decode.
		for i, link := range got {
			got[i] = html.UnescapeString(link)
		}
		got = normalize(got)
		missed := difference(canonical, got)
		spurious := difference(got, canonical)
		if len(missed)+len(spurious) > roughLinkTolerance {
			t.Errorf("%s disagrees on %d links (tolerance %d)\n missed: %v\n spurious: %v",
				tc.name, len(missed)+len(spurious), roughLinkTolerance, missed, spurious)
		}
	}
}

// normalize returns the links sorted with duplicates removed.
func normalize(links []string) []string {
	links = slices.Clone(links)
	slices.Sort(links)
	return slices.Compact(links)
}

// difference returns the elements of a that are not in b.
func difference(a, b []string) []string {
	var diff []string
	for _, s := range a {
		if !slices.Contains(b, s) {
			diff = append(diff, s)
		}
	}
	return diff
}
