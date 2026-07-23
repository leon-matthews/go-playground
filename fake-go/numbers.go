package fake

import "math"

// Bool returns a random boolean value.
func (f *Faker) Bool() bool {
	return f.rng.IntN(2) == 1
}

// Int returns a random integer in the inclusive range [low, high].
//
// It panics if low is not less than high.
func (f *Faker) Int(low, high int) int {
	if low >= high {
		panic("fake: range high must be greater than low")
	}
	return low + f.rng.IntN(high-low+1)
}

// Float returns a random floating-point number in the range [low, high).
//
// It panics if low is not less than high.
func (f *Faker) Float(low, high float64) float64 {
	if low >= high {
		panic("fake: range high must be greater than low")
	}
	return low + f.rng.Float64()*(high-low)
}

// Price returns a random price in cents, in the range [low, high).
//
// The bounds are given in whole currency units (for example 1.00 and 100.00)
// and the result is rounded to the nearest cent. It panics if low is not less
// than high.
func (f *Faker) Price(low, high float64) int64 {
	if low >= high {
		panic("fake: range high must be greater than low")
	}
	value := low + f.rng.Float64()*(high-low)
	return int64(math.Round(value * 100))
}
