package main

import "fmt"

type base struct {
	num int
}

type container struct {
	base
	str string
}

func (b base) describe() string {
	return fmt.Sprintf("base with num=%v", b.num)
}

type describer interface {
	describe() string
}

func main() {
	// Base struct instance
	b := base{3}
	fmt.Println(b.describe())

	// Container struct, positional syntax
	c := container{base{42}, "Hello"}

	// Long 'key: value' syntax
	c2 := container{
		base: base{num: 42},
		str: "Hello",
	}

	// Access embedded fields directly...
	fmt.Printf("c={num: %v, str: %v}\n", c.num, c.str)

	// ...or spell out full path
	fmt.Printf("c={num: %v, str: %v}\n", c.base.num, c.str)

	// The same applies to methods
	fmt.Println(c2.describe())
	fmt.Println(c2.base.describe())

	// Which means that our container now implements `describe()` method
	var d describer = c2
	fmt.Println(d.describe())
}
