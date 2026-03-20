package categorise

import (
	"testing"
)

func TestLoadPrefixes(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		got, err := LoadPrefixes("testdata/valid.csv")
		if err != nil {
			t.Fatal(err)
		}

		want := []Prefix{
			{Text: "walmart", Category: "Shopping"},
			{Text: "amazon", Category: "Online"},
		}
		if len(got) != len(want) {
			t.Fatalf("got %d prefixes, want %d", len(got), len(want))
		}
		for i := range want {
			if got[i] != want[i] {
				t.Errorf("prefix[%d] = %+v, want %+v", i, got[i], want[i])
			}
		}
	})

	t.Run("lowercases and trims", func(t *testing.T) {
		got, err := LoadPrefixes("testdata/trim.csv")
		if err != nil {
			t.Fatal(err)
		}

		if len(got) != 1 {
			t.Fatalf("got %d prefixes, want 1", len(got))
		}
		if got[0].Text != "walmart" {
			t.Errorf("Text = %q, want %q", got[0].Text, "walmart")
		}
		if got[0].Category != "Shopping" {
			t.Errorf("Category = %q, want %q", got[0].Category, "Shopping")
		}
	})

	t.Run("empty file", func(t *testing.T) {
		got, err := LoadPrefixes("testdata/empty.csv")
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != 0 {
			t.Fatalf("got %d prefixes, want 0", len(got))
		}
	})

	t.Run("file not found", func(t *testing.T) {
		_, err := LoadPrefixes("testdata/nonexistent.csv")
		if err == nil {
			t.Fatal("expected error for missing file")
		}
	})

	t.Run("wrong field count", func(t *testing.T) {
		_, err := LoadPrefixes("testdata/wrong_fields.csv")
		if err == nil {
			t.Fatal("expected error for wrong field count")
		}
	})
}

func TestMatch(t *testing.T) {
	matcher := NewMatcher([]Prefix{
		{Text: "cafe", Category: "Food/Cafe"},
		{Text: "woolworths", Category: "Food/Groceries"},
		{Text: "woolworths nz/lynnmall", Category: "Food/Groceries/Local"},
		{Text: "wool", Category: "Clothing"},
	})

	t.Run("exact match", func(t *testing.T) {
		got := matcher.Match("cafe")
		if got != "Food/Cafe" {
			t.Errorf("got %q, want %q", got, "Food/Cafe")
		}
	})

	t.Run("prefix match", func(t *testing.T) {
		got := matcher.Match("woolworths nz/26 custo")
		if got != "Food/Groceries" {
			t.Errorf("got %q, want %q", got, "Food/Groceries")
		}
	})

	t.Run("longest match wins", func(t *testing.T) {
		got := matcher.Match("woolworths nz/lynnmall new lynn")
		if got != "Food/Groceries/Local" {
			t.Errorf("got %q, want %q", got, "Food/Groceries/Local")
		}
	})

	t.Run("falls back to shorter prefix", func(t *testing.T) {
		got := matcher.Match("woolly jumper shop")
		if got != "Clothing" {
			t.Errorf("got %q, want %q", got, "Clothing")
		}
	})

	t.Run("case insensitive", func(t *testing.T) {
		got := matcher.Match("WOOLWORTHS NZ/Lynnmall New Lynn")
		if got != "Food/Groceries/Local" {
			t.Errorf("got %q, want %q", got, "Food/Groceries/Local")
		}
	})

	t.Run("no match", func(t *testing.T) {
		got := matcher.Match("burger king new lynn")
		if got != Unknown {
			t.Errorf("got %q, want %q", got, Unknown)
		}
	})

	t.Run("empty prefixes", func(t *testing.T) {
		empty := NewMatcher(nil)
		got := empty.Match("anything")
		if got != Unknown {
			t.Errorf("got %q, want %q", got, Unknown)
		}
	})
}

func TestSummary(t *testing.T) {
	t.Run("single category", func(t *testing.T) {
		var s Summary
		s.Add("Food/Groceries", "Woolworths NZ")

		if len(s.Groups) != 1 {
			t.Fatalf("got %d root groups, want 1", len(s.Groups))
		}
		food := s.Groups[0]
		if food.Name != "Food" {
			t.Errorf("root name = %q, want %q", food.Name, "Food")
		}
		if food.Count() != 1 {
			t.Errorf("Food count = %d, want 1", food.Count())
		}
		if len(food.Children) != 1 {
			t.Fatalf("Food children = %d, want 1", len(food.Children))
		}
		groceries := food.Children[0]
		if groceries.Name != "Groceries" {
			t.Errorf("child name = %q, want %q", groceries.Name, "Groceries")
		}
		if groceries.Count() != 1 {
			t.Errorf("Groceries count = %d, want 1", groceries.Count())
		}
	})

	t.Run("multiple details same category", func(t *testing.T) {
		var s Summary
		s.Add("Food/Groceries", "Woolworths NZ")
		s.Add("Food/Groceries", "Countdown Auckland")

		food := s.Groups[0]
		if food.Count() != 2 {
			t.Errorf("Food count = %d, want 2", food.Count())
		}
		groceries := food.Children[0]
		if groceries.Count() != 2 {
			t.Errorf("Groceries count = %d, want 2", groceries.Count())
		}
	})

	t.Run("sibling categories", func(t *testing.T) {
		var s Summary
		s.Add("Food/Groceries", "Woolworths NZ")
		s.Add("Food/Cafe", "Cafe Mocha")

		food := s.Groups[0]
		if food.Count() != 2 {
			t.Errorf("Food count = %d, want 2", food.Count())
		}
		if len(food.Children) != 2 {
			t.Fatalf("Food children = %d, want 2", len(food.Children))
		}
		if food.Children[0].Name != "Groceries" || food.Children[1].Name != "Cafe" {
			t.Errorf("children = [%q, %q], want [Groceries, Cafe]",
				food.Children[0].Name, food.Children[1].Name)
		}
	})

	t.Run("distinct root categories", func(t *testing.T) {
		var s Summary
		s.Add("Food/Groceries", "Woolworths NZ")
		s.Add("Transport/Public", "Auckland Transport")

		if len(s.Groups) != 2 {
			t.Fatalf("got %d root groups, want 2", len(s.Groups))
		}
		if s.Groups[0].Name != "Food" || s.Groups[1].Name != "Transport" {
			t.Errorf("roots = [%q, %q], want [Food, Transport]",
				s.Groups[0].Name, s.Groups[1].Name)
		}
	})

	t.Run("single segment category", func(t *testing.T) {
		var s Summary
		s.Add("Misc", "Random Purchase")

		if len(s.Groups) != 1 {
			t.Fatalf("got %d root groups, want 1", len(s.Groups))
		}
		if s.Groups[0].Name != "Misc" {
			t.Errorf("root name = %q, want %q", s.Groups[0].Name, "Misc")
		}
		if s.Groups[0].Count() != 1 {
			t.Errorf("count = %d, want 1", s.Groups[0].Count())
		}
		if len(s.Groups[0].Children) != 0 {
			t.Errorf("children = %d, want 0", len(s.Groups[0].Children))
		}
	})
}

func BenchmarkMatch(b *testing.B) {
	prefixes, err := LoadPrefixes("testdata/prefixes.csv")
	if err != nil {
		b.Fatal(err)
	}
	matcher := NewMatcher(prefixes)

	details := []string{
		"Woolworths Nz/Lynnmall New Lynn Nz",
		"Auckland Transport Auckland Nz",
		"Burger King New Lynn Nz",
		"Online Payment - Thank You",
	}

	for b.Loop() {
		for _, d := range details {
			matcher.Match(d)
		}
	}
}
