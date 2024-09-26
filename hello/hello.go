
package main


import (
    "fmt"

    "example.com/greetings"
)


func main() {
    message := greetings.Hello("Leon")
    fmt.Println(message)
}
