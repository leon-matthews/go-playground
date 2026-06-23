package main

import (
	"testing"

	wordwrap "github.com/mitchellh/go-wordwrap"
)

// A long string to simulate real-world text processing
const largeInput = "Go is an open source programming language that makes it easy to build simple, reliable, and efficient software. It was designed at Google to improve programming productivity in an era of multicore, networked machines, and large codebases."

// equivalenceCases exercise the behaviours WrapString shares with go-wordwrap.
var equivalenceCases = []struct {
	name string
	text string
	lim  uint
}{
	{"empty", "", 10},
	{"short", "hello world", 100},
	{"large width 20", largeInput, 20},
	{"large width 40", largeInput, 40},
	{"large width 80", largeInput, 80},
	{"leading and trailing spaces", "   leading and trailing   ", 10},
	{"collapsed runs preserved", "multiple    spaces    between", 12},
	{"embedded newlines", "line one\nline two\nline three", 20},
	{"blank line paragraph", "para one here\n\npara two here", 10},
	{"long word overflows", "supercalifragilisticexpialidocious word", 10},
	{"non-breaking space joins", "non breaking spaces stay together here", 8},
	{"accented latin", "café naïve résumé über fünf", 6},
	{"cjk", "日本語 の テキスト を 折り返す", 4},
	{"tabs", "tabs\tand\tspaces here", 9},
	{"trailing newline", "trailing newline\n", 30},
	{"leading newlines", "\n\n\nleading newlines", 30},
	{"single chars", "a b c d e f g h i j k l m n", 5},
}

// Confirm WrapString matches go-wordwrap on valid UTF-8 input.
func TestWrapStringMatchesReference(t *testing.T) {
	for _, c := range equivalenceCases {
		t.Run(c.name, func(t *testing.T) {
			want := wordwrap.WrapString(c.text, c.lim)
			got := WrapString(c.text, c.lim)
			if got != want {
				t.Errorf("lim=%d\n in=%q\nwant=%q\n got=%q", c.lim, c.text, want, got)
			}
		})
	}
}

// Benchmark the reference go-wordwrap implementation
func BenchmarkReferenceWrap(b *testing.B) {
	for b.Loop() {
		wordwrap.WrapString(largeInput, 20)
	}
}

// Benchmark the single-pass WrapString implementation
func BenchmarkWrapString(b *testing.B) {
	for b.Loop() {
		WrapString(largeInput, 20)
	}
}
