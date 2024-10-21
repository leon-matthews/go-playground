package main

import (
	"fmt"
	"time"
)

// During compilation, constants have at least 256 bits of precision
// These two were found in the actual Go language standard library
const (
	E       = 2.71828182845904523536028747135266249775724709369995957496696763
	OneDay  = time.Duration(time.Hour * 24)
	NoDelay = time.Duration(0) // Up to 290 years worth of nanoseconds, int64
	Pi      = 3.14159265358979323846264338327950288419716939937510582097494459
)

func main() {
	durations()
	unusualPrecision()
	basicIota()
	iotaExpressions()
}

// Type and expression copied down list
func basicIota() {
	type Weekday int

	const (
		Sunday Weekday = iota
		Monday
		Tuesday
		Wednesday
		Thursday
		Friday
		Saturday
	)

	fmt.Printf("%T %[1]v\n", Friday)
}

func durations() {
	fmt.Printf("%T %[1]v\n", OneDay)
	fmt.Printf("%T %[1]v\n", NoDelay)
	fmt.Printf("%T %[1]v\n", time.Minute)
}

// Use constant expressions to create bit-flags
func iotaExpressions() {
	type flags byte

	const (
		FlagUp flags = 1 << iota
		FlagBroadcast
		FlagLoopback
		FlagPointToPoint
		FlagMulticast
	)
	fmt.Printf("%T %[1]v 0b%08[1]b\n", FlagLoopback)

	connection := FlagUp | FlagLoopback | FlagMulticast
	fmt.Printf("%T %[1]v 0b%08[1]b\n", connection)
}

func unusualPrecision() {
	var smaller float32 = Pi
	var bigger float64 = Pi
	defaultSize := Pi
	fmt.Println(smaller)
	fmt.Println(bigger)
	fmt.Println(defaultSize)
}
