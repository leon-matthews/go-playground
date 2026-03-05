package main

import (
	"flag"
	"log"
	"time"
)

var (
	max     = flag.Int("max", 0, "largest exponent to test, eg. 3000")
	verbose = flag.Bool("verbose", false, "also print non-prime exponents")
)

func main() {
	flag.Parse()
	err := run(*max, *verbose)
	if err != nil {
		log.Fatal(err)
	}
}

// run finds Mersenne primes
// These are prime numbers in the form of 2^e-1 where e is also a prime.
// Set max to zero to never exit
// Set verbose to true to also log candidates that fail prime checking
func run(max int, verbose bool) error {
	start := time.Now()
	for n := range primes() {
		if n.Int64() > int64(max) {
			log.Printf("Finished checking exponents up to %d in %v", max, time.Since(start))
			return nil
		}

		prime, err := mersenne(n)
		if err != nil {
			if verbose {
				log.Printf("2^%v-1 is NOT a prime", n)
			}
			continue
		}
		log.Printf("2^%v-1 is a prime: %v", n, prime)
	}
	return nil
}
