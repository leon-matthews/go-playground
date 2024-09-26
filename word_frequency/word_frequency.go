
package main

import (
    "fmt"
    "log"
    "os"
)


func main() {
    log.SetFlags(0)
    if len(os.Args) != 2 {
        log.Fatal("usage: word_frequency PATH")
    }

    path := os.Args[1]
    fmt.Println("Count words in %v", path)
}
