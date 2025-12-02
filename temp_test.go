package main

import (
	"strings"
	"testing"
)

var expected = "Sphinx of black quartz, judge my vow!"
var parts = []string{
	"Sphinx",
	"of",
	"black",
	"quartz,",
	"judge",
	"my",
	"vow!",
}

func BenchmarkAppend(b *testing.B) {
	var got string
	for b.Loop() {
		got = ""
		for _, part := range parts {
			got += part
			got += " "
		}
		got = strings.TrimSpace(got)
	}

	if expected != got {
		b.Errorf("Unexpected output: %q", got)
	}
}

func BenchmarkStringBuilder(b *testing.B) {
	var got string
	var sb strings.Builder

	for b.Loop() {
		for _, part := range parts {
			sb.WriteString(part)
			sb.WriteString(" ")

		}
		got = strings.TrimSpace(sb.String())
		sb.Reset()
	}

	if expected != got {
		b.Errorf("Unexpected output: %q", got)
	}
}

func BenchmarkStringsJoin(b *testing.B) {
	var got string
	for b.Loop() {
		got = strings.Join(parts, " ")
	}

	if expected != got {
		b.Errorf("Unexpected output: %q", got)
	}
}
