package main

import "testing"

func TestHello(t *testing.T) {
	t.Run("uses name when provided", func(t *testing.T) {
		got := Hello("Leon", "")
		want := "Hello, Leon"
		assertMessageEqual(t, got, want)
	})

	t.Run("uses default if name empty", func(t *testing.T) {
		got := Hello("", "")
		want := "Hello, world!"
		assertMessageEqual(t, got, want)
	})

	t.Run("uses Spanish if requested", func(t *testing.T) {
		got := Hello("José", "es")
		want := "Hola, José"
		assertMessageEqual(t, got, want)
	})

	t.Run("uses French if requested", func(t *testing.T) {
		got := Hello("Mila", "fr")
		want := "Bonjour, Mila"
		assertMessageEqual(t, got, want)
	})
}

func assertMessageEqual(t testing.TB, got, want string) {
	t.Helper()
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
