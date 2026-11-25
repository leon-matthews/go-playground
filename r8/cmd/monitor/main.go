package main

import (
	"fmt"

	"local.dev/r8"
)

func main() {
	cpu := r8.NewCPU()
	cpu.Mem[0] = r8.Inc
	cpu.Mem[1] = r8.Inc
	cpu.Mem[2] = r8.Inc

	fmt.Println("R8 Monitor (press enter to step CPU)")
	fmt.Println("  PC  A")
	for {
		fmt.Printf("%04d %02d > ", cpu.PC, cpu.A)

		// Wait for user to press input
		fmt.Scanln()
		cpu.Step()
	}
}
