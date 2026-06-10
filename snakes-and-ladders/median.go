package main

import (
	"errors"
	"maps"
	"math"
	"slices"
)

// medianMode selects which flavour of median multisetMedian returns.
type medianMode int

const (
	// medianMean uses the 'mean of middle two' method, like Python's statistics package.
	medianMean medianMode = iota
	// medianLow uses the low median instead; the value returned will be in the set.
	medianLow
	// medianHigh is as per medianLow, but the high median.
	medianHigh
)

// multisetMedian efficiently determines the median of a multiset of integers.
//
// The multiset is given as a mapping of values against how many times each
// value occurs. An error is returned if the multiset is empty.
func multisetMedian(counts map[int]int, mode medianMode) (float64, error) {
	total := 0
	for _, count := range counts {
		total += count
	}
	if total == 0 {
		return 0, errors.New("cannot calculate median of empty counts")
	}

	// Find median values
	middle := float64(total+1) / 2
	lowerPos := int(math.Floor(middle))
	upperPos := int(math.Ceil(middle))
	var lower, upper int
	lowerFound := false
	count := 0

	for _, value := range slices.Sorted(maps.Keys(counts)) {
		count += counts[value]
		if !lowerFound && count >= lowerPos {
			lower, lowerFound = value, true
		}
		if count >= upperPos {
			upper = value
			break
		}
	}

	// Median present, no interpolation needed
	if lowerPos == upperPos {
		return float64(lower), nil
	}

	// Pick your flavour! Middle-of-two, High, or Low median.
	switch mode {
	case medianHigh:
		return float64(upper), nil
	case medianLow:
		return float64(lower), nil
	default:
		return float64(lower+upper) / 2, nil
	}
}
