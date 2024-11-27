package main

import (
	"clockface"
	"os"
	"time"
)

func main() {
	now := time.Now()
	clockface.SVGWriter(os.Stdout, now)
}
