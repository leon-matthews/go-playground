package main

import (
	"bytes"
	"testing"
)

func TestGreet(t *testing.T) {
	buffer := bytes.Buffer{}
	want := "Hello, Leon!"

	Greet(&buffer, "Leon")
	got := buffer.String()

	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
