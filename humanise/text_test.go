package humanise_test

import (
	"testing"

	"local.dev/humanise"
)

// TestAnd covers the count boundaries and the verbatim, no-filter behaviour.
func TestAnd(t *testing.T) {
	tests := []struct {
		name  string
		items []string
		want  string
	}{
		{"nil", nil, ""},
		{"empty", []string{}, ""},
		{"one", []string{"apples"}, "apples"},
		{"two takes no comma", []string{"apples", "oranges"}, "apples and oranges"},
		{"three take the serial comma", []string{"apples", "oranges", "bananas"}, "apples, oranges, and bananas"},
		{"four", []string{"a", "b", "c", "d"}, "a, b, c, and d"},
		{"blank items are kept", []string{"a", "", "b"}, "a, , and b"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := humanise.And(tt.items); got != tt.want {
				t.Errorf("And(%q) = %q, want %q", tt.items, got, tt.want)
			}
		})
	}
}

// TestOr mirrors TestAnd for the "or" conjunction.
func TestOr(t *testing.T) {
	tests := []struct {
		name  string
		items []string
		want  string
	}{
		{"nil", nil, ""},
		{"one", []string{"apples"}, "apples"},
		{"two takes no comma", []string{"apples", "oranges"}, "apples or oranges"},
		{"three take the serial comma", []string{"apples", "oranges", "bananas"}, "apples, oranges, or bananas"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := humanise.Or(tt.items); got != tt.want {
				t.Errorf("Or(%q) = %q, want %q", tt.items, got, tt.want)
			}
		})
	}
}

// BenchmarkAnd measures joining a three-item list, exercising the serial-comma path.
func BenchmarkAnd(b *testing.B) {
	items := []string{"apples", "oranges", "bananas"}
	for b.Loop() {
		humanise.And(items)
	}
}

// TestTitle covers minor-word handling, preserved caps and the all-caps line we drew.
func TestTitle(t *testing.T) {
	tests := []struct {
		name  string
		title string
		want  string
	}{
		{"empty", "", ""},
		{"whitespace only", "   \t ", ""},
		{"docstring example", "taming of the shrew", "Taming of the Shrew"},
		{"tidies capitalised minor words", "Taming Of The Shrew", "Taming of the Shrew"},
		{"all caps is left shouting", "TAMING OF THE SHREW", "TAMING of the SHREW"},
		{"first and middle minor words", "the lord of the rings", "The Lord of the Rings"},
		{"last minor word is capitalised", "what are you waiting for", "What Are You Waiting For"},
		{"internal caps preserved", "the McDonald story", "The McDonald Story"},
		{"acronym preserved", "a NASA mission", "A NASA Mission"},
		{"collapses whitespace", "  the   lord\tof  rings  ", "The Lord of Rings"},
		{"single word", "hello", "Hello"},
		{"single minor word", "the", "The"},
		{"hyphen not internally capitalised", "mother-in-law", "Mother-in-law"},
		{"apostrophe not internally capitalised", "james's diary", "James's Diary"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := humanise.Title(tt.title); got != tt.want {
				t.Errorf("Title(%q) = %q, want %q", tt.title, got, tt.want)
			}
		})
	}
}

// BenchmarkTitle measures capitalising a short title with a mix of minor words.
func BenchmarkTitle(b *testing.B) {
	for b.Loop() {
		humanise.Title("the lord of the rings")
	}
}
