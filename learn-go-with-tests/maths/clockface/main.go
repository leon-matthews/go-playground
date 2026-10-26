package main

import (
	"os"
	"time"

	"local.dev/clockface"
)

func main() {
	now := time.Now()
	clockface.SVGWriter(os.Stdout, now)
}
