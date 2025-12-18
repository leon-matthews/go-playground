package main

// #include <stdlib.h>
import "C" // This import statement MUST be stand-alone

import (
    "fmt"
    "time"
)

func Random() int {
    return int(C.random())
}

func Seed(i int) {
    C.srandom(C.uint(i))
}

func main() {
    now := time.Now()
    Seed(int(now.Unix()))
    fmt.Println("Today's lucky number is", Random())
}
