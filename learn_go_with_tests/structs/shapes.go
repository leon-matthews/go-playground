package structs

import "math"

type Circle struct {
	Radius float64
}

// Calculate area of a circle
func (c Circle) Area() float64 {
	return math.Pi * c.Radius * c.Radius
}

type Rectangle struct {
	Width  float64
	Height float64
}

// Calculate area of rectangle
func (r Rectangle) Area() float64 {
	return r.Width * r.Height
}

// Calculate perimeter of rectangle
func (r Rectangle) Perimeter() float64 {
	return 2 * (r.Width + r.Height)
}
