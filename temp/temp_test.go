package main_test

import (
	"strings"
	"testing"
)

var loremIpsum = `Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed` +
	`do eiusmod tempor incididunt ut labore et dolore magna aliqua. `

func BenchmarkStringSplit(b *testing.B) {
	for b.Loop() {
		for _, w := range strings.Split(loremIpsum, " ") {
			if w == "adipiscing" {
				break
			}
		}
	}
}

func BenchmarkStringSplitSeq(b *testing.B) {
	for b.Loop() {
		for w := range strings.SplitSeq(loremIpsum, " ") {
			if w == "adipiscing" {
				break
			}
		}
	}
}

func BenchmarkStringFields(b *testing.B) {
	for b.Loop() {
		for _, w := range strings.Fields(loremIpsum) {
			if w == "adipiscing" {
				break
			}
		}
	}
}

func BenchmarkStringFieldsSeq(b *testing.B) {
	for b.Loop() {
		for w := range strings.FieldsSeq(loremIpsum) {
			if w == "adipiscing" {
				break
			}
		}
	}
}
