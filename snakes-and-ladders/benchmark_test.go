package main

import (
	"maps"
	"math"
	"math/rand/v2"
	"slices"
	"testing"
)

// BenchmarkPlayCount measures playCount throughput in games per second.
func BenchmarkPlayCount(b *testing.B) {
	// Fixed seed keeps runs comparable between benchmark sessions
	rng := rand.NewPCG(1, 2)
	const gamesPerOp = 1_000

	numGames := 0
	for b.Loop() {
		playCount(rng, gamesPerOp)
		numGames += gamesPerOp
	}
	b.ReportMetric(float64(numGames)/b.Elapsed().Seconds(), "games/s")
}

// TestPlayCount checks game counting and record keeping over a small run.
func TestPlayCount(t *testing.T) {
	rng := rand.NewPCG(1, 2)
	result := playCount(rng, 1000)
	if result.NumGames != 1000 {
		t.Errorf("NumGames = %d, want 1000", result.NumGames)
	}
	var total int64
	for _, count := range result.Counts {
		total += count
	}
	if total != 1000 {
		t.Errorf("sum of counts = %d, want 1000", total)
	}
	if len(result.Shortest) < 7 {
		t.Errorf("shortest game has %d moves, fewer than the 7 move minimum", len(result.Shortest))
	}
	if len(result.Longest) < len(result.Shortest) {
		t.Errorf("longest game (%d moves) is shorter than the shortest (%d)",
			len(result.Longest), len(result.Shortest))
	}
	if last := result.Shortest[len(result.Shortest)-1]; last.Square != 100 {
		t.Errorf("shortest game ends on square %d, want 100", last.Square)
	}
}

// TestPlayTimeMinimum checks a zero-second budget still plays the minimum batch.
func TestPlayTimeMinimum(t *testing.T) {
	rng := rand.NewPCG(3, 4)
	result := playTime(rng, 0)
	if result.NumGames != 100 {
		t.Errorf("NumGames = %d, want 100", result.NumGames)
	}
}

// TestBenchmarkResultAdd checks merging of counts, totals, and record games.
func TestBenchmarkResultAdd(t *testing.T) {
	first := BenchmarkResult{
		Counts:   gameCounts{7: 1, 30: 5},
		Elapsed:  1.5,
		NumGames: 6,
		Shortest: make(Game, 7),
		Longest:  make(Game, 30),
	}
	second := BenchmarkResult{
		Counts:   gameCounts{30: 2, 90: 1},
		Elapsed:  2.5,
		NumGames: 3,
		Shortest: make(Game, 9),
		Longest:  make(Game, 90),
	}
	combined := first.Add(second)
	if combined.NumGames != 9 {
		t.Errorf("NumGames = %d, want 9", combined.NumGames)
	}
	if combined.Elapsed != 4.0 {
		t.Errorf("Elapsed = %v, want 4.0", combined.Elapsed)
	}
	want := gameCounts{7: 1, 30: 7, 90: 1}
	if !maps.Equal(combined.Counts, want) {
		t.Errorf("Counts = %v, want %v", combined.Counts, want)
	}
	if len(combined.Shortest) != 7 || len(combined.Longest) != 90 {
		t.Errorf("records are %d and %d moves, want 7 and 90",
			len(combined.Shortest), len(combined.Longest))
	}
}

// TestBenchmarkResultAddZero checks combining with the zero value, as benchmarkParallel does.
func TestBenchmarkResultAddZero(t *testing.T) {
	var zero BenchmarkResult
	other := BenchmarkResult{
		Counts:   gameCounts{8: 2},
		NumGames: 2,
		Shortest: make(Game, 8),
		Longest:  make(Game, 8),
	}
	combined := zero.Add(other)
	if combined.NumGames != 2 || len(combined.Shortest) != 8 || len(combined.Longest) != 8 {
		t.Errorf("Add onto zero value = %+v, want copy of other", combined)
	}
}

// TestSplitCount checks entries differ by at most one and sum to the total.
func TestSplitCount(t *testing.T) {
	tests := []struct {
		total int64
		parts int
		want  []int64
	}{
		{10, 4, []int64{3, 3, 2, 2}},
		{2, 4, []int64{1, 1, 0, 0}},
		{7, 1, []int64{7}},
		{100, 3, []int64{34, 33, 33}},
	}
	for _, test := range tests {
		if got := splitCount(test.total, test.parts); !slices.Equal(got, test.want) {
			t.Errorf("splitCount(%d, %d) = %v, want %v", test.total, test.parts, got, test.want)
		}
	}

	// The arithmetic must hold right up to the int64 limit
	var sum int64
	for _, count := range splitCount(math.MaxInt64, 5) {
		sum += count
	}
	if sum != math.MaxInt64 {
		t.Errorf("splitCount(MaxInt64, 5) sums to %d, want %d", sum, int64(math.MaxInt64))
	}
}

// TestCurrencySeriesStart checks the series starts at the first value >= start.
func TestCurrencySeriesStart(t *testing.T) {
	tests := []struct {
		start int64
		want  int64
	}{
		{1, 1},
		{2, 2},
		{3, 5},
		{100, 100},
		{101, 200},
		{750, 1000},
	}
	for _, test := range tests {
		for value := range currencySeries(test.start) {
			if value != test.want {
				t.Errorf("currencySeries(%d) starts at %d, want %d", test.start, value, test.want)
			}
			break
		}
	}
}

// TestCurrencySeriesTerminates checks the series stays increasing and ends before int64 overflow.
func TestCurrencySeriesTerminates(t *testing.T) {
	var last int64
	values := 0
	for value := range currencySeries(1) {
		if value <= last {
			t.Fatalf("series is not increasing: %d follows %d", value, last)
		}
		last = value
		values++
	}
	if last != 5_000_000_000_000_000_000 {
		t.Errorf("series ends at %d, want 5e18", last)
	}
	if values != 57 {
		t.Errorf("series yielded %d values, want 57", values)
	}
}

// TestGameCountsMarshalJSON checks keys are written in ascending numeric order.
func TestGameCountsMarshalJSON(t *testing.T) {
	counts := gameCounts{100: 3, 7: 1, 20: 2}
	encoded, err := counts.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON returned error: %v", err)
	}
	want := `{"7":1,"20":2,"100":3}`
	if string(encoded) != want {
		t.Errorf("MarshalJSON = %s, want %s", encoded, want)
	}
}
