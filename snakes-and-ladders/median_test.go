package main

import "testing"

// TestMultisetMedianEmpty checks that an empty multiset returns an error.
func TestMultisetMedianEmpty(t *testing.T) {
	if _, err := multisetMedian(map[int]int64{}, medianMean); err == nil {
		t.Error("expected an error for empty counts")
	}
}

// TestMultisetMedian checks every mode, including totals beyond 32 bits.
func TestMultisetMedian(t *testing.T) {
	tests := []struct {
		name   string
		counts map[int]int64
		mode   medianMode
		want   float64
	}{
		{"single", map[int]int64{5: 1}, medianMean, 5},
		{"odd total", map[int]int64{1: 1, 2: 1, 3: 1}, medianMean, 2},
		{"even mean", map[int]int64{1: 2, 9: 2}, medianMean, 5},
		{"even low", map[int]int64{1: 2, 9: 2}, medianLow, 1},
		{"even high", map[int]int64{1: 2, 9: 2}, medianHigh, 9},
		{"weighted", map[int]int64{1: 9, 100: 1}, medianMean, 1},
		{"billions mean", map[int]int64{10: 3_000_000_000, 20: 3_000_000_000}, medianMean, 15},
		{"billions high", map[int]int64{10: 3_000_000_000, 20: 3_000_000_000}, medianHigh, 20},
	}
	for _, test := range tests {
		got, err := multisetMedian(test.counts, test.mode)
		if err != nil {
			t.Errorf("%s: unexpected error: %v", test.name, err)
			continue
		}
		if got != test.want {
			t.Errorf("%s: median = %v, want %v", test.name, got, test.want)
		}
	}
}
