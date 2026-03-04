package main

import (
	"log"
)

func main() {
	for n := range primes() {
		prime, err := mersenne(n)
		if err != nil {
			continue
		}
		log.Printf("2^%v-1 is a prime number: %v", n, prime)
	}
}
