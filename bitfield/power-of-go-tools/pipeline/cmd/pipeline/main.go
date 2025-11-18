package main

import "pipeline"

func main() {
	pipeline.FromString("Hello, world\n").Stdout()
}
