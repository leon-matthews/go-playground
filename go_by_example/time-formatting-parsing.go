package main

import (
	"fmt"
	"time"
)

func main() {
	now := time.Now()

	// Format uses example-based layout string, based on
	format := "2 Jan 2006"
	fmt.Println(now.Format(format))

	// Several layout strings are predefined
	fmt.Println(now.Format(time.RFC822))
	fmt.Println(now.Format(time.RFC3339))
	fmt.Println(now.Format(time.Kitchen))

	// Parse uses the same layout strings
	t1, _ := time.Parse(time.RFC3339, "2012-11-01T22:08:41+00:00")
	fmt.Println(t1)

	// Parse error
	_, err := time.Parse(time.RFC822, "8:41PM")
	fmt.Println(err) // parsing time "8:41PM" as "02 Jan 06 15:04 MST": cannot parse "8:41PM" as "02"
}
