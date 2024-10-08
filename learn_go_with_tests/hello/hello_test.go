package main

import "testing"

func assertCorrectMessage(t testing.TB, got, want string) {
	t.Helper()
	if got != want {
		t.Errorf("got %q want %q", got, want)
	}
}

func TestHello(t *testing.T) {
	t.Run("passing name as a parameter", func(t *testing.T) {
		got := Hello("Leon", "")
		want := "Hello, Leon"
		assertCorrectMessage(t, got, want)
	})

	t.Run("using default if name empty", func(t *testing.T) {
		got := Hello("", "")
		want := "Hello, world"
		assertCorrectMessage(t, got, want)
	})

	t.Run("in Spanish", func(t *testing.T) {
		got := Hello("León", "Spanish")
		want := "Hola, León"
		assertCorrectMessage(t, got, want)
	})

	t.Run("in French", func(t *testing.T) {
		got := Hello("Leon", "French")
		want := "Bonjour, Leon"
		assertCorrectMessage(t, got, want)
	})
}
