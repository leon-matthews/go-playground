package main

import (
	"bytes"
	"testing"
)

func TestGreet(t *testing.T) {
	buffer := bytes.Buffer{}
	Greet(&buffer, "Leon")

	got := buffer.String()
	want := "Hello, Leon!"

	if got != want {
		t.Errorf("got %q want %q", got, want)
	}
}
