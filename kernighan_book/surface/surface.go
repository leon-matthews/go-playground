// Surface computes as SVG rendering of a 3D surface function
package main

import (
	"fmt"
	"math"
)

const (
	width, height = 600, 500            // canvas size, in pixels
	cells         = 100                 // number of grid cells
	xyrange       = 30.0                // axis ranges (-xyrange..+xyrange)
	xyscale       = width / 2 / xyrange // pixels per x or y unit
	zscale        = height * 0.4        // pixels per z unit
	angle         = math.Pi / 6         // angle of x,y axes (30 degrees)
)

var sin30, cos30 = math.Sin(angle), math.Cos(angle)

func main() {
	fmt.Printf(
		"<svg xmlns='http://www.w3.org/2000/svg' \n"+
			"\tstyle='stroke: grey; fill: white; stroke-width: 0.7' \n"+
			"\twidth='%d' height='%d'>\n", width, height,
	)

	for i := 0; i < cells; i++ {
		for j := 0; j < cells; j++ {
			// Calculate four corners for each cell in 3D space
			ax, ay := corner(i+1, j)
			bx, by := corner(i, j)
			cx, cy := corner(i, j+1)
			dx, dy := corner(i+1, j+1)

			// Output polygon in 2D space
			fmt.Printf(
				"<polygon points='%g,%g %g,%g %g,%g %g,%g' />\n",
				ax, ay, bx, by, cx, cy, dx, dy,
			)
		}

	}

	fmt.Println("</svg>")
}

func corner(i, j int) (float64, float64) {
	// Find point (x, y) at corner of cell (i, j)
	x := xyrange * (float64(i)/cells - 0.5)
	y := xyrange * (float64(j)/cells - 0.5)

	// Compute surface height z
	z := f(x, y)

	// Project (x, y, z) isometrically onto 2D SVG canvas (sx, sy)
	sx := width/2 + (x-y)*cos30*xyscale
	sy := width/2 + (x+y)*sin30*xyscale - z*zscale
	return sx, sy
}

func f(x, y float64) float64 {
	r := math.Hypot(x, y) // Distance from (0, 0)
	return math.Sin(r) / r
}
