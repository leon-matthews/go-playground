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
	for cpu.Step() {
	}
}

// Execute a single instruction, returning false if halted
// Halts and catches fire (panics) if unknown opcode found
func (cpu *CPU) Step() bool {
	opcode := cpu.Mem[cpu.PC]
	cpu.PC++

	switch opcode {
	case Dec:
		cpu.A--
	case Halt:
		return false
	case Inc:
		cpu.A++

	default:
		panic("unknown opcode")
	}
	return true
}
