
package main


import (
    "fmt"
    "log"

    "example.com/greetings"
)


func main() {
    // Setup logger
    log.SetPrefix("greetings: ")
    log.SetFlags(0)

    // Slice of names
    names := []string {
        "Leon",
        "Alyson",
        "Blake",
        "Stella",
    }
    messages, err := greetings.Hellos(names)

    // Error? Log and exit
    if err != nil {
        log.Fatal(err)
    }

    // Print formatted message
    fmt.Println(messages)
}
