package tui

import (
	"bytes"
	"flag"
	"os"
	"path/filepath"
	"testing"
	"time"

	"banking/categorise"
	"banking/common"
)

var update = flag.Bool("update", false, "update golden files")

func TestMeasure(t *testing.T) {
	t.Run("group names only", func(t *testing.T) {
		var s categorise.Summary
		s.Add("Food", "test", -10)
		s.Add("Transport", "test", -5)

		p := TreePrinter{MaxDepth: 1}
		got := p.Measure(s.Groups)

		// "Transport" is 9 chars, the longest at depth 0
		if got != 9 {
			t.Errorf("got %d, want 9", got)
		}
	})

	t.Run("includes indent for children", func(t *testing.T) {
		var s categorise.Summary
		s.Add("Food/Groceries", "test", -10)

		p := TreePrinter{MaxDepth: 2}
		got := p.Measure(s.Groups)

		// depth 0: "Food" = 4
		// depth 1: 2 (indent) + "Groceries" (9) = 11
		if got != 11 {
			t.Errorf("got %d, want 11", got)
		}
	})

	t.Run("respects maxDepth", func(t *testing.T) {
		var s categorise.Summary
		s.Add("Food/Groceries", "test", -10)

		p := TreePrinter{MaxDepth: 1}
		got := p.Measure(s.Groups)

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

		p := TreePrinter{MaxDepth: 1<<31 - 1, ShowTx: true, ByCategory: byCategory}
		got := p.Measure(s.Groups)

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

		p := TreePrinter{MaxDepth: 1, ByCategory: byCategory}
		got := p.Measure(s.Groups)

		// Only the group name: "Food" = 4
		if got != 4 {
			t.Errorf("got %d, want 4", got)
		}
	})

	t.Run("empty groups", func(t *testing.T) {
		p := TreePrinter{MaxDepth: 10}
		got := p.Measure(nil)
		if got != 0 {
			t.Errorf("got %d, want 0", got)
		}
	})
}

func TestPrint(t *testing.T) {
	t.Run("group_names_only", func(t *testing.T) {
		var s categorise.Summary
		s.Add("Food", "test", -10)
		s.Add("Transport", "test", -5)
		s.Sort()

		var buf bytes.Buffer
		p := TreePrinter{W: &buf, MaxDepth: 1, LeftWidth: 20}
		p.Print(s.Groups)
		compareGolden(t, "PrintTree_group_names_only", buf.Bytes())
	})

	t.Run("nested_groups", func(t *testing.T) {
		var s categorise.Summary
		s.Add("Food/Groceries", "woolworths", -55)
		s.Add("Food/Cafe", "cafe mocha", -8.50)
		s.Add("Transport/Public", "bus", -3.50)
		s.Sort()

		var buf bytes.Buffer
		p := TreePrinter{W: &buf, MaxDepth: 3, LeftWidth: 20}
		p.Print(s.Groups)
		compareGolden(t, "PrintTree_nested_groups", buf.Bytes())
	})

	t.Run("with_transactions", func(t *testing.T) {
		var s categorise.Summary
		s.Add("Food", "Woolworths Auckland", -55)
		s.Add("Food", "Countdown Mt Eden", -30)

		tx1 := &common.Transaction{
			Date:    time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC),
			Account: "Visa",
			Details: "Woolworths Auckland",
			Amount:  -55,
		}
		tx2 := &common.Transaction{
			Date:    time.Date(2026, 1, 16, 0, 0, 0, 0, time.UTC),
			Account: "Visa",
			Details: "Countdown Mt Eden",
			Amount:  -30,
		}
		byCategory := map[string][]*common.Transaction{
			"Food": {tx1, tx2},
		}

		var buf bytes.Buffer
		p := TreePrinter{W: &buf, MaxDepth: 1<<31 - 1, ShowTx: true, ByCategory: byCategory, LeftWidth: 50}
		p.Print(s.Groups)
		compareGolden(t, "PrintTree_with_transactions", buf.Bytes())
	})

	t.Run("truncated_details", func(t *testing.T) {
		var s categorise.Summary
		s.Add("Food", "Woolworths Nz/Lynnmall New Lynn Auckland", -55)

		tx := &common.Transaction{
			Date:    time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC),
			Account: "Visa",
			Details: "Woolworths Nz/Lynnmall New Lynn Auckland",
			Amount:  -55,
		}
		byCategory := map[string][]*common.Transaction{
			"Food": {tx},
		}

		var buf bytes.Buffer
		p := TreePrinter{W: &buf, MaxDepth: 1<<31 - 1, ShowTx: true, ByCategory: byCategory, LeftWidth: 30}
		p.Print(s.Groups)
		compareGolden(t, "PrintTree_truncated_details", buf.Bytes())
	})

	t.Run("empty", func(t *testing.T) {
		var buf bytes.Buffer
		p := TreePrinter{W: &buf, MaxDepth: 10, LeftWidth: 20}
		p.Print(nil)
		compareGolden(t, "PrintTree_empty", buf.Bytes())
	})
}

func compareGolden(t *testing.T, name string, got []byte) {
	t.Helper()
	golden := filepath.Join("testdata", name+".txt")
	if *update {
		if err := os.WriteFile(golden, got, 0o644); err != nil {
			t.Fatalf("write golden file: %v", err)
		}
	}
	want, err := os.ReadFile(golden)
	if err != nil {
		t.Fatalf("read golden file: %v (run with -update to create)", err)
	}
	if !bytes.Equal(got, want) {
		t.Errorf("output mismatch\nwant:\n%s\ngot:\n%s", want, got)
	}
}
