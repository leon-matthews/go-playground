// A method can be passed like a closure - the receiver is closed-over
package main

import (
	"fmt"
	"math"
)

func main() {
	p := Point{1, 2}
	q := Point{4, 6}

	// Call method
	d1 := p.Distance(q)
	fmt.Println(d1) // 5.0

	// Method value
	distanceP := p.Distance
	d2 := distanceP(q)
	fmt.Println(d2) // 5.0

	// Changing p doesn't change closed-over value
	p = Point{2, 2}
	fmt.Println(distanceP(q))	// 5.0
	fmt.Println(p.Distance(q))	// 4.4721

	// BUT not for a pointer receiver, as then the closed-over value is the
	// pointer to the actual value.
}

type Point struct {
	x, y float64
}

func (p Point) Distance(q Point) float64 {
	return math.Hypot(q.x-p.x, q.y-p.y)
}
