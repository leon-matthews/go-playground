package main

import (
	"fmt"
	"testing"
)

func BenchmarkPrefixes(b *testing.B) {
	for b.Loop() {
		for _ = range HexStrings(5) {
		}
	}
}

func BenchmarkLoop(b *testing.B) {
	limit := 0x01 << 20
	for b.Loop() {
		for i := 0; i < limit; i++ {
			fmt.Sprintf("%05x", i)
		}
	}
}
