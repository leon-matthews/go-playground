// Package r8 emulates a simple CPU called the R8.
package r8

import (
	"encoding/binary"
	"errors"
)

// Opcodes
const (
	AddA = 90  // Add next operand into A
	Bra  = 240 // Branch always skips ahead by its argument's value
	DecA = 64  // Decrement value in A
	Halt = 0   // Halt stepping
	IncA = 48  // Increment value in A
	Jmp  = 247 // Set PC to the 16-bit value from next two operands
	LdA  = 16  // Load next operand into A
)

const MaxRAM = 65536 // 64 KiB

var ErrHalted = errors.New("CPU halted")

// CPU holds the state of the R8 processor and its memory
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

// Fetch16 consumes two opcodes into a 16-bit value, in little endian mode.
func (cpu *CPU) Fetch16() uint16 {
	low, high := cpu.Fetch(), cpu.Fetch()
	addr := binary.LittleEndian.Uint16([]byte{low, high})
	return addr
}

// Run steps through opcodes in memory until a Halt is encountered
func (cpu *CPU) Run() {
	for cpu.Step() {
	}
}

// Load given program into memory and run until halted
func (cpu *CPU) RunProgram(program []byte) {
	cpu.PC = 0
	copy(cpu.Mem[:], program)
	cpu.Run()
}

// Execute a single instruction, returning false if halted
// Halts and catches fire (panics) if unknown opcode found
func (cpu *CPU) Step() bool {
	opcode := cpu.Fetch()
	switch opcode {
	case AddA:
		cpu.A += cpu.Fetch()
	case Bra:
		cpu.PC += uint16(cpu.Fetch())
	case DecA:
		cpu.A--
	case Halt:
		return false
	case IncA:
		cpu.A++
	case Jmp:
		cpu.PC = cpu.Fetch16()
	case LdA:
		cpu.A = cpu.Fetch()
	default:
		panic("unknown opcode")
	}
	return true
}
