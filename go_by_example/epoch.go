// Functionality related to the Unix epoch.
package main

import (
	"fmt"
	"time"
)

func main() {
	now := time.Now()
	fmt.Println(now)

	// Methods on [Time] convert to Unix epoch
	fmt.Println(now.Unix())
	fmt.Println(now.UnixMilli())
	fmt.Println(now.UnixNano())

	// Package functions convert Unix epoch to a [Time]
	fmt.Println(time.Unix(0, 0)) // Nanoseconds in 2nd arg
}
