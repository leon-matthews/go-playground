package tui

import (
	"testing"

	"banking/common"
)

func TestRewritePrefixes(t *testing.T) {
	t.Run("rename leaf", func(t *testing.T) {
		m := NewCategoriesModel([]common.Prefix{
			{Text: "woolworths", Category: "Food/Groceries"},
			{Text: "cafe", Category: "Food/Cafe"},
		})
		m.rewritePrefixes("Food/Cafe", "Food/Coffee")

		assertCategory(t, m.Prefixes, "woolworths", "Food/Groceries")
		assertCategory(t, m.Prefixes, "cafe", "Food/Coffee")
		if !m.Changed {
			t.Error("expected Changed to be true")
		}
	})

	t.Run("rename parent rewrites children", func(t *testing.T) {
		m := NewCategoriesModel([]common.Prefix{
			{Text: "woolworths", Category: "Food/Groceries"},
			{Text: "cafe", Category: "Food/Cafe"},
			{Text: "woolworths lynnmall", Category: "Food/Groceries/Local"},
		})
		m.rewritePrefixes("Food", "Dining")

		assertCategory(t, m.Prefixes, "woolworths", "Dining/Groceries")
		assertCategory(t, m.Prefixes, "cafe", "Dining/Cafe")
		assertCategory(t, m.Prefixes, "woolworths lynnmall", "Dining/Groceries/Local")
	})

	t.Run("no false matches", func(t *testing.T) {
		m := NewCategoriesModel([]common.Prefix{
			{Text: "woolworths", Category: "Food/Groceries"},
			{Text: "bus", Category: "Transport/Public"},
		})
		m.rewritePrefixes("Food/Groceries", "Food/Supermarket")

		assertCategory(t, m.Prefixes, "woolworths", "Food/Supermarket")
		assertCategory(t, m.Prefixes, "bus", "Transport/Public")
	})

	t.Run("move subtree", func(t *testing.T) {
		m := NewCategoriesModel([]common.Prefix{
			{Text: "cafe", Category: "Food/Cafe"},
			{Text: "fancy cafe", Category: "Food/Cafe/Fancy"},
		})
		m.rewritePrefixes("Food/Cafe", "Dining/Cafe")

		assertCategory(t, m.Prefixes, "cafe", "Dining/Cafe")
		assertCategory(t, m.Prefixes, "fancy cafe", "Dining/Cafe/Fancy")
	})

	t.Run("rebuilds tree", func(t *testing.T) {
		m := NewCategoriesModel([]common.Prefix{
			{Text: "cafe", Category: "Food/Cafe"},
		})
		m.rewritePrefixes("Food/Cafe", "Dining/Cafe")

		if len(m.tree) == 0 {
			t.Fatal("tree was not rebuilt")
		}
		if m.tree[0].Name != "Dining" {
			t.Errorf("first tree node = %q, want Dining", m.tree[0].Name)
		}
	})
}

func TestDeletePrefixes(t *testing.T) {
	t.Run("delete leaf", func(t *testing.T) {
		m := NewCategoriesModel([]common.Prefix{
			{Text: "woolworths", Category: "Food/Groceries"},
			{Text: "cafe", Category: "Food/Cafe"},
		})
		got := m.deletePrefixes("Food/Cafe")

		if len(got) != 1 {
			t.Fatalf("got %d prefixes, want 1", len(got))
		}
		if got[0].Text != "woolworths" {
			t.Errorf("remaining prefix = %q, want woolworths", got[0].Text)
		}
	})

	t.Run("delete parent removes children", func(t *testing.T) {
		m := NewCategoriesModel([]common.Prefix{
			{Text: "woolworths", Category: "Food/Groceries"},
			{Text: "cafe", Category: "Food/Cafe"},
			{Text: "bus", Category: "Transport/Public"},
		})
		got := m.deletePrefixes("Food")

		if len(got) != 1 {
			t.Fatalf("got %d prefixes, want 1", len(got))
		}
		if got[0].Category != "Transport/Public" {
			t.Errorf("remaining category = %q, want Transport/Public", got[0].Category)
		}
	})

	t.Run("no false matches", func(t *testing.T) {
		m := NewCategoriesModel([]common.Prefix{
			{Text: "woolworths", Category: "Food/Groceries"},
			{Text: "bus", Category: "Transport/Public"},
		})
		got := m.deletePrefixes("Food/Groceries")

		if len(got) != 1 {
			t.Fatalf("got %d prefixes, want 1", len(got))
		}
		if got[0].Category != "Transport/Public" {
			t.Errorf("remaining = %q, want Transport/Public", got[0].Category)
		}
	})

	t.Run("delete all", func(t *testing.T) {
		m := NewCategoriesModel([]common.Prefix{
			{Text: "cafe", Category: "Food/Cafe"},
		})
		got := m.deletePrefixes("Food")

		if len(got) != 0 {
			t.Fatalf("got %d prefixes, want 0", len(got))
		}
	})

	t.Run("delete nonexistent path", func(t *testing.T) {
		m := NewCategoriesModel([]common.Prefix{
			{Text: "cafe", Category: "Food/Cafe"},
		})
		got := m.deletePrefixes("Transport")

		if len(got) != 1 {
			t.Fatalf("got %d prefixes, want 1", len(got))
		}
	})
}

func assertCategory(t *testing.T, prefixes []common.Prefix, text, wantCategory string) {
	t.Helper()
	for _, p := range prefixes {
		if p.Text == text {
			if p.Category != wantCategory {
				t.Errorf("prefix %q: category = %q, want %q", text, p.Category, wantCategory)
			}
			return
		}
	}
	t.Errorf("prefix %q not found", text)
}
