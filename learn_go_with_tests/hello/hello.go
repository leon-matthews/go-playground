package main

import "fmt"

const (
	french          = "French"
	spanish         = "Spanish"
	englishGreeting = "Hello, "
	frenchGreeting  = "Bonjour, "
	spanishGreeting = "Hola, "
)

func Hello(name string, language string) string {
	if name == "" {
		name = "world"
	}

	greeting := englishGreeting
	switch language {
	case french:
		greeting = frenchGreeting
	case spanish:
		greeting = spanishGreeting
	}

	return greeting + name
}

func main() {
	fmt.Println(Hello("world", ""))
}
