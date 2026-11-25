// Package r8 emulates a simple CPU called the R8.
package r8

import (
	"errors"
)

// Opcodes
const (
	Dec  = 49 // Decrement value in A
	Halt = 0  // Halt stepping
	Inc  = 48 // Increment value in A
)

const MaxRAM = 65536 // 64 KiB

var ErrHalted = errors.New("CPU halted")

type CPU struct {
	PC  uint16
	Mem [MaxRAM]byte

	// Registers
	A byte
}

func NewCPU() *CPU {
	return &CPU{}
}

// Run steps through opcodes until a Halt is encountered
func (cpu *CPU) Run() {
	for {
		err := cpu.Step()
		if errors.Is(err, ErrHalted) {
			return
		}
	}
}

// Step through instructions
// Halts and catches fire (panics) if unknown opcode found
func (cpu *CPU) Step() error {
	opcode := cpu.Mem[cpu.PC]
	cpu.PC++

	switch opcode {
	case Dec:
		cpu.A--
	case Halt:
		return ErrHalted
	case Inc:
		cpu.A++

	default:
		panic("unknown opcode")
	}
	return nil
}
