package structs

import "testing"

func TestArea(t *testing.T) {
	// Check output of a shape's Area() method
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

func TestPerimeter(t *testing.T) {
	rectangle := Rectangle{5.0, 10.0}
	got := rectangle.Perimeter()
	want := 30.0

	if got != want {
		t.Errorf("got %.2f want %.2f", got, want)
	}
}
