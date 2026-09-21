package main

import (
	"slices"
	"testing"
)

func TestSum(t *testing.T) {
	numbers := []int{1, 2, 3, 4, 5}
	got := Sum(numbers)
	want := 15
	if got != want {
	    t.Errorf("got %d, want %d, given %v", got, want, numbers)
    }
}

func TestSumAll(t *testing.T) {
	t.Run("one slice", func(t *testing.T) {
		got := SumAll([]int{1, 1, 1})
		want := []int{3}
		assertEqualSums(t, got, want)
	})

	t.Run("two slices", func(t *testing.T) {
		got := SumAll([]int{1, 2}, []int{0, 9})
		want := []int{3, 9}
		assertEqualSums(t, got, want)
	})
}

func TestSumTail(t *testing.T) {
	t.Run("two slices", func(t *testing.T) {
		got := SumTail([]int{1, 2}, []int{3, 4})
		want := []int{2, 4}
		assertEqualSums(t, got, want)
	})

	t.Run("empty", func(t *testing.T) {
		got := SumTail([]int{})
		want := []int{0}
		assertEqualSums(t, got, want)
	})
}

func assertEqualSums(t *testing.T, got, want []int) {
    t.Helper()
	if !slices.Equal(got, want) {
		t.Errorf("got %d, want %d", got, want)
	}
}
