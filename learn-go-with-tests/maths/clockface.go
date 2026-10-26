package clockface

import (
	"fmt"
	"io"
	"math"
	"time"
)

const (
	secondHandLength = 90
	clockCentreX     = 150
	clockCentreY     = 150

	bezel    = `<circle cx="150" cy="150" r="100" style="fill:#fff;stroke:#000;stroke-width:5px;"/>`
	svgStart = `<?xml version="1.0" encoding="UTF-8" standalone="no"?>
	<!DOCTYPE svg PUBLIC "-//W3C//DTD SVG 1.1//EN" "http://www.w3.org/Graphics/SVG/1.1/DTD/svg11.dtd">
	<svg xmlns="http://www.w3.org/2000/svg"
     width="100%"
     height="100%"
     viewBox="0 0 300 300"
     version="2.0">`
	svgEnd = `</svg>`
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

func SVGWriter(w io.Writer, now time.Time) {
	h := SecondHand(now)
	io.WriteString(w, svgStart)
	io.WriteString(w, bezel)
	io.WriteString(w, secondHandTag(h))
	io.WriteString(w, svgEnd)
}

func secondHandTag(p Point) string {
	return fmt.Sprintf(`<line x1="150" y1="150" x2="%.3f" y2="%.3f" style="fill:none;stroke:#f00;stroke-width:3px;"/>`, p.X, p.Y)
}
