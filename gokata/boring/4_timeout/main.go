// We use select to let talk Ann or Joe (whosever is ready) until no one speaks
// for 600ms.
//
// Select statement is another way to handle multiple channels. It's like switch
// but each case is a communication.
//
//   - All channels are evaluated.
//   - Blocks until one communication can proceed.
//   - If multiple can proceed, chooses (pseudo-)randomly.
//   - A default case, if present, executes immediately if no channel is ready.
package main

import (
	"fmt"
	"time"

	"github.com/gokatas/boring"
)

func main() {
	c := fanIn(boring.Person("Ann"), boring.Person("Joe"))
	for {
		select {
		case s := <-c:
			fmt.Println(s)
		case <-time.After(time.Millisecond * 900):
			fmt.Println("timeout")
			return
		}
	}
}

func fanIn(c1, c2 <-chan string) <-chan string {
	c := make(chan string)
	go func() {
		for {
			select {
			case c <- <-c1:	// Read from c1, then write to c
			case c <- <-c2:
			}
		}
	}()
	return c
}
