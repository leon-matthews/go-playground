package main

import (
	"fmt"
	"strconv"
)

func main() {
	// Parse into 64-bit float
	f, _ := strconv.ParseFloat("1.234", 64)
	fmt.Println(f)

	// Parse integer, auto-detecting base, 64-bit
	i, _ := strconv.ParseInt("123", 0, 64)
	fmt.Println(i)

	// Atoi is equivalent to ParseInt(s, 10, 0)
	i2, _ := strconv.Atoi("4023")
	fmt.Println(i2)

	_, err := strconv.Atoi("wot?")
	fmt.Println(err) // strconv.Atoi: parsing "wot?": invalid syntax

	// String quoting
	fmt.Println(strconv.QuoteToASCII("Hello, 世界")) // "Hello, \u4e16\u754c"
}
