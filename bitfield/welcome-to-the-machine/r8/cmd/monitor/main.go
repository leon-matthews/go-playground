package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"local.dev/r8"
)

func main() {
	// Build program from args
	program := make([]byte, 0, len(os.Args)-1)
	for _, arg := range os.Args[1:] {
		opcode, err := strconv.ParseUint(arg, 10, 8)
		if err != nil {
			log.Fatal("reading program:", err)
		}
		program = append(program, byte(opcode))
	}

	// Create CPU & load program
	cpu := r8.NewCPU()
	copy(cpu.Mem[:], program)

	// Run program
	fmt.Println("R8 Monitor (press enter to step CPU)")
	fmt.Println("  PC  A")
	for {
		fmt.Printf("%04d %02d > ", cpu.PC, cpu.A)

		// Wait for user to press input
		fmt.Scanln()
		cpu.Step()
	}
}
