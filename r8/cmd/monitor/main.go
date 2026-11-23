package main

import (
	"fmt"

	"local.dev/r8"
)

func main() {
	fmt.Println("R8 Monitor (press enter to step CPU)")
	fmt.Println("  PC")
	cpu := r8.NewCPU()
	for {
		fmt.Printf("%04d > ", cpu.PC)

		// Wait for user to press input
		fmt.Scanln()
		cpu.Step()
	}
}
