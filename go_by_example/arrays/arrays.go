package main

import "fmt"

func main() {
	// Zero values
	var a [5]int
	fmt.Println("empty:", a)

	// Set & get using index
	a[4] = 100
	fmt.Println("set:", a)
	fmt.Println("get:", a[4])
	fmt.Println("len:", len(a))

	// Literal declaration
	b := [5]int{1, 2, 3, 4, 5}
	fmt.Println("declaration:", b)

	// Let the compiler calculate count
	b = [...]int{5, 4, 3, 2, 1}
	fmt.Println("declaration:", b)

	// Sparse
	b = [...]int{100, 3: 400, 500}
	fmt.Println("sparse:", b)

	// 2D
	var twoD [2][3]int
	fmt.Println("2D:", twoD)

	// 2D too. Can only elide the outer length
	twoD = [...][3]int{
		{1, 2, 3},
		{1, 2, 3},
	}
	fmt.Println("2D:", twoD)
}