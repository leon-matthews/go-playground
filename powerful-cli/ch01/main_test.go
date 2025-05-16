package main

import (
	"bytes"
	"testing"
)

func TestCountLines(t *testing.T) {
	b := bytes.NewBufferString("word1, word2, word3\nline2\nline3 word4")
	want := 3
	got := count(b, true)
	if want != got {
		t.Errorf("Expected %d lines, got %d", want, got)
	}
}

func TestCountWords(t *testing.T) {
	b := bytes.NewBufferString("word1, word2, word3 word4\n")
	want := 4
	got := count(b, false)
	if want != got {
		t.Errorf("Expected %d words, got %d", want, got)
	}
}
