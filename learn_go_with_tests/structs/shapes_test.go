package structs

import (
	"math"
	"testing"
)

func assertAlmostEqual(t *testing.T, shape Shape, got, want float64) {
	t.Helper()
	epsilon := 1.0e-6
	if math.Abs(got-want) > epsilon {
		t.Errorf("%#v got %.6f want %.6f", shape, got, want)
	}
}

func TestArea(t *testing.T) {
	// Helper to check output of any `Shape`s Area() method
	checkArea := func(t testing.TB, shape Shape, want float64) {
		t.Helper()
		got := shape.Area()
		if got != want {
			t.Errorf("got %g want %g", got, want)
		}
	}

	t.Run("circles", func(t *testing.T) {
		circle := Circle{10}
		checkArea(t, circle, 314.1592653589793)
	})

	t.Run("rectangles", func(t *testing.T) {
		rectangle := Rectangle{5.0, 10.0}
		checkArea(t, rectangle, 50.0)
	})
}

func TestAreaTableDriven(t *testing.T) {
	// Anonymous struct, slice of structs with two fields
	areaTests := []struct {
		name  string
		shape Shape
		want  float64
	}{
		{name: "Circle", shape: Circle{Radius: 10}, want: 314.259265},
		{name: "Rectangle", shape: Rectangle{Width: 12, Height: 6}, want: 72.0},
		{name: "Triangle", shape: Triangle{Base: 12, Height: 6}, want: 36.0},
	}

	// Wrap each case in `t.Run`
	for _, tt := range areaTests {
		t.Run(tt.name, func(t *testing.T) {
			assertAlmostEqual(t, tt.shape, tt.shape.Area(), tt.want)
		})
	}
}

func TestPerimeter(t *testing.T) {
	rectangle := Rectangle{5.0, 10.0}
	got := rectangle.Perimeter()
	want := 30.0

	if got != want {
		t.Errorf("got %.2f want %.2f", got, want)
	}
}
