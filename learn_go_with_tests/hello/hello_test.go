package main

import "testing"

func TestHello(t *testing.T) {
	got := Hello("Leon")
	want := "Hello, Leon"
	if got != want {
		t.Errorf("got %q want %q", got, want)
	}
}
