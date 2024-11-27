package clockface

import (
	"fmt"
	"io"
	"math"
	"time"
)

const secondHandLength = 90
const clockCentreX = 150
const clockCentreY = 150

const bezel = `<circle cx="150" cy="150" r="100" style="fill:#fff;stroke:#000;stroke-width:5px;"/>`
const svgEnd = `</svg>`
const svgStart = `<?xml version="1.0" encoding="UTF-8" standalone="no"?>
<!DOCTYPE svg PUBLIC "-//W3C//DTD SVG 1.1//EN" "http://www.w3.org/Graphics/SVG/1.1/DTD/svg11.dtd">
<svg xmlns="http://www.w3.org/2000/svg"
     width="100%"
     height="100%"
     viewBox="0 0 300 300"
     version="2.0">`

type Point struct {
	X, Y float64
}

// SecondHand calculates position of the end of the second hand
func SecondHand(t time.Time) Point {
	p := secondHandPoint(t)
	// Scale to the length of the hand
	p = Point{p.X * secondHandLength, p.Y * secondHandLength}

	// Flip it over the X axis, as SVG has its origin in the top-left corner
	p = Point{p.X, -p.Y}

	// Translate it to an origin of (150,150)
	p = Point{p.X + clockCentreX, p.Y + clockCentreY}
	return p
}

func secondsInRadians(t time.Time) float64 {
	return (math.Pi / (30.0 / float64(t.Second())))
}

// secondHandPoint calculates position of second hand in terms of unit circle
func secondHandPoint(t time.Time) Point {
	angle := secondsInRadians(t)
	x := math.Sin(angle)
	y := math.Cos(angle)
	return Point{x, y}
}

// SVGWriter writes an SVG representation of an analogue clock
func SVGWriter(w io.Writer, t time.Time) {
	io.WriteString(w, svgStart)
	io.WriteString(w, bezel)
	secondHand(w, t)
	io.WriteString(w, svgEnd)
}

func secondHand(w io.Writer, t time.Time) {
	p := secondHandPoint(t)
	p = Point{p.X * secondHandLength, p.Y * secondHandLength} // scale
	p = Point{p.X, -p.Y}                                      // flip
	p = Point{p.X + clockCentreX, p.Y + clockCentreY}         // translate
	fmt.Fprintf(w, `<line x1="150" y1="150" x2="%.3f" y2="%.3f" style="fill:none;stroke:#f00;stroke-width:3px;"/>`, p.X, p.Y)
}
