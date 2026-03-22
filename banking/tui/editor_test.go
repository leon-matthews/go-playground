package tui

import (
	"testing"

	"banking/categorise"
)

func TestBuildCategoryTree(t *testing.T) {
	t.Run("single category", func(t *testing.T) {
		prefixes := []categorise.Prefix{
			{Text: "walmart", Category: "Shopping"},
		}
		got := BuildCategoryTree(prefixes)

		want := []TreeRow{
			{Name: "Shopping", Depth: 0, Path: "Shopping"},
		}
		assertTreeRows(t, got, want)
	})

	t.Run("hierarchical categories", func(t *testing.T) {
		prefixes := []categorise.Prefix{
			{Text: "woolworths", Category: "Food/Groceries"},
			{Text: "cafe", Category: "Food/Cafe"},
		}
		got := BuildCategoryTree(prefixes)

		want := []TreeRow{
			{Name: "Food", Depth: 0, Path: "Food"},
			{Name: "Cafe", Depth: 1, Path: "Food/Cafe"},
			{Name: "Groceries", Depth: 1, Path: "Food/Groceries"},
		}
		assertTreeRows(t, got, want)
	})

	t.Run("multiple roots", func(t *testing.T) {
		prefixes := []categorise.Prefix{
			{Text: "woolworths", Category: "Food/Groceries"},
			{Text: "bus", Category: "Transport/Public"},
		}
		got := BuildCategoryTree(prefixes)

		want := []TreeRow{
			{Name: "Food", Depth: 0, Path: "Food"},
			{Name: "Groceries", Depth: 1, Path: "Food/Groceries"},
			{Name: "Transport", Depth: 0, Path: "Transport"},
			{Name: "Public", Depth: 1, Path: "Transport/Public"},
		}
		assertTreeRows(t, got, want)
	})

	t.Run("three levels deep", func(t *testing.T) {
		prefixes := []categorise.Prefix{
			{Text: "woolworths lynnmall", Category: "Food/Groceries/Local"},
		}
		got := BuildCategoryTree(prefixes)

		want := []TreeRow{
			{Name: "Food", Depth: 0, Path: "Food"},
			{Name: "Groceries", Depth: 1, Path: "Food/Groceries"},
			{Name: "Local", Depth: 2, Path: "Food/Groceries/Local"},
		}
		assertTreeRows(t, got, want)
	})

	t.Run("deduplicates categories", func(t *testing.T) {
		prefixes := []categorise.Prefix{
			{Text: "woolworths", Category: "Food/Groceries"},
			{Text: "new world", Category: "Food/Groceries"},
			{Text: "costco", Category: "Food/Groceries"},
		}
		got := BuildCategoryTree(prefixes)

		want := []TreeRow{
			{Name: "Food", Depth: 0, Path: "Food"},
			{Name: "Groceries", Depth: 1, Path: "Food/Groceries"},
		}
		assertTreeRows(t, got, want)
	})

	t.Run("sorted output", func(t *testing.T) {
		prefixes := []categorise.Prefix{
			{Text: "bus", Category: "Transport/Public"},
			{Text: "cafe", Category: "Food/Cafe"},
			{Text: "woolworths", Category: "Food/Groceries"},
			{Text: "netflix", Category: "Entertainment"},
		}
		got := BuildCategoryTree(prefixes)

		if got[0].Name != "Entertainment" {
			t.Errorf("first root = %q, want Entertainment", got[0].Name)
		}
		if got[1].Name != "Food" {
			t.Errorf("second root = %q, want Food", got[1].Name)
		}
		if got[2].Name != "Cafe" {
			t.Errorf("first Food child = %q, want Cafe", got[2].Name)
		}
		if got[3].Name != "Groceries" {
			t.Errorf("second Food child = %q, want Groceries", got[3].Name)
		}
	})

	t.Run("empty input", func(t *testing.T) {
		got := BuildCategoryTree(nil)
		if len(got) != 0 {
			t.Fatalf("got %d rows, want 0", len(got))
		}
	})
}

func assertTreeRows(t *testing.T, got, want []TreeRow) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("got %d rows, want %d\n  got:  %+v\n  want: %+v", len(got), len(want), got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("row[%d] = %+v, want %+v", i, got[i], want[i])
		}
	}
}
