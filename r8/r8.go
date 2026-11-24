// Package r8 emulates a simple CPU called the R8.
package r8

// Opcodes
const (
	Inc  = 48 // Increment value in A
	Noop = 0  // No-operation
)

type CPU struct {
	PC  int
	Mem [65536]int

	// Registers
	A int
}

func NewCPU() *CPU {
	return &CPU{}
}

// Step
// Halts and catches fire (panics) if unknown opcode found
func (cpu *CPU) Step() {
	opcode := cpu.Mem[cpu.PC]
	cpu.PC++

	switch opcode {
	case Inc:
		cpu.A++
	case Noop:
		// Do nothing
	default:
		panic("unknown opcode")
	}
}
