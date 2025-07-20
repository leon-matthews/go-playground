#!/usr/bin/env gorun

package main

import (
    "fmt"
    "slices"
)

func main() {
    s := []string{"zero", "one", "two", "three", "four"}
    for c := range slices.Chunk(s, 2) {
        fmt.Print(c)
    }
}
