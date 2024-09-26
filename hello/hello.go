
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

    // Format message
    message, err := greetings.Hello("Leon")
    //~ message, err := greetings.Hello("")

    // Error? Logand exit
    if err != nil {
        log.Fatal(err)
    }

    // Print formatted message
    fmt.Println(message)
}
