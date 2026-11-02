package clockface_test

import (
	"bytes"
	"encoding/xml"
	"slices"
	"testing"
	"time"

	"local.dev/clockface"
)

// SVG provides just the fields we need to unmarshal our clockface for testing
type SVG struct {
	XMLName xml.Name `xml:"svg"`
	Text    string   `xml:",chardata"`
	Xmlns   string   `xml:"xmlns,attr"`
	Width   string   `xml:"width,attr"`
	Height  string   `xml:"height,attr"`
	ViewBox string   `xml:"viewBox,attr"`
	Version string   `xml:"version,attr"`
	Circle  Circle   `xml:"circle"`
	Line    []Line   `xml:"line"`
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

func TestSVGWriterSecondHand(t *testing.T) {
	tests := map[string]struct {
		time time.Time
		line Line
	}{
		"midnight": {
			date(0, 0, 0),
			Line{150, 150, 150, 60},
		},
		"30 seconds": {
			date(0, 0, 30),
			Line{150, 150, 150, 240},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			buf := bytes.Buffer{}
			clockface.SVGWriter(&buf, tt.time)

			svg := SVG{}
			xml.Unmarshal(buf.Bytes(), &svg)

			if !containsLine(tt.line, svg.Line) {
				t.Errorf("Expected second hand at %+v in SVG lines %+v", tt.line, svg.Line)
			}
		})
	}
}

func TestSVGWriterMinuteHand(t *testing.T) {
	tests := map[string]struct {
		time time.Time
		line Line
	}{
		"midnight": {
			date(0, 0, 0),
			Line{150, 150, 150, 70},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			buf := bytes.Buffer{}
			clockface.SVGWriter(&buf, tt.time)

			svg := SVG{}
			xml.Unmarshal(buf.Bytes(), &svg)

			if !containsLine(tt.line, svg.Line) {
				t.Errorf("Expected minute hand at %+v in SVG lines %+v", tt.line, svg.Line)
			}
		})
	}
}

func date(hours, minutes, seconds int) time.Time {
	return time.Date(312, time.October, 14, hours, minutes, seconds, 0, time.UTC)
}

func containsLine(needle Line, haystack []Line) bool {
	return slices.Contains(haystack, needle)
}
