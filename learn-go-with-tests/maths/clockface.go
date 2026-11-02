package clockface

import (
	"math"
	"time"
)

type Point struct {
	X float64
	Y float64
}

// secondsInRadians calculates angle of second hand for the given time
func secondsInRadians(t time.Time) float64 {
	return float64(t.Second()) * (math.Pi / 30.0)
}

// minutesInRadians calculates angle of minute hand for the given time
func minutesInRadians(t time.Time) float64 {
	return math.Pi/(30/float64(t.Minute())) + secondsInRadians(t)/60
}

// secondHandPoint returns points on unit circle
func secondHandPoint(t time.Time) Point {
	return angleToPoint(secondsInRadians(t))
}

func minuteHandPoint(t time.Time) Point {
	return angleToPoint(minutesInRadians(t))
}

func angleToPoint(angle float64) Point {
	x := math.Sin(angle)
	y := math.Cos(angle)
	return Point{x, y}
}
