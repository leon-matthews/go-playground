package clockface

import (
	"math"
	"time"
)

type Point struct {
	X float64
	Y float64
}

// SecondHand calculates end point of second hand
func SecondHand(t time.Time) Point {
	return Point{150, 60}
}

// secondsInRadians calculates angle of second hand
func secondsInRadians(t time.Time) float64 {
	return float64(t.Second()) * (math.Pi / 30.0)
}
