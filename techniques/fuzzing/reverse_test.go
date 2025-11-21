package main

import (
	"testing"
	"unicode/utf8"
)

func TestReverse(t *testing.T) {
	testcases := []struct {
		in   string
		want string
	}{
		{"Hello, world", "dlrow ,olleH"},
		{" ", " "},
		{"!123456", "654321!"},
	}

	for _, tt := range testcases {
		rev, err := Reverse(tt.in)
		if err != nil {
			t.Error(err)
		}
		if rev != tt.want {
			t.Errorf("Got %q, want %q", rev, tt.want)
		}
	}
}

func FuzzReverse(f *testing.F) {
	// Populate test's 'seed corpus'
	testcases := []string{"Hello, world", " ", "!123456"}
	for _, str := range testcases {
		f.Add(str)
	}

	// A fuzz target must accept a *T parameter, followed by one or more
	// parameters for random inputs. The types of arguments passed to F.Add
	// must be identical to the types of these parameters.
	target := func(t *testing.T, original string) {
		// Round-trip
		reversed, err := Reverse(original)
		if err != nil {
			return
		}

		doubled, err := Reverse(reversed)
		if err != nil {
			t.Error("Output error", err)
		}
		if original != doubled {
			t.Errorf("Before: %q, after: %q", original, doubled)
		}

		// Valid UTF-8 input should result in valid output
		if !utf8.ValidString(reversed) {
			t.Errorf("Reverse produced invalid UTF-8 string %q", reversed)
		}
	}

	f.Fuzz(target)
}
