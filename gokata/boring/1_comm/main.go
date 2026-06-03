// The main goroutine communicates with the goroutine launched by Person via a
// channel. A channel allows for communication and synchronization between
// goroutines.
package main

import (
	"fmt"

	"github.com/gokatas/boring"
)

func main() {
	c := boring.Person("Joe")
	for range 10 {
		fmt.Println(<-c)
	}
}
