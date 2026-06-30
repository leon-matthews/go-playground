package quickselect_test

import (
	"cmp"
	"fmt"

	"local.dev/quickselect"
)

// ExampleNthElement finds the median of a slice without fully sorting it.
func ExampleNthElement() {
	scores := []int{50, 40, 20, 90, 10, 80, 30}
	median := quickselect.NthElement(scores, len(scores)/2)
	fmt.Println(median)
	// Output: 40
}

// ExampleNthElementFunc selects by a custom comparator, here the 2nd cheapest bottle of wine.
func ExampleNthElementFunc() {
	type product struct {
		name  string
		price int
	}
	products := []product{
		{"Golden Sunset Blend", 18},
		{"Serenity Reserve", 14},
		{"Ethereal Elegance Rose", 32},
		{"Mystic Twilight Cabernet", 7},
		{"Enchanted Forest Merlot", 56},
	}

	// A comparator is needed because the < operator does not apply to structs.
	byPrice := func(a, b product) int { return cmp.Compare(a.price, b.price) }

	// Never pick the cheapest bottle of wine on the menu! 1st is k=0, 2nd is k=1
	secondCheapest := quickselect.NthElementFunc(products, 1, byPrice)

	fmt.Printf("%s at $%d\n", secondCheapest.name, secondCheapest.price)
	// Output: Serenity Reserve at $14
}
