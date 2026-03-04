package main

import (
	"fmt"
	"iter"
	"math/big"
)

// primes returns an iterator over prime numbers
// It uses [big.Int.ProbablyPrime] which is 100% accurate less than 2^64
func primes() iter.Seq[*big.Int] {
	return func(yield func(*big.Int) bool) {
		if !yield(big.NewInt(2)) {
			return
		}
		n := big.NewInt(3)
		for {

			if n.ProbablyPrime(0) {
				if !yield(n) {
					return
				}
			}
			n.Add(n, big.NewInt(2))
		}
	}
}

// Mersenne attempts to verify the Mersenne prime with the given exponent
// This is a prime in the form 2^exponent - 1, see https://en.wikipedia.org/wiki/Mersenne_prime.
// The calculated prime integer is returned, if prime, otherwise an error is set.
func mersenne(exponent *big.Int) (*big.Int, error) {
	candidate := big.NewInt(0)
	candidate = candidate.Exp(big.NewInt(2), exponent, nil)
	candidate.Add(candidate, big.NewInt(-1))
	if !candidate.ProbablyPrime(32) {
		return nil, fmt.Errorf("2^%v - 1 is not a Mersenne prime", exponent)
	}
	return candidate, nil
}
