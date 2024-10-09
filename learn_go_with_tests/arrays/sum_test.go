package main

import "testing"

func TestSum(t *testing.T) {
	t.Run("array of five numbers", func(t *testing.T) {
		numbers := []int{1, 2, 3, 4, 5}

		got := Sum(numbers)
		want := 15

		if got != want {
			t.Errorf("got %d but wanted %d, given %v", got, want, numbers)
		}
	})

	t.Run("slice of any size", func(t *testing.T) {
		numbers := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

		got := Sum(numbers)
		want := 55

		if got != want {
			t.Errorf("got %d want %d, given %v", got, want, numbers)
		}
	})
}
