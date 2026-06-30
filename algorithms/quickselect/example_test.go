package quickselect_test

import (
	"cmp"
	"fmt"
	"slices"

	"local.dev/quickselect"
)

// ExampleNthElement finds the median of a slice without fully sorting it.
func ExampleNthElement() {
	scores := []int{50, 20, 90, 40, 10, 80, 30}
	median := quickselect.NthElement(scores, len(scores)/2)
	fmt.Println(median)
	// Output: 40
}

// ExampleNthElement_partition shows that selecting also moves smaller elements to the left.
func ExampleNthElement_partition() {
	values := []int{9, 1, 8, 2, 7, 3, 6, 4, 5}

	// Place the 4th-smallest (k=3); the three smaller values land to its left.
	fourth := quickselect.NthElement(values, 3)

	smaller := slices.Clone(values[:3])
	slices.Sort(smaller) // sort only so the printed order is stable
	fmt.Printf("4th smallest: %d\n", fourth)
	fmt.Printf("the three smaller: %v\n", smaller)
	// Output:
	// 4th smallest: 4
	// the three smaller: [1 2 3]
}

// ExampleNthElementFunc selects by a custom comparator, here the cheapest product by price.
func ExampleNthElementFunc() {
	type product struct {
		name  string
		price int
	}
	products := []product{
		{"hammer", 18},
		{"nails", 4},
		{"saw", 32},
		{"glue", 7},
		{"drill", 56},
	}

	// A comparator is needed because the < operator does not apply to structs.
	byPrice := func(a, b product) int { return cmp.Compare(a.price, b.price) }
	cheapest := quickselect.NthElementFunc(products, 0, byPrice)

	fmt.Printf("%s at $%d\n", cheapest.name, cheapest.price)
	// Output: nails at $4
}
