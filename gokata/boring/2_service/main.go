// You can use channel as a handle on a service. Here we make Ann and Joe talk
// in lockstep, i.e. one after another.
package main

import (
	"fmt"

	"github.com/gokatas/boring"
)

func main() {
	ann := boring.Person("Ann")
	joe := boring.Person("Joe")
	for range 10 {
		fmt.Println(<-ann)
		fmt.Println(<-joe)
	}
}
