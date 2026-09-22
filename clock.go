package main

import (
	"fmt"
	"time"
)

func main() {
	for {
		// Print the time
		now := time.Now()
		fmt.Print(now.Format("15:04:05.000") + "\r")

		// Sleep until exact time
		next := now.Truncate(10 * time.Second).Add(10 * time.Second)
		time.Sleep(time.Until(next))
	}
}
