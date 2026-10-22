package clockface

import (
	"math"
	"time"
)

const (
	secondHandLength = 90
	clockCentreX     = 150
	clockCentreY     = 150
)

type Point struct {
	X float64
	Y float64
}

// SecondHand calculates end point of second hand
func SecondHand(t time.Time) Point {
	p := secondHandPoint(t)
	p = Point{p.X * secondHandLength, p.Y * secondHandLength} // Scale to length of hand
	p = Point{p.X, -p.Y}                                      // Flip x-axis as SVG has origin in top-left corner
	p = Point{p.X + clockCentreX, p.Y + clockCentreY}         // Translate to clock centre position
	return p
}

// secondsInRadians calculates angle of second hand
func secondsInRadians(t time.Time) float64 {
	return float64(t.Second()) * (math.Pi / 30.0)
}

// secondHandPoint returns points on unit circle
func secondHandPoint(t time.Time) Point {
	angle := secondsInRadians(t)
	x := math.Sin(angle)
	y := math.Cos(angle)

	return Point{x, y}
}
