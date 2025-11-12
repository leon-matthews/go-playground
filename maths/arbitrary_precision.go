package main

import (
	"fmt"
	"math/big"
)

func main() {
	mersennePrimes()
}

// mersennePrimes showcases the stdlib's "signed multi-precision integers"
// It's implemented as a slice of uint values
func mersennePrimes() {
	fmt.Println("Caluculating Mersenne Primes!")
	x := big.NewInt(2)
	y := big.NewInt(2)
	candidate := big.NewInt(0)
	minusOne := big.NewInt(-1)
	plusOne := big.NewInt(1)

	for {
		candidate.Exp(x, y, nil)
		candidate.Add(candidate, minusOne)
		isPrime := candidate.ProbablyPrime(32)

		fmt.Printf("2^%v-1 is ", y)
		if isPrime {
			fmt.Printf("a prime: %v\n", candidate)
		} else {
			fmt.Println("not a prime")
		}

		y = y.Add(y, plusOne)
	}
}
