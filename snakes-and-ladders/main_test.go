package main

import (
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"slices"
	"testing"
)

// TestParseDefaults checks the options produced when no arguments are given.
func TestParseDefaults(t *testing.T) {
	opts, err := parse(nil)
	if err != nil {
		t.Fatalf("parse(nil) returned error: %v", err)
	}
	if opts.jobs != 1 || opts.numGames != 0 || opts.seconds != 10 {
		t.Errorf("parse(nil) = %+v, want jobs 1, numGames 0, seconds 10", opts)
	}
	if len(opts.jsonPaths) != 0 {
		t.Errorf("parse(nil) jsonPaths = %v, want none", opts.jsonPaths)
	}
}

// TestParsePaths checks positional arguments are collected as JSON file paths.
func TestParsePaths(t *testing.T) {
	tests := []struct {
		args []string
		want []string
	}{
		{[]string{"A.json"}, []string{"A.json"}},
		{[]string{"-j=2", "A.json"}, []string{"A.json"}},
		{[]string{"A.json", "-n", "100"}, []string{"A.json"}},
		{[]string{"A.json", "B.json", "C.json"}, []string{"A.json", "B.json", "C.json"}},
	}
	for _, test := range tests {
		opts, err := parse(test.args)
		if err != nil {
			t.Errorf("parse(%v) returned error: %v", test.args, err)
			continue
		}
		if !slices.Equal(opts.jsonPaths, test.want) {
			t.Errorf("parse(%v) jsonPaths = %v, want %v", test.args, opts.jsonPaths, test.want)
		}
	}
}

// TestParseGames checks game counts parse exactly, right up to the int64 limit.
func TestParseGames(t *testing.T) {
	tests := []struct {
		args []string
		want int64
	}{
		{[]string{"-n", "100"}, 100},
		{[]string{"-n", "1_000_000"}, 1_000_000},
		{[]string{"--games", "1000000000"}, 1_000_000_000},
		{[]string{"-n", "9223372036854775807"}, math.MaxInt64},
		{[]string{"-n", "9223372036854775807", "-j=4"}, math.MaxInt64},
	}
	for _, test := range tests {
		opts, err := parse(test.args)
		if err != nil {
			t.Errorf("parse(%v) returned error: %v", test.args, err)
			continue
		}
		if opts.numGames != test.want {
			t.Errorf("parse(%v) numGames = %d, want %d", test.args, opts.numGames, test.want)
		}
	}
}

// TestNormalizeJobs checks make-style job counts are rewritten to pflag's form.
func TestNormalizeJobs(t *testing.T) {
	tests := []struct {
		args []string
		want []string
	}{
		{nil, []string{}},
		{[]string{"-j"}, []string{"-j"}},
		{[]string{"-j4"}, []string{"-j=4"}},
		{[]string{"-j", "4"}, []string{"-j=4"}},
		{[]string{"-j=4"}, []string{"-j=4"}},
		{[]string{"--jobs"}, []string{"--jobs"}},
		{[]string{"--jobs", "16"}, []string{"-j=16"}},
		{[]string{"--jobs=8"}, []string{"--jobs=8"}},
		{[]string{"-j", "--json"}, []string{"-j", "--json"}},
		{[]string{"-j4x"}, []string{"-j4x"}},
		{[]string{"-j", "-4"}, []string{"-j", "-4"}},
		{[]string{"-n", "100", "-j", "2"}, []string{"-n", "100", "-j=2"}},
		{[]string{"--", "-j", "4"}, []string{"--", "-j", "4"}},
		{[]string{"-j", "--", "4"}, []string{"-j", "--", "4"}},
	}
	for _, test := range tests {
		if got := normalizeJobs(test.args); !slices.Equal(got, test.want) {
			t.Errorf("normalizeJobs(%v) = %v, want %v", test.args, got, test.want)
		}
	}
}

// TestParseJobs checks every make-style spelling of the jobs flag.
func TestParseJobs(t *testing.T) {
	counted := [][]string{
		{"-j=2"},
		{"-j2"},
		{"-j", "2"},
		{"--jobs", "2"},
		{"--jobs=2"},
	}
	for _, args := range counted {
		opts, err := parse(args)
		if err != nil {
			t.Fatalf("parse(%v) returned error: %v", args, err)
		}
		if opts.jobs != 2 {
			t.Errorf("parse(%v) jobs = %d, want 2", args, opts.jobs)
		}
	}

	for _, args := range [][]string{{"-j"}, {"--jobs"}} {
		opts, err := parse(args)
		if err != nil {
			t.Fatalf("parse(%v) returned error: %v", args, err)
		}
		if opts.jobs != runtime.NumCPU() {
			t.Errorf("parse(%v) jobs = %d, want NumCPU (%d)", args, opts.jobs, runtime.NumCPU())
		}
	}
}

// TestParseErrors checks invalid values and combinations are rejected.
func TestParseErrors(t *testing.T) {
	tests := [][]string{
		{"-n", "100", "-s", "5"},
		{"-j=0"},
		{"-j0"},
		{"-s", "0"},
		{"-s=-3"},
		{"-n", "0"},
		{"-n=-100"},
		{"-j2", "A.json", "B.json"},
		{"-n", "100", "A.json", "B.json"},
		{"-s", "5", "A.json", "B.json"},
		{"A.json", "A.json"},
		{"./A.json", "A.json"},
		{"sub/../A.json", "B.json", "A.json"},
	}
	for _, args := range tests {
		if _, err := parse(args); err == nil {
			t.Errorf("parse(%v) expected an error, got none", args)
		}
	}
}

// TestReadResults checks file combining and the missing-output-target rule.
func TestReadResults(t *testing.T) {
	dir := t.TempDir()
	first := filepath.Join(dir, "first.json")
	second := filepath.Join(dir, "second.json")
	content := `{"counts":{"8":2},"elapsed":1.5,"wall":2.5,"num_games":2,` +
		`"shortest":[[4,14],[6,100]],"longest":[[4,14],[6,100]]}`
	if err := os.WriteFile(first, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	// The last file is the output target, so may be missing
	combined, err := readResults([]string{first, second})
	if err != nil {
		t.Fatalf("readResults returned error: %v", err)
	}
	if combined.NumGames != 2 || combined.Wall != 2.5 {
		t.Errorf("combined = %+v, want NumGames 2 and Wall 2.5", combined)
	}

	// Every readable file joins the combined result
	if err := os.WriteFile(second, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	combined, err = readResults([]string{first, second})
	if err != nil {
		t.Fatalf("readResults returned error: %v", err)
	}
	if combined.NumGames != 4 || combined.Elapsed != 3.0 || combined.Wall != 5.0 {
		t.Errorf("combined = %+v, want NumGames 4, Elapsed 3.0, Wall 5.0", combined)
	}
	if want := (gameCounts{8: 4}); !slices.Equal(combined.Counts, want) {
		t.Errorf("Counts = %v, want %v", combined.Counts, want)
	}

	// A missing file anywhere else is an error
	if _, err := readResults([]string{filepath.Join(dir, "missing.json"), second}); err == nil {
		t.Error("expected an error for a missing input file")
	}

	// As is a file that does not parse
	garbled := filepath.Join(dir, "garbled.json")
	if err := os.WriteFile(garbled, []byte("not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := readResults([]string{garbled}); err == nil {
		t.Error("expected an error for an unparseable file")
	}
}

// TestWriteResults checks an existing target is replaced and no temp file remains.
func TestWriteResults(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "results.json")
	if err := os.WriteFile(target, []byte("old content"), 0o644); err != nil {
		t.Fatal(err)
	}

	result := BenchmarkResult{
		Counts:   gameCounts{8: 2},
		Elapsed:  1.5,
		Wall:     2.5,
		NumGames: 2,
		Shortest: Game{{4, 14}, {6, 100}},
		Longest:  Game{{4, 14}, {6, 100}},
	}
	if err := writeResults(result, target); err != nil {
		t.Fatalf("writeResults returned error: %v", err)
	}

	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	var written BenchmarkResult
	if err := json.Unmarshal(data, &written); err != nil {
		t.Fatalf("written file does not parse: %v", err)
	}
	if !reflect.DeepEqual(written, result) {
		t.Errorf("written = %+v, want %+v", written, result)
	}

	// The write must leave nothing beside the target, even when replacing it
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if entry.Name() != "results.json" {
			t.Errorf("unexpected file left behind: %s", entry.Name())
		}
	}
}

// TestMerge checks the combined result is written intact to the last named file.
func TestMerge(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "target.json")
	combined := BenchmarkResult{
		Counts:   gameCounts{8: 2},
		Elapsed:  1.5,
		Wall:     2.5,
		NumGames: 2,
		Shortest: Game{{4, 14}, {6, 100}},
		Longest:  Game{{4, 14}, {6, 100}},
	}
	if code := merge(combined, []string{filepath.Join(dir, "source.json"), target}); code != 0 {
		t.Fatalf("merge returned %d, want 0", code)
	}
	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	var written BenchmarkResult
	if err := json.Unmarshal(data, &written); err != nil {
		t.Fatalf("written file does not parse: %v", err)
	}
	if !reflect.DeepEqual(written, combined) {
		t.Errorf("written = %+v, want %+v", written, combined)
	}
}

// TestComma checks thousands separators across magnitudes and signs.
func TestComma(t *testing.T) {
	tests := []struct {
		n    int64
		want string
	}{
		{0, "0"},
		{999, "999"},
		{1000, "1,000"},
		{1234567, "1,234,567"},
		{-42, "-42"},
		{-1234567, "-1,234,567"},
		{math.MaxInt64, "9,223,372,036,854,775,807"},
		{math.MinInt64, "-9,223,372,036,854,775,808"},
	}
	for _, test := range tests {
		if got := comma(test.n); got != test.want {
			t.Errorf("comma(%d) = %q, want %q", test.n, got, test.want)
		}
	}
}
