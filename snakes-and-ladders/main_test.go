package main

import (
	"math"
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
	want := options{jobs: 1, json: false, numGames: 0, seconds: 10}
	if opts != want {
		t.Errorf("parse(nil) = %+v, want %+v", opts, want)
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
		{"positional"},
		{"-j=0"},
		{"-j0"},
		{"-s", "0"},
		{"-s=-3"},
		{"-n", "0"},
		{"-n=-100"},
	}
	for _, args := range tests {
		if _, err := parse(args); err == nil {
			t.Errorf("parse(%v) expected an error, got none", args)
		}
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
