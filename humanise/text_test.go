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
