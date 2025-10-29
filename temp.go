package main

import (
    "fmt"
)

func main() {
    fmt.Println("twice:")
}

type MyInteger int

func (i MyInteger) Twice() MyInteger {
    return i * 2
}
