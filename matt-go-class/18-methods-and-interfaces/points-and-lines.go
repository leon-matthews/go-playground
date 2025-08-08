package main

import (
	"fmt"
	"math"
)

type Point struct {
	X, Y float64
}

type Line struct {
	Begin, End Point
}

func (l Line) Distance() float64 {
	return math.Hypot(l.End.X-l.Begin.X, l.End.Y-l.Begin.Y)
}

type Path []Point

func (p Path) Distance() float64 {
	var sum float64
	for i := 1; i < len(p); i++ {
		// Can't take address of Line literal, so couldn't have pointer
		// receiver on Line.Distance()
		sum += Line{p[i-1], p[i]}.Distance()
	}
	return sum
}

func main() {
	side := Line{Point{1, 2}, Point{4, 6}}
	fmt.Println(side, side.Distance())

	perimeter := Path{{1, 1}, {5, 1}, {5, 4}, {1, 1}}
	fmt.Println(perimeter, perimeter.Distance())
}
