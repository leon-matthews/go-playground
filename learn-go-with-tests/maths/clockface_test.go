package clockface

import (
	"math"
	"reflect"
	"testing"
	"time"

	"github.com/alecthomas/assert/v2"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestSecondsInRadians(t *testing.T) {
	tests := map[string]struct {
		time  time.Time
		angle float64
	}{
		"midnight":   {date(0, 0, 0), 0},
		"30 seconds": {date(0, 0, 30), math.Pi},
		"45 seconds": {date(0, 0, 45), math.Pi / 2 * 3},
		"7 seconds":  {date(0, 0, 7), math.Pi / 30 * 7},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := secondsInRadians(tt.time)
			assertAlmostEqual(t, tt.angle, got)
		})
	}
}

// func assertAlmostEqual(tb testing.TB, want, got float64) {
// 	tb.Helper()
// 	const tolerance = 1e-9
// 	if math.Abs(want-got) > tolerance {
// 		tb.Errorf("want %v, got %v (tolerance %v)", want, got, tolerance)
// 	}
// }

func date(hours, minutes, seconds int) time.Time {
	return time.Date(312, time.October, 14, hours, minutes, seconds, 0, time.UTC)
}

// assertAlmostEqual compares two floating-point numbers within reasonable precision.
//
// The tolerance is increased automatically for float32 values, and when the numbers to
// be compared are larger. Equal infinities pass, unequal infinities and NaN always fail.
func assertAlmostEqual[T ~float32 | ~float64](tb testing.TB, want, got T) {
	tb.Helper()

	// Determine defaults for relative and absolute tolerance, far above float64
	// rounding noise, far below the size of a typical bug.
	relative, absolute := 1e-9, 1e-12

	// Relax defaults for float32 values
	// (Use Kind() to handle named types)
	if reflect.TypeFor[T]().Kind() == reflect.Float32 {
		relative, absolute = 1e-4, 1e-6
	}

	// EquateApprox skips defined types such as celsius, so compare as float64
	approx := cmpopts.EquateApprox(relative, absolute)
	if !cmp.Equal(float64(want), float64(got), approx) {
		tb.Errorf("want %v, got %v", want, got)
	}
}

func TestAssertAlmostEqual(t *testing.T) {
	inf, nan := math.Inf(1), math.NaN()
	tests := map[string]struct {
		want, got float64
		wantFail  bool
	}{
		"equal infinities":     {inf, inf, false},
		"opposite infinities":  {inf, -inf, true},
		"infinity vs finite":   {inf, 1, true},
		"NaN":                  {nan, nan, true},
		"rounding noise":       {0, math.Sin(math.Pi), false},
		"large and close":      {1e6, 1e6 + 1e-7, false},
		"near zero and far":    {0, 1e-6, true},
		"small relative error": {1, 1.001, true},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			tb := &failureRecorder{TB: t}
			assertAlmostEqual(tb, tt.want, tt.got)
			assert.Equal(t, tt.wantFail, tb.failed)
		})
	}
}

func TestAssertAlmostEqualTypes(t *testing.T) {
	type celsius float64

	t.Run("float32 is relaxed", func(t *testing.T) {
		tb := &failureRecorder{TB: t}
		assertAlmostEqual(tb, float32(1), float32(1+1e-6))
		assert.False(t, tb.failed)
	})

	t.Run("float32 still fails", func(t *testing.T) {
		tb := &failureRecorder{TB: t}
		assertAlmostEqual(tb, float32(1), float32(1.01))
		assert.True(t, tb.failed)
	})

	t.Run("defined type passes", func(t *testing.T) {
		tb := &failureRecorder{TB: t}
		assertAlmostEqual(tb, celsius(1), celsius(1+1e-12))
		assert.False(t, tb.failed)
	})

	t.Run("defined type fails", func(t *testing.T) {
		tb := &failureRecorder{TB: t}
		assertAlmostEqual(tb, celsius(1), celsius(1.001))
		assert.True(t, tb.failed)
	})
}

// failureRecorder is a testing.TB that records failures instead of reporting them.
type failureRecorder struct {
	testing.TB
	failed bool
}

// Helper does nothing, as there is no failure output to attribute.
func (*failureRecorder) Helper() {}

// Errorf records the failure without failing the enclosing test.
func (r *failureRecorder) Errorf(string, ...any) {
	r.failed = true
}
