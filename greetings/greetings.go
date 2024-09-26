
package greetings


import (
    "errors"
    "fmt"
)


// Hello returns a greeting for the named person.
func Hello(name string) (string, error) {
    // Report error back to caller
    if name == "" {
        return "", errors.New("empty name")
    }

    // Return a greeting that embeds the name in a message.
    message := fmt.Sprintf("Hi, %v. Welcome!", name)
    return message, nil
}
