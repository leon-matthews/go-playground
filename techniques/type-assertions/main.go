package main

import (
	"fmt"
	"reflect"
)

func main() {
	typeAssertions()
	typeSwitch()
	anonymousInterface()
}

func typeAssertions() {
	// Dynamic typing in a statically typed language?!
	var a any
	fmt.Printf("[%T] %#[1]v\n", a)

	a = 27
	fmt.Printf("[%T] %#[1]v\n", a)

	a = "Hello, world!"
	fmt.Printf("[%T] %#[1]v\n", a)

	// Type assertion to concrete string, but can panic!
	b := a.(string)
	fmt.Printf("[%T] %#[1]v\n", b)

	// Type assertion to int, comma-ok won't panic
	if c, ok := a.(int); ok {
		fmt.Printf("[%T] %#[1]v\n", c)
	} else {
		fmt.Println("a is not an int")
	}

	// It's difficult to differentiate a & b as Go hides the interface wrapper
	// [reflect.TypeOf] has an argument type of any, so b gets boxed too
	fmt.Printf("reflect.TypeOf(x):  a:%v b:%v\n", reflect.TypeOf(a), reflect.TypeOf(b))
	// Unless we pass them as pointers
	fmt.Printf("reflect.TypeOf(&x): a:%v b:%v\n", reflect.TypeOf(&a), reflect.TypeOf(&b))
}

// A Pooer can Poo
type Pooer interface {
	Poo()
}

type Baby struct{}

// A baby can Pee
func (b Baby) Pee() {}

// A baby can Poo
func (b Baby) Poo() {}

// anonymousInterface demonstrates looking for a specific function
func anonymousInterface() {
	var p Pooer
	p = Baby{}
	p.Poo()

	// Can't call p.Pee() as p is a Pooer interface
	if v, ok := p.(interface{ Pee() }); ok {
		v.Pee()
	}
}

// typeSwitch demonstrates handling various concrete types
func typeSwitch() {
	processType(4)
	processType("banana")
	processType(12.5)
}

func processType(v any) {
	switch v := v.(type) {
	case int:
		fmt.Println("int", v)
	case string:
		fmt.Println("string", v)
	default:
		fmt.Println("default", v)
	}
}
