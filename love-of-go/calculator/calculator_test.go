package calculator_test

import (
	"math"
	"testing"

	"calculator"
)

func TestAdd(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		a, b float64
		want float64
	}{
		{a: 2, b: 2, want: 4},
		{a: -1, b: -1, want: -2},
		{a: 5, b: 0, want: 5},
	}
	for _, tt := range testCases {
		got := calculator.Add(tt.a, tt.b)
		if tt.want != got {
			t.Errorf("Add(%f, %f): want %f, got %f", tt.a, tt.b, tt.want, got)
		}
	}
}

func TestDivide(t *testing.T) {
	t.Parallel()
	type testCase struct {
		a, b float64
		want float64
	}
	testCases := []testCase{
		{a: 2, b: 2, want: 1},
		{a: -1, b: -1, want: 1},
		{a: 10, b: 2, want: 5},
		{a: 1, b: 3, want: 0.333333},
	}
	for _, tt := range testCases {
		got, err := calculator.Divide(tt.a, tt.b)
		if err != nil {
			t.Fatalf("want no error for valid input, got %v", err)
		}
		if !almostEqual(tt.want, got, 0.001) {
			t.Errorf("Divide(%f, %f): want %f, got %f", tt.a,
				tt.b, tt.want, got)
		}
	}
}

func TestDivideInvalid(t *testing.T) {
	t.Parallel()
	_, err := calculator.Divide(1, 0)
	if err == nil {
		t.Error("want error for invalid input, got nil")
	}
}

func TestMultiply(t *testing.T) {
	t.Parallel()
	var want float64 = 8
	got := calculator.Multiply(4, 2)
	if want != got {
		t.Errorf("want %f, got %f", want, got)
	}
}

func TestSquareRoot(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		x    float64
		want float64
	}{
		{x: 4, want: 2},
		{x: 100, want: 10},
		{x: 2, want: 1.414},
	}

	for _, tt := range testCases {
		got, err := calculator.SquareRoot(tt.x)
		if err != nil {
			t.Fatalf("want no error for valid input, got %v", err)
		}
		if !almostEqual(tt.want, got, 0.001) {
			t.Errorf("want %f, got %f", tt.want, got)
		}
	}
}

func TestSquareRootInvalid(t *testing.T) {
	t.Parallel()
	_, err := calculator.SquareRoot(-1)
	if err == nil {
		t.Error("want error for invalid input, got nil")
	}
}

func TestSubtract(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		a, b float64
		want float64
	}{
		{a: 2, b: 2, want: 0},
		{a: -1, b: -1, want: 0},
		{a: 10, b: 2, want: 8},
	}
	for _, tt := range testCases {
		got := calculator.Subtract(tt.a, tt.b)
		if tt.want != got {
			t.Errorf("Subtract(%f, %f): want %f, got %f", tt.a,
				tt.b, tt.want, got)
		}
	}
}

func almostEqual(a, b, tolerance float64) bool {
	return math.Abs(a-b) <= tolerance
}
