// Chapter 2, Predeclared Types and Declarations

package main

import "fmt"

const Value = 42

func main() {
	TypeConversion()
	NarrowingConversion()
	Declarations()
	Exercise1()
	Exercise2()
	Exercise3()
}

// Explicit type conversion required!
func TypeConversion() {
	var x int = 10
	var y float64 = 30.2
	var sum1 float64 = float64(x) + y
	var sum2 int = x + int(y)
	fmt.Println(sum1, sum2)
}

// Experiment with narrowing conversion
func NarrowingConversion() {
	var x uint = 300
	var y uint8 = uint8(x)
	// Prints "300 44" - it's doing modulo conversion!
	fmt.Println(x, y)
}

// Various declaration styles
func Declarations() {
	var a int = 65        // Keyword `var` plus explicit type
	var b = 66            // Literal's default type
	var c int             // Declare & assign zero value
	var d, e int = 68, 69 // Multiple variables with same type
	var f, g int          // Multiple zero values with same type
	var (                 // Declaration list
		h int = 72
		i     = 73
	)

	// Short declaration and assignment format, only available within
	// functions, and uses type inference. May shadow existing variables.
	j := 74

	// It is a compile-time error to declare then not at least read variables!
	fmt.Println(a, b, c, d, e, f, g, h, i, j)
}

func Exercise1() {
	var i int = 20
	var f = float64(i)
	fmt.Println(i, f)
}

func Exercise2() {
	var i int = Value
	var f float64 = Value
	fmt.Println(i, f)
}

func Exercise3() {
	// Max size for type
	var b byte = 255
	var smallI int32 = 2_147_483_647
	var bigI uint64 = 18_446_744_073_709_551_615
	fmt.Println(b, smallI, bigI)

	// Force overflow
	b += 1      // 0
	smallI += 1 // -2_147_483_648
	bigI += 1   // 0
	fmt.Println(b, smallI, bigI)
}
