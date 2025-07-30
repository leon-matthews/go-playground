package main

import (
    "fmt"
    "time"
)

func main() {
    c1 := make(chan any)
    close(c1)
    c2 := make(chan any)
    close(c2)

    var count1, count2 int
    timeout := time.NewTimer(1 * time.Second)
    defer timeout.Stop() // Not needed since Go 1.23

loop:
    for {
        select {
        case <-c1:
            count1++
        case <-c2:
            count2++
        case <-timeout.C:
            fmt.Println("Timed out")
            break loop
        }
    }
    fmt.Println("c1:", count1, "c2:", count2)
}
