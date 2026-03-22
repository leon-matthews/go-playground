package tui

import (
	"testing"
	"time"

	"banking/categorise"
	"banking/common"
)

func TestMeasureTree(t *testing.T) {
	t.Run("group names only", func(t *testing.T) {
		var s categorise.Summary
		s.Add("Food", "test", -10)
		s.Add("Transport", "test", -5)

		got := MeasureTree(s.Groups, 0, 1, false, nil, "")

		// "Transport" is 9 chars, the longest at depth 0
		if got != 9 {
			t.Errorf("got %d, want 9", got)
		}
	})

	t.Run("includes indent for children", func(t *testing.T) {
		var s categorise.Summary
		s.Add("Food/Groceries", "test", -10)

		got := MeasureTree(s.Groups, 0, 2, false, nil, "")

		// depth 0: "Food" = 4
		// depth 1: 2 (indent) + "Groceries" (9) = 11
		if got != 11 {
			t.Errorf("got %d, want 11", got)
		}
	})

	t.Run("respects maxDepth", func(t *testing.T) {
		var s categorise.Summary
		s.Add("Food/Groceries", "test", -10)

		got := MeasureTree(s.Groups, 0, 1, false, nil, "")

		// Only depth 0: "Food" = 4
		if got != 4 {
			t.Errorf("got %d, want 4", got)
		}
	})

	t.Run("includes transactions", func(t *testing.T) {
		var s categorise.Summary
		s.Add("Food", "test", -10)

		tx := &common.Transaction{
			Date:    time.Date(2026, 3, 5, 0, 0, 0, 0, time.UTC),
			Account: "Visa",
			Details: "Woolworths",
			Amount:  -55,
		}
		byCategory := map[string][]*common.Transaction{
			"Food": {tx},
		}

		got := MeasureTree(s.Groups, 0, 1<<31-1, true, byCategory, "")

		// depth 1 indent (2) + "5 Mar 2026" (10) + 2 + "Visa" (4) + 2 + "Woolworths" (10) = 30
		want := 30
		if got != want {
			t.Errorf("got %d, want %d", got, want)
		}
	})

	t.Run("transactions not measured when showTx false", func(t *testing.T) {
		var s categorise.Summary
		s.Add("Food", "test", -10)

		tx := &common.Transaction{
			Date:    time.Date(2026, 3, 5, 0, 0, 0, 0, time.UTC),
			Account: "Visa",
			Details: "Woolworths Nz/Lynnmall New Lynn",
			Amount:  -55,
		}
		byCategory := map[string][]*common.Transaction{
			"Food": {tx},
		}

		got := MeasureTree(s.Groups, 0, 1, false, byCategory, "")

		// Only the group name: "Food" = 4
		if got != 4 {
			t.Errorf("got %d, want 4", got)
		}
	})

	t.Run("empty groups", func(t *testing.T) {
		got := MeasureTree(nil, 0, 10, false, nil, "")
		if got != 0 {
			t.Errorf("got %d, want 0", got)
		}
	})
}
