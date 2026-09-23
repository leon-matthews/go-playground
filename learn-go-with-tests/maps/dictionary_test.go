package main

import (
	"errors"
	"testing"
)

func TestAdd(t *testing.T) {
	t.Run("new word", func(t *testing.T) {
		dict := Dictionary{}
		word := "banana"
		definition := "Things are usually this size"
		err := dict.Add(word, definition)
		assertNil(t, err)

		got, err := dict.Search(word)
		assertNil(t, err)
		assertStrings(t, got, definition)
	})

	t.Run("existing word", func(t *testing.T) {
		dict := Dictionary{}
		word := "carrot"
		definition := "The orangest vegetable, pumpkins be damned"

		// Add once is fine
		err := dict.Add(word, definition)
		assertNil(t, err)

		// Add twice is an error
		err = dict.Add(word, definition)
		assertError(t, err, ErrExisting)
	})
}

func TestSearch(t *testing.T) {
	dict := Dictionary{"test": "This is just a test"}

	t.Run("known word", func(t *testing.T) {
		got, err := dict.Search("test")
		want := "This is just a test"
		assertNil(t, err)
		assertStrings(t, got, want)
	})

	t.Run("word not found", func(t *testing.T) {
		_, err := dict.Search("unknown")
		if err == nil {
			t.Fatal("expected an error")
		}
		assertError(t, err, ErrNotFound)
	})
}

func TestUpdate(t *testing.T) {
	t.Run("existing word", func(t *testing.T) {
		word := "potato"
		definition := "A starchy tuberous vegetable native to the Americas"
		updated := "What you use to make chips!"
		dict := Dictionary{word: definition}

		err := dict.Update(word, updated)
		assertNil(t, err)

		got, err := dict.Search(word)
		assertNil(t, err)
		assertStrings(t, got, updated)
	})

	t.Run("new word", func(t *testing.T) {
		word := "potato"
		definition := "A starchy tuberous vegetable native to the Americas"
		dict := Dictionary{}

		err := dict.Update(word, definition)
		assertError(t, err, ErrNotFound)
	})
}

func assertError(t testing.TB, got, want error) {
	t.Helper()
	if !errors.Is(got, want) {
		t.Errorf("got error %q, want %q", got, want)
	}
}

func assertNil(t testing.TB, got any) {
	t.Helper()
	if got != nil {
		t.Errorf("expected nil, got %v", got)
	}
}

func assertStrings(t testing.TB, got, want string) {
	t.Helper()
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
