package main

import "testing"

// TestMultisetMedianEmpty checks that empty and all-zero multisets return an error.
func TestMultisetMedianEmpty(t *testing.T) {
	for _, counts := range [][]int64{{}, {0, 0, 0}} {
		if _, err := multisetMedian(counts, medianMean); err == nil {
			t.Errorf("expected an error for counts %v", counts)
		}
	}
}

// TestMultisetMedian checks every mode, including totals beyond 32 bits.
func TestMultisetMedian(t *testing.T) {
	tests := []struct {
		name   string
		counts []int64
		mode   medianMode
		want   float64
	}{
		{"single", []int64{5: 1}, medianMean, 5},
		{"odd total", []int64{1: 1, 2: 1, 3: 1}, medianMean, 2},
		{"even mean", []int64{1: 2, 9: 2}, medianMean, 5},
		{"even low", []int64{1: 2, 9: 2}, medianLow, 1},
		{"even high", []int64{1: 2, 9: 2}, medianHigh, 9},
		{"weighted", []int64{1: 9, 100: 1}, medianMean, 1},
		{"billions mean", []int64{10: 3_000_000_000, 20: 3_000_000_000}, medianMean, 15},
		{"billions high", []int64{10: 3_000_000_000, 20: 3_000_000_000}, medianHigh, 20},
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
