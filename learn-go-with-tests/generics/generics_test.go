package generics_test

import "testing"

func TestAssertHelpers(t *testing.T) {
	t.Run("assert on integers", func(t *testing.T) {
		AssertEqual(t, 1, 1)
		AssertNotEqual(t, 1, 2)
	})

	t.Run("assert on strings", func(t *testing.T) {
		AssertEqual(t, "hello", "hello")
		AssertNotEqual(t, "hello", "Hello")
	})
}

func AssertEqual[C comparable](t *testing.T, want, got C) {
	if want != got {
		t.Errorf("want %+v, got: %+v", want, got)
	}
}

func AssertNotEqual[C comparable](t *testing.T, want, got C) {
	if want == got {
		t.Errorf("didn't want %+v", got)
	}
}
