// Package r8 emulates a simple CPU called the R8.
package r8

// Opcodes
const (
	Dec  = 49 // Decrement value in A
	Inc  = 48 // Increment value in A
	Noop = 0  // No-operation
)

const MaxRAM = 65536 // 64 KiB

type CPU struct {
	PC  uint16
	Mem [MaxRAM]byte

	// Registers
	A byte
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
	case Dec:
		cpu.A--
	case Inc:
		cpu.A++
	case Noop:
		// Do nothing
	default:
		panic("unknown opcode")
	}
}
