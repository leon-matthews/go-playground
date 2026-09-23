package main

import (
	"fmt"
	"time"
)

func main() {
	for {
		// Print the time
		now := time.Now()
		fmt.Println(now.Format("15:04:05.000"))

		// Sleep until exact time. Refetch time to avoid time taken to print
		next := time.Now().Truncate(10 * time.Second).Add(10 * time.Second)
		time.Sleep(time.Until(next))
	}
}
