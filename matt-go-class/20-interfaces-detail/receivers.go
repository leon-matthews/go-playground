// Pointer vs value receivers
package main

import (
	"fmt"
)

func main() {
	fmt.Println("Pointer vs value receivers")
	p1 := new(Point)  // *Point at (0, 0)
	p2 := Point{1, 1} // Point at (1, 1)

	// Go performs automatic conversion...
	p1.OffsetOf(p2) // same as (*p1).OffsetOf(p2)

	// ...but address-of '&' only works if object is 'addressable', ie. has a name.
	p2.Add(3, 4) // same as (&p2).Add(3, 4)

	var p Point
	p.Add(1, 2) // okay, because &p is allowed
	// not allowed, because &Point{1, 1} is not legal:
	// Point{1, 1}.Add(2, 3)
}

type Point struct {
	x, y float32
}

// Add takes pointer reciever, changes actual value
func (p *Point) Add(x, y float32) {
	p.x, p.y = p.x+x, p.y+y
}

// OffsetOf takes copy, cannot change value
func (p Point) OffsetOf(p1 Point) (x float32, y float32) {
	x, y = p.x-p1.x, p.y-p1.y
	return
}
