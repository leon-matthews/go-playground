package main

import "fmt"

const (
	englishHelloPrefix  = "Hello, "
	englishHelloDefault = "world!"
)

func Hello(name string) string {
	if name == "" {
		name = "world!"
	}
	return englishHelloPrefix + name
}

func main() {
	fmt.Println(Hello(""))
}
