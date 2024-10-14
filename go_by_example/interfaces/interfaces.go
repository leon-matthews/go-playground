package main

import (
	"fmt"
	"math"
)

type geometry interface {
	area() float64
	perimeter() float64
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
func measure(g geometry) {
    fmt.Println(g)
    fmt.Println(g.area())
    fmt.Println(g.perimeter())
}

func main() {
	c := circle{4.0}
	r := rectangle{12, 20}
	measure(c)
	measure(r)
}
