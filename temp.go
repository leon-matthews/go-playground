package main

import (
    "image/color"
    "fmt"
)

var (
    red = color.RGBA{R: 0xff}
    blue = color.RGBA{B: 0xff}
    green = color.RGBA{G: 0xff}
    colors = []color.RGBA{red, blue, green}
)

func main() {
    for i := range colors {
        fmt.Printf("[%T]%+[1]v\n", colors[i:i+1])
    }
}
