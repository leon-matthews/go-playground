package main

import (
    "fmt"

    "github.com/alecthomas/kong"
)

var CLI struct {
}

func main() {
    ctx := kong.Parse(&CLI)
    fmt.Println(ctx)
}
