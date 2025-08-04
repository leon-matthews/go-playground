// Basic time functionality. See also epoch.go
package main

import (
	"fmt"
	"time"
)

func main() {
	now := time.Now()
	fmt.Println(now)

	// Field access
	fmt.Println(now.Year())
	fmt.Println(now.Month())
	fmt.Println(now.Weekday())
	fmt.Println(now.Day())
	fmt.Println(now.Hour())
	fmt.Println(now.Minute())
	fmt.Println(now.Second())
	fmt.Println(now.Nanosecond())
	fmt.Println(now.Location())

	// Boolean comparison
	then := time.Date(2025, 3, 17, 12, 0, 0, 0, time.UTC)
	fmt.Println(then.Before(now)) // true
	fmt.Println(then.After(now))  // false
	fmt.Println(then.Equal(now))  // false

	// Sub() returns a [Duration]
	diff := now.Sub(then)
	fmt.Println(diff)
	fmt.Println(diff.Hours())
	fmt.Println(diff.Minutes())
	fmt.Println(diff.Seconds())
	fmt.Println(diff.Nanoseconds())

	// Date calculations
	fmt.Println(then.Add(diff))
	fmt.Println(now.Add(-diff))
}
