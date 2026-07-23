package fake_test

import (
	"fmt"
	"time"

	"local.dev/fake-go"
)

// Example gives a quick tour of the most common generators.
func Example() {
	f := fake.New(42)
	fmt.Println(f.FullName())
	fmt.Println(f.Email())
	fmt.Println(f.City())
	fmt.Printf("$%.2f\n", float64(f.Price(1, 100))/100)
	fmt.Println(f.Words(4, 4))
	// Output:
	// Gabriella Griffiths
	// 65922784@example.com
	// Auckland
	// $77.01
	// esse pariatur qui veniam
}

// ExampleNew shows that a seed makes the output reproducible.
func ExampleNew() {
	a, b := fake.New(42), fake.New(42)
	fmt.Println(a.FullName() == b.FullName())

	c := fake.New(7)
	fmt.Println(a.FullName() == c.FullName())
	// Output:
	// true
	// false
}

// ExampleFaker_Address assembles a postal address from its parts.
func ExampleFaker_Address() {
	f := fake.New(42)
	a := f.Address()
	fmt.Println(a.Address1)
	fmt.Println(a.City)
	fmt.Println(a.PostCode)
	fmt.Println()
	fmt.Println(f.AddressMultiline())
	// Output:
	// 63 Ellerslie-Panmure Way
	// Hamilton
	// 7840
	//
	// 52B Higgs Drive
	// Upper Hutt 4230
}

// ExampleFaker_Words samples the text generators.
func ExampleFaker_Words() {
	f := fake.New(42)
	fmt.Println(f.Words(4, 4))
	fmt.Println(f.Code(6, 6))
	fmt.Println(f.Digits(4, 4))
	fmt.Println(f.Letters(8, 8))
	// Output:
	// accusantium quam ullam alias
	// TV4006
	// 1396
	// MHIAFJJW
}

// ExampleFaker_Price samples the numeric generators.
func ExampleFaker_Price() {
	f := fake.New(42)
	fmt.Println(f.Int(1, 100))
	fmt.Printf("%.4f\n", f.Float(0, 1))
	fmt.Printf("$%.2f\n", float64(f.Price(1, 100))/100)
	fmt.Println(f.Bool())
	// Output:
	// 62
	// 0.3861
	// $86.35
	// false
}

// ExampleFaker_Between picks a random time from within a fixed range.
func ExampleFaker_Between() {
	f := fake.New(42)
	start := time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2020, time.December, 31, 0, 0, 0, 0, time.UTC)
	fmt.Println(f.Between(start, end).Format("2006-01-02"))
	// Output: 2020-04-21
}

// ExampleFaker_RelativeTime parses times relative to now.
//
// Values are "now", signed pairs such as "+3 days" or "-40 years", and compact
// forms such as "2y4w7d". Units are d, h, w, m (month), and y.
func ExampleFaker_RelativeTime() {
	f := fake.New(42)
	birth, _ := f.RelativeTime("-40 years")
	renewal, _ := f.RelativeTime("+2y6m")
	fmt.Println(birth.Before(renewal))
	// Output: true
}
