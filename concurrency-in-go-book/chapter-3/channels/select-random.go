// Does select always choose the first match - or is it random?
package main

import (
    "fmt"
)

func main() {
    c1 := make(chan any); close(c1)
    c2 := make(chan any); close(c2)

    var count1, count2 int

    // Both channels are closed, therefore are open for reading zero values
    for range 1_000_000 {
        select {
        case <-c1:
            count1++
        case <-c2:
            count2++
        }
    }

    fmt.Println("Count 1:", count1)
    fmt.Println("Count 2:", count2)
}
