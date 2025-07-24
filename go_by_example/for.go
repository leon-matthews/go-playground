package main

import "fmt"

func main() {
    // One condition
    i := 1
    for i <= 3 {
        fmt.Print(i)
        i = i + 1
    }
    fmt.Println()

    // Initial/condition/after
    for j := 0; j < 3; j++ {
        fmt.Print(j)
    }
    fmt.Println()

    // Range over integer
    for i := range 3 {
        fmt.Print(i)
    }
    fmt.Println()

    // Infinite loop - broken
    for {
        fmt.Println("loop")
        break
    }

    // Continue
    for i := range 6 {
        if i % 2 == 0 {
            continue
        }
        fmt.Print(i)
    }
}
