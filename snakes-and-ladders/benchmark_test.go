package ladders

import (
	"context"
	"encoding/json"
	"math"
	"math/rand/v2"
	"reflect"
	"slices"
	"sync/atomic"
	"testing"
	"time"
)

// BenchmarkPlayGames measures playGames throughput in games per second.
func BenchmarkPlayGames(b *testing.B) {
	// Fixed seed keeps runs comparable between benchmark sessions
	rng := rand.New(rand.NewPCG(1, 2))
	const gamesPerOp = 1_000
	ctx := context.Background()

	var remaining atomic.Int64
	numGames := 0
	for b.Loop() {
		remaining.Store(gamesPerOp)
		playGames(ctx, rng, &remaining)
		numGames += gamesPerOp
	}
	b.ReportMetric(float64(numGames)/b.Elapsed().Seconds(), "games/s")
}

// TestPlayGames checks game counting and record keeping over a small run.
func TestPlayGames(t *testing.T) {
	rng := rand.New(rand.NewPCG(1, 2))
	var remaining atomic.Int64
	remaining.Store(1000)
	result := playGames(context.Background(), rng, &remaining)
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

// TestPlayGamesCancelled checks a cancelled context plays no games at all.
func TestPlayGamesCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	var remaining atomic.Int64
	remaining.Store(1_000_000)
	result := playGames(ctx, rand.New(rand.NewPCG(1, 2)), &remaining)
	if result.NumGames != 0 {
		t.Errorf("NumGames = %d, want 0", result.NumGames)
	}
}

// TestPlayGamesDeadline checks a deadline stops an effectively unbounded run.
func TestPlayGamesDeadline(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	var remaining atomic.Int64
	remaining.Store(math.MaxInt64)
	result := playGames(ctx, rand.New(rand.NewPCG(3, 4)), &remaining)
	if result.NumGames < 1 {
		t.Error("expected at least one game before the deadline")
	}
	// Generous bound; just proves the deadline ended an otherwise endless run
	if result.Elapsed > 5 {
		t.Errorf("run took %.2f seconds, deadline was 0.05", result.Elapsed)
	}
}

// TestBenchmarkParallelExactCount checks workers sharing the pool play exactly the total.
func TestBenchmarkParallelExactCount(t *testing.T) {
	result := Run(context.Background(), 4, 10_000)
	if result.NumGames != 10_000 {
		t.Errorf("NumGames = %d, want 10000", result.NumGames)
	}
	var total int64
	for _, count := range result.Counts {
		total += count
	}
	if total != 10_000 {
		t.Errorf("sum of counts = %d, want 10000", total)
	}
}

// TestBenchmarkResultAdd checks merging of counts, totals, and record games.
func TestBenchmarkResultAdd(t *testing.T) {
	first := BenchmarkResult{
		Counts:   GameCounts{7: 1, 30: 5},
		Elapsed:  1.5,
		Wall:     0.75,
		NumGames: 6,
		Shortest: make(Game, 7),
		Longest:  make(Game, 30),
	}
	second := BenchmarkResult{
		Counts:   GameCounts{30: 2, 90: 1},
		Elapsed:  2.5,
		Wall:     1.25,
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
	if combined.Wall != 2.0 {
		t.Errorf("Wall = %v, want 2.0", combined.Wall)
	}
	want := GameCounts{7: 1, 30: 7, 90: 1}
	if !slices.Equal(combined.Counts, want) {
		t.Errorf("Counts = %v, want %v", combined.Counts, want)
	}
	if len(combined.Shortest) != 7 || len(combined.Longest) != 90 {
		t.Errorf("records are %d and %d moves, want 7 and 90",
			len(combined.Shortest), len(combined.Longest))
	}
}

// TestBenchmarkResultAddZero checks combining with the zero value, as Run does.
func TestBenchmarkResultAddZero(t *testing.T) {
	var zero BenchmarkResult
	other := BenchmarkResult{
		Counts:   GameCounts{8: 2},
		NumGames: 2,
		Shortest: make(Game, 8),
		Longest:  make(Game, 8),
	}
	combined := zero.Add(other)
	if combined.NumGames != 2 || len(combined.Shortest) != 8 || len(combined.Longest) != 8 {
		t.Errorf("Add onto zero value = %+v, want copy of other", combined)
	}
}

// TestBenchmarkResultValidate checks consistent results pass and broken ones do not.
func TestBenchmarkResultValidate(t *testing.T) {
	good := BenchmarkResult{
		Counts:   GameCounts{2: 1, 3: 2},
		NumGames: 3,
		Shortest: Game{{4, 14}, {6, 100}},
		Longest:  Game{{1, 38}, {2, 40}, {6, 100}},
	}
	if err := good.Validate(); err != nil {
		t.Errorf("validate returned error for consistent result: %v", err)
	}
	if err := (BenchmarkResult{}).Validate(); err != nil {
		t.Errorf("validate returned error for empty result: %v", err)
	}

	bads := map[string]BenchmarkResult{
		"counts sum": {
			Counts:   GameCounts{2: 1},
			NumGames: 2,
			Shortest: Game{{4, 14}, {6, 100}},
			Longest:  Game{{4, 14}, {6, 100}},
		},
		"shortest length": {
			Counts:   GameCounts{2: 1},
			NumGames: 1,
			Shortest: Game{{1, 38}, {2, 40}, {6, 100}},
			Longest:  Game{{4, 14}, {6, 100}},
		},
		"longest length": {
			Counts:   GameCounts{2: 1},
			NumGames: 1,
			Shortest: Game{{4, 14}, {6, 100}},
			Longest:  Game{{1, 38}, {2, 40}, {6, 100}},
		},
	}
	for name, bad := range bads {
		if err := bad.Validate(); err == nil {
			t.Errorf("validate expected a %s error, got none", name)
		}
	}
}

// TestGameCountsMarshalJSON checks keys are written in ascending numeric order.
func TestGameCountsMarshalJSON(t *testing.T) {
	counts := GameCounts{100: 3, 7: 1, 20: 2}
	encoded, err := counts.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON returned error: %v", err)
	}
	want := `{"7":1,"20":2,"100":3}`
	if string(encoded) != want {
		t.Errorf("MarshalJSON = %s, want %s", encoded, want)
	}
}

// TestGameCountsUnmarshalJSON checks the object form decodes, and bad input is rejected.
func TestGameCountsUnmarshalJSON(t *testing.T) {
	var counts GameCounts
	if err := counts.UnmarshalJSON([]byte(`{"7":1,"20":2,"100":3}`)); err != nil {
		t.Fatalf("UnmarshalJSON returned error: %v", err)
	}
	want := GameCounts{7: 1, 20: 2, 100: 3}
	if !slices.Equal(counts, want) {
		t.Errorf("UnmarshalJSON = %v, want %v", counts, want)
	}
	for _, bad := range []string{`{"x":1}`, `{"-1":1}`, `{"7":-2}`, `[1,2]`} {
		if err := counts.UnmarshalJSON([]byte(bad)); err == nil {
			t.Errorf("UnmarshalJSON(%s) expected an error, got none", bad)
		}
	}
}

// TestBenchmarkResultRoundTrip checks a result survives JSON encode and decode.
func TestBenchmarkResultRoundTrip(t *testing.T) {
	original := BenchmarkResult{
		Counts:   GameCounts{7: 1, 9: 2},
		Elapsed:  1.25,
		Wall:     2.5,
		NumGames: 3,
		Shortest: Game{{4, 14}, {6, 100}},
		Longest:  Game{{1, 38}, {6, 44}, {6, 100}},
	}
	encoded, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal returned error: %v", err)
	}
	var decoded BenchmarkResult
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("Unmarshal returned error: %v", err)
	}
	if !reflect.DeepEqual(original, decoded) {
		t.Errorf("round trip = %+v, want %+v", decoded, original)
	}
}
