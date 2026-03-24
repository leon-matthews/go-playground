package categorise

import (
	"path/filepath"
	"testing"

	"banking/common"
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

func TestSavePrefixes(t *testing.T) {
	t.Run("round-trip", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "prefixes.csv")
		input := []Prefix{
			{Text: "walmart", Category: "Shopping"},
			{Text: "amazon", Category: "Online"},
		}

		if err := SavePrefixes(path, input); err != nil {
			t.Fatal(err)
		}

		got, err := LoadPrefixes(path)
		if err != nil {
			t.Fatal(err)
		}

		// SavePrefixes sorts, so expect alphabetical order
		want := []Prefix{
			{Text: "amazon", Category: "Online"},
			{Text: "walmart", Category: "Shopping"},
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

	t.Run("empty", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "prefixes.csv")

		if err := SavePrefixes(path, nil); err != nil {
			t.Fatal(err)
		}

		got, err := LoadPrefixes(path)
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != 0 {
			t.Fatalf("got %d prefixes, want 0", len(got))
		}
	})

	t.Run("does not mutate input", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "prefixes.csv")
		input := []Prefix{
			{Text: "walmart", Category: "Shopping"},
			{Text: "amazon", Category: "Online"},
		}

		if err := SavePrefixes(path, input); err != nil {
			t.Fatal(err)
		}

		// Original slice should still be in its original order
		if input[0].Text != "walmart" || input[1].Text != "amazon" {
			t.Errorf("input was mutated: %+v", input)
		}
	})

	t.Run("unwritable path", func(t *testing.T) {
		err := SavePrefixes("/no/such/directory/prefixes.csv", []Prefix{
			{Text: "test", Category: "Test"},
		})
		if err == nil {
			t.Fatal("expected error for unwritable path")
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
		s.Add("Food/Groceries", new(common.Transaction{Details: "Woolworths NZ", Amount: -55.00}))

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
		if food.Total != -55.00 {
			t.Errorf("Food total = %.2f, want -55.00", food.Total)
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
		s.Add("Food/Groceries", new(common.Transaction{Details: "Woolworths NZ", Amount: -55.00}))
		s.Add("Food/Groceries", new(common.Transaction{Details: "Countdown Auckland", Amount: -30.00}))

		food := s.Groups[0]
		if food.Count() != 2 {
			t.Errorf("Food count = %d, want 2", food.Count())
		}
		if food.Total != -85.00 {
			t.Errorf("Food total = %.2f, want -85.00", food.Total)
		}
		groceries := food.Children[0]
		if groceries.Count() != 2 {
			t.Errorf("Groceries count = %d, want 2", groceries.Count())
		}
	})

	t.Run("sibling categories", func(t *testing.T) {
		var s Summary
		s.Add("Food/Groceries", new(common.Transaction{Details: "Woolworths NZ", Amount: -55.00}))
		s.Add("Food/Cafe", new(common.Transaction{Details: "Cafe Mocha", Amount: -8.50}))

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
		s.Add("Food/Groceries", new(common.Transaction{Details: "Woolworths NZ", Amount: -55.00}))
		s.Add("Transport/Public", new(common.Transaction{Details: "Auckland Transport", Amount: -3.50}))

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
		s.Add("Misc", new(common.Transaction{Details: "Random Purchase", Amount: -12.99}))

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

func TestSort(t *testing.T) {
	var s Summary
	s.Add("Transport/Public", new(common.Transaction{Details: "bus", Amount: -3.50}))
	s.Add("Food/Groceries", new(common.Transaction{Details: "woolworths", Amount: -55.00}))
	s.Add("Food/Cafe", new(common.Transaction{Details: "cafe", Amount: -8.50}))
	s.Add("Entertainment", new(common.Transaction{Details: "netflix", Amount: -15.00}))

	s.Sort()

	// Roots should be alphabetical
	if len(s.Groups) != 3 {
		t.Fatalf("got %d root groups, want 3", len(s.Groups))
	}
	if s.Groups[0].Name != "Entertainment" {
		t.Errorf("root[0] = %q, want Entertainment", s.Groups[0].Name)
	}
	if s.Groups[1].Name != "Food" {
		t.Errorf("root[1] = %q, want Food", s.Groups[1].Name)
	}
	if s.Groups[2].Name != "Transport" {
		t.Errorf("root[2] = %q, want Transport", s.Groups[2].Name)
	}

	// Children under Food should be alphabetical
	food := s.Groups[1]
	if len(food.Children) != 2 {
		t.Fatalf("Food children = %d, want 2", len(food.Children))
	}
	if food.Children[0].Name != "Cafe" {
		t.Errorf("Food child[0] = %q, want Cafe", food.Children[0].Name)
	}
	if food.Children[1].Name != "Groceries" {
		t.Errorf("Food child[1] = %q, want Groceries", food.Children[1].Name)
	}
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
