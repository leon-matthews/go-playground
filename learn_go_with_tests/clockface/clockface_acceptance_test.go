package clockface_test

import (
	"bytes"
	"encoding/xml"
	"testing"
	"time"

	"clockface"
)

type SVG struct {
	XMLName xml.Name `xml:"svg"`
	Xmlns   string   `xml:"xmlns,attr"`
	Width   string   `xml:"width,attr"`
	Height  string   `xml:"height,attr"`
	ViewBox string   `xml:"viewBox,attr"`
	Version string   `xml:"version,attr"`
	Circle  Circle   `xml:"circle"`
	Lines   []Line   `xml:"line"`
}

type Circle struct {
	Cx float64 `xml:"cx,attr"`
	Cy float64 `xml:"cy,attr"`
	R  float64 `xml:"r,attr"`
}

type Line struct {
	X1 float64 `xml:"x1,attr"`
	Y1 float64 `xml:"y1,attr"`
	X2 float64 `xml:"x2,attr"`
	Y2 float64 `xml:"y2,attr"`
}

func containsLine(l Line, ls []Line) bool {
	for _, line := range ls {
		if line == l {
			return true
		}
	}
	return false
}

func TestSecondHandAtMidnight(t *testing.T) {
	moment := time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC)
	// Second hand should be pointing UP
	want := clockface.Point{X: 150, Y: 150 - 90}
	got := clockface.SecondHand(moment)
	if got != want {
		t.Errorf("got %v want %v", got, want)
	}
}

func TestSecondHandAtThirtySeconds(t *testing.T) {
	moment := time.Date(2024, time.January, 1, 0, 0, 30, 0, time.UTC)
	// Second hand should be pointing DOWN
	want := clockface.Point{X: 150, Y: 150 + 90}
	got := clockface.SecondHand(moment)
	if got != want {
		t.Errorf("got %v want %v", got, want)
	}
}

func TestSVGWriterAtMidnight(t *testing.T) {
	tm := time.Date(1337, time.January, 1, 0, 0, 0, 0, time.UTC)

	b := bytes.Buffer{}
	clockface.SVGWriter(&b, tm)

	svg := SVG{}
	xml.Unmarshal(b.Bytes(), &svg)

	want := Line{150, 150, 150, 60}

	if !containsLine(want, svg.Lines) {
		t.Errorf("Expected to find the second hand line %+v, in the SVG lines %+v", want, svg.Lines)
	}
}
