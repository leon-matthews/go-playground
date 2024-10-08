package main

import "fmt"

const englishGreeting = "Hello, "

func Hello(name string) string {
	if name == "" {
		name = "world"
	}
	return englishGreeting + name
}

func main() {
	fmt.Println(Hello("world"))
}
