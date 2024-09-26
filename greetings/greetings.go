
package greetings


import (
    "errors"
    "fmt"
    "math/rand"
)


// Hello returns a greeting for the named person.
func Hello(name string) (string, error) {
    // Report error back to caller
    if name == "" {
        return "", errors.New("empty name")
    }

    // Return a greeting that embeds the name in a message.
    message := fmt.Sprintf(randomFormat(), name)
    return message, nil
}


func randomFormat() string {
    formats := []string {
        "Hi, %v. Welcome!",
        "Great to see you, %v!",
        "Hail, %v! Well met!",
    }

    return formats[rand.Intn(len(formats))]
}
