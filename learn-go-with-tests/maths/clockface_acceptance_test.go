package clockface_test

import (
	"bytes"
	"encoding/xml"
	"testing"
	"time"

	"github.com/alecthomas/assert/v2"

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

	Circle struct {
		Text  string `xml:",chardata"`
		Cx    string `xml:"cx,attr"`
		Cy    string `xml:"cy,attr"`
		R     string `xml:"r,attr"`
		Style string `xml:"style,attr"`
	} `xml:"circle"`

	Line []struct {
		Text  string `xml:",chardata"`
		X1    string `xml:"x1,attr"`
		Y1    string `xml:"y1,attr"`
		X2    string `xml:"x2,attr"`
		Y2    string `xml:"y2,attr"`
		Style string `xml:"style,attr"`
	} `xml:"line"`
}

func TestSecondHand(t *testing.T) {
	t.Run("at midnight", func(t *testing.T) {
		now := time.Date(1337, time.January, 1, 0, 0, 0, 0, time.UTC)
		want := clockface.Point{X: 150, Y: 150 - 90}
		got := clockface.SecondHand(now)
		assert.Equal(t, want, got)
	})

	t.Run("at 30 seconds", func(t *testing.T) {
		now := time.Date(1337, time.January, 1, 0, 0, 30, 0, time.UTC)
		want := clockface.Point{X: 150, Y: 150 + 90}
		got := clockface.SecondHand(now)
		clockface.AssertAlmostEqual(t, want.X, got.X)
		clockface.AssertAlmostEqual(t, want.Y, got.Y)
	})
}

func TestSVGWriter(t *testing.T) {
	t.Run("at midnight", func(t *testing.T) {
		now := time.Date(2027, time.January, 1, 0, 0, 0, 0, time.UTC)
		buf := bytes.Buffer{}
		svg := SVG{}

		clockface.SVGWriter(&buf, now)
		xml.Unmarshal(buf.Bytes(), &svg)

		// Just one hand (for now)
		assert.Equal(t, 1, len(svg.Line))

		for _, line := range svg.Line {
			// Check end of second hand
			assert.Equal(t, "150.000", line.X2)
			assert.Equal(t, "60.000", line.Y2)
		}
	})
}
