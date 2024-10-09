package main

import (
	"fmt"
	//~ "time"
)

func main() {
	number := Stringify(2)
	fmt.Println(number)
}

func Stringify(number int) (value string) {
	switch number {
		case 1:
			value = "one"
		case 2:
			value = "two"
		case 3:
			value = "three"
	}
	return
}
