package main

import (
	"errors"
	"math"
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
// The multiset is given as a slice of counts indexed by value, eg. counts[33]
// holding how many times 33 occurs. An error is returned if the multiset is empty.
func multisetMedian(counts []int64, mode medianMode) (float64, error) {
	var total int64
	for _, count := range counts {
		total += count
	}
	if total == 0 {
		return 0, errors.New("cannot calculate median of empty counts")
	}

	// Find median values; walking by index visits the values smallest first
	middle := float64(total+1) / 2
	lowerPos := int64(math.Floor(middle))
	upperPos := int64(math.Ceil(middle))
	var lower, upper int
	lowerFound := false
	var running int64

	for value, count := range counts {
		if count == 0 {
			continue
		}
		running += count
		if !lowerFound && running >= lowerPos {
			lower, lowerFound = value, true
		}
		if running >= upperPos {
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
