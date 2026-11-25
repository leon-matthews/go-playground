// Package r8 emulates a simple CPU called the R8.
package r8

import (
	"errors"
)

// Opcodes
const (
	AddA = 90 // Add next operand into A
	DecA = 64 // Decrement value in A
	Halt = 0  // Halt stepping
	IncA = 48 // Increment value in A
	LdA  = 16 // Load next operand into A
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

// Fetch returns the opcode at PC and increments it
func (cpu *CPU) Fetch() byte {
	opcode := cpu.Mem[cpu.PC]
	cpu.PC++
	return opcode
}

// Run steps through opcodes until a Halt is encountered
func (cpu *CPU) Run() {
	for cpu.Step() {
	}
}

// Execute a single instruction, returning false if halted
// Halts and catches fire (panics) if unknown opcode found
func (cpu *CPU) Step() bool {
	opcode := cpu.Fetch()
	switch opcode {
	case AddA:
		cpu.A += cpu.Fetch()
	case DecA:
		cpu.A--
	case Halt:
		return false
	case IncA:
		cpu.A++
	case LdA:
		cpu.A = cpu.Fetch()
	default:
		panic("unknown opcode")
	}
	return true
}
