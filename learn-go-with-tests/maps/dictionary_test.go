package main

import "testing"

func TestSearch(t *testing.T) {
	dict := Dictionary{"test": "This is just a test"}
	got := Search(dict, "test")
	want := "This is just a test"
	assertStrings(t, got, want)
}

func assertStrings(t testing.TB, got, want string) {
	t.Helper()
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
