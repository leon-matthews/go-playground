package main

import "fmt"

type rect struct {
	width, height int
}

func (r rect) area() int {
	return r.width * r.height
}

// Use a pointer receiver type to avoid copying on method calls or
// to allow the method to mutate the receiving struct.
func (r *rect) perimeter() int {
	return (2 * r.width) + (2 * r.height)
}


func main() {
	r := rect{20, 30}
	fmt.Println(r, "has area", r.area(), "and perimeter", r.perimeter())

	// Go automatically converts values and pointers for method calls
	r2 := &r
	fmt.Println(r2, "has area", r2.area(), "and perimeter", r2.perimeter())
}
