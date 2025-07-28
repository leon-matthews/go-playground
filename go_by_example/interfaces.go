package main

import (
	"fmt"
	"math"
)

func main() {
	c := circle{4.0}
	r := rectangle{12, 20}
	measure(c)
	detectCircle(c)
	measure(r)
	detectCircle(r)
}

// Circle
type circle struct {
	radius float64
}

func (c circle) area() float64 {
	return math.Pi * c.radius * c.radius
}

func (c circle) perimeter() float64 {
	return 2 * c.radius * math.Pi
}

// Rectangle
type rectangle struct {
	width, height float64
}

func (r rectangle) area() float64 {
	return r.width * r.height
}

func (r rectangle) perimeter() float64 {
	return (2 * r.width) + (2 * r.height)
}

// Function takes structure with interface `geometry`
type geometry interface {
	area() float64
	perimeter() float64
}

func detectCircle(g geometry) {
	fmt.Printf("%T%[1]v is ", g)
	if c, ok := g.(circle); ok {
		fmt.Printf("a circle, radius=%.1f\n", c.radius)
	} else {
		fmt.Println("NOT a circle")
	}
}

func measure(g geometry) {
	fmt.Printf("%T%[1]v area=%.1f perimeter=%.1f\n", g, g.area(), g.perimeter())
}
