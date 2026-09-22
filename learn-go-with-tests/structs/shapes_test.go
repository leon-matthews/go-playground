package main

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestPerimeter(t *testing.T) {
	r := Rectangle{10.0, 10.0}
	got := r.Perimeter()
	want := 40.0
	if got != want {
		t.Errorf("got %.2f, want %.2f", got, want)
	}
}

func TestArea(t *testing.T) {
	tests := map[string]struct {
		shape Shape
		want  float64
	}{
		"circle": {
			Circle{10}, 314.159265,
		},
		"rectangle": {
			Rectangle{12, 6}, 72.0,
		},
		"triangle": {
			Triangle{12, 6}, 35.0,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			assertShapeArea(t, tt.shape, tt.want)
		})
	}
}

func assertShapeArea(t *testing.T, shape Shape, want float64) {
	t.Helper()
	got := shape.Area()
	option := cmpopts.EquateApprox(0, 1.0e-6)
	if !cmp.Equal(got, want, option) {
		t.Errorf("%#v got %v, want %v", shape, got, want)
	}
}
