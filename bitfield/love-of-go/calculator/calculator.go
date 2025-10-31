// Package calculator does simple calculations
package calculator

import (
	"errors"
	"math"
)

// Add takes two numbers and returns their sum
func Add(a, b float64) float64 {
	return a + b
}

// Divide returns the ratio between two numbers.
// It is an error for the divisor, b, to be zero.
func Divide(a, b float64) (float64, error) {
	if b == 0 {
		return 0, errors.New("division by zero not allowed")
	}
	return a / b, nil
}

// Multiply returns the product of two numbers
func Multiply(a, b float64) float64 {
	return a * b
}

// SquareRoot returns the square root of a number
// It is error to provide a negative number
func SquareRoot(x float64) (float64, error) {
	if x < 0 {
		return 0, errors.New("square root cannot be negative")
	}
	return math.Sqrt(x), nil
}

// Subtract returns the difference between two numbers
func Subtract(a, b float64) float64 {
	return a - b
}
