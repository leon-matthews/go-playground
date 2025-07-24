package main

import "fmt"

func main() {
	// Zero values
	var a [5]int
	fmt.Println("empty:", a)

	// Set & get using index
	a[3] = 100
	fmt.Println("set:", a)
	fmt.Println("get:", a[3])
	fmt.Println("len:", len(a))

	// Literal declaration
	b := [5]int{1, 2, 3, 4, 5}
	fmt.Println("literal:", b)

	// Let the compiler calculate count
	c := [...]int{5, 4, 3, 2, 1}
	fmt.Println("ellipsis", c)

	// Sparse
	d := [...]int{100, 9: 400, 500}
	fmt.Println("sparse:", d)

	// 2D
	var twoD [2][3]int
	fmt.Println("2D:", twoD)

	// Sparse 2D. Can only elide the outer length
	twoD = [...][3]int{
		{1, 2, 3},
		{4, 5, 6},
	}
	fmt.Println("2D:", twoD)
}
