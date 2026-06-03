// Quit makes a person stop talking by sending on a channel. They will
// also let us know when they're done quitting via another channel.
package main

import (
	"fmt"
	"math/rand/v2"
	"time"

	"github.com/gokatas/boring"
)

func main() {
	quit := make(chan bool)
	c := quitter("Jack", quit)
	for range rand.N(10) {
		fmt.Println(<-c)
	}

	// Send a quit signal
	quit <- true
	fmt.Println("quitting...")

	// Wait for a single response on the same channel
	<-quit
}

// quitter is like Person, but exits if the caller (or anybody) writes to quit
// A value is written back to quit after cleanup has finished.
func quitter(name string, quit chan bool) <-chan string {
	c := make(chan string)
	go func() {
		for i := 0; ; i++ {
			select {
			case c <- fmt.Sprintf(boring.Format, name, i):
			case <-quit:
				cleanup()
				quit <- true
				return
			}
			time.Sleep(rand.N(1e3 * time.Millisecond))
		}
	}()
	return c
}

func cleanup() {
	time.Sleep(rand.N(3e3 * time.Millisecond))
}
