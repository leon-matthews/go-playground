// Struct embedding and anonymous fields
package main

import (
	"fmt"
)

func main() {
	var w Wheel

	// Embedding structs makes dot-notation very clean
	w.Radius = 40
	w.X = 8
	w.Y = 8
	w.Spokes = 32
	fmt.Println(w)

	// But not the literal syntax

	// Error: unknown field Radius in struct literal of type Wheel
	//~ w2 := Wheel{Radius: 10}
	w2 := Wheel{
		Circle: Circle{
			Point: Point{
				X: 8,
				Y: 8,
			},
			Radius: 40,
		},
		Spokes: 32,
	}
	fmt.Println(w2)

	// Compare
	fmt.Println(w == w2)
}

type Point struct {
	X int
	Y int
}

// Types with no name in a struct are called *anonymous fields*...
type Circle struct {
	Point
	Radius int
}

// ...We can say that `Circle` is 'embedded' within `Wheel`
type Wheel struct {
	Circle
	Spokes int
}
