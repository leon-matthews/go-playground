package main

import (
	"fmt"
	"math/rand/v2"
)

func main() {
	// Int in range [0, 100)
	fmt.Print(rand.IntN(100), ",")
	fmt.Print(rand.IntN(100))
	fmt.Println()

	// 0.0 >= float64 < 1.0
	fmt.Println(rand.Float64())

	// 5.0 >= f < 10.0
	fmt.Print((rand.Float64()*5)+5, ",")
	fmt.Print((rand.Float64() * 5) + 5)
	fmt.Println()

	// Known seed for predictable test output - mostly.
	s2 := rand.NewPCG(42, 1024) // Two uint64 numbers
	r2 := rand.New(s2)
	fmt.Print(r2.IntN(100), ",") // 94
	fmt.Print(r2.IntN(100))      // 49
	fmt.Println()
}
