package main

import "fmt"

const (
	englishHelloPrefix = "Hello"
	frenchHelloPrefix  = "Bonjour"
	spanishHelloPrefix = "Hola"
)

func main() {
	fmt.Println(Hello("", ""))
}

func Hello(name, language string) string {
	// Zero value defaults
	if name == "" {
		name = "world!"
	}
	if language == "" {
		language = "en"
	}
	prefix := greetingPrefix(language)
	return fmt.Sprintf("%s, %s", prefix, name)
}

func greetingPrefix(language string) string {
	var prefix string
	switch language {
	case "es":
		prefix = spanishHelloPrefix
	case "fr":
		prefix = frenchHelloPrefix
	default:
		prefix = englishHelloPrefix
	}
	return prefix
}
