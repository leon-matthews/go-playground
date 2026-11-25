package r8_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"local.dev/r8"
)

func TestNewInitialisesCPU(t *testing.T) {
	t.Parallel()
	cpu := r8.NewCPU()
	assert.Equal(t, uint16(0), cpu.PC)
	assert.Equal(t, byte(0), cpu.Mem[0])
	assert.Equal(t, byte(0), cpu.A)
}

func TestStepIncrementsPC(t *testing.T) {
	t.Parallel()
	t.Run("step increments PC", func(t *testing.T) {
		t.Parallel()
		cpu := r8.NewCPU()
		assert.Equal(t, uint16(0), cpu.PC)

		cpu.Step()
		assert.Equal(t, uint16(1), cpu.PC)

		cpu.Step()
		assert.Equal(t, uint16(2), cpu.PC)
	})

	t.Run("PC wraps around to zero", func(t *testing.T) {
		t.Parallel()
		cpu := r8.NewCPU()
		cpu.PC = uint16(r8.MaxRAM - 1)
		assert.Equal(t, uint16(0xffff), cpu.PC)
		cpu.Step()
		assert.Equal(t, uint16(0x0000), cpu.PC)
	})
}

func TestDecrement(t *testing.T) {
	t.Parallel()
	t.Run("decrements A", func(t *testing.T) {
		t.Parallel()
		cpu := r8.NewCPU()
		cpu.A = 17
		cpu.Mem[0] = r8.DecA
		cpu.Step()
		assert.Equal(t, byte(16), cpu.A)
	})

	t.Run("wraps around to 255", func(t *testing.T) {
		t.Parallel()
		cpu := r8.NewCPU()
		cpu.A = 1
		cpu.Mem[0] = r8.DecA
		cpu.Mem[1] = r8.DecA

		assert.Equal(t, byte(1), cpu.A)
		cpu.Step()
		assert.Equal(t, byte(0), cpu.A)
		cpu.Step()
		assert.Equal(t, byte(255), cpu.A)
	})
}

func TestHalt(t *testing.T) {
	t.Parallel()
	t.Run("halts at second opcode", func(t *testing.T) {
		cpu := r8.NewCPU()
		cpu.Mem[0] = r8.IncA
		cpu.Mem[1] = r8.Halt

		assert.Equal(t, uint16(0), cpu.PC)
		cpu.Run()
		assert.Equal(t, uint16(2), cpu.PC)
	})
}

func TestLoad(t *testing.T) {
	t.Parallel()
	t.Run("load value into A", func(t *testing.T) {
		t.Parallel()
		cpu := r8.NewCPU()
		cpu.Mem[0] = r8.LdA
		cpu.Mem[1] = byte(42)
		cpu.Mem[2] = r8.Halt

		assert.Equal(t, byte(0), cpu.A)
		assert.Equal(t, uint16(0), cpu.PC)
		cpu.Run()
		assert.Equal(t, byte(42), cpu.A)
		assert.Equal(t, uint16(3), cpu.PC)
	})
}

func TestIncrement(t *testing.T) {
	t.Parallel()
	t.Run("increments A", func(t *testing.T) {
		t.Parallel()
		cpu := r8.NewCPU()
		assert.Equal(t, byte(0), cpu.A)
		cpu.Mem[0] = r8.IncA
		cpu.Step()
		assert.Equal(t, byte(1), cpu.A)
	})

	t.Run("wraps around to zero", func(t *testing.T) {
		t.Parallel()
		cpu := r8.NewCPU()
		cpu.A = 255
		cpu.Mem[0] = r8.IncA
		cpu.Step()
		assert.Equal(t, byte(0), cpu.A)
	})
}

func TestMemory(t *testing.T) {
	t.Parallel()
	t.Run("location wraps at 255", func(t *testing.T) {
		t.Parallel()
		cpu := r8.NewCPU()
		const addr = 42
		cpu.Mem[addr] = 254

		cpu.Mem[addr]++
		assert.Equal(t, byte(255), cpu.Mem[addr])
		cpu.Mem[addr]++
		assert.Equal(t, byte(0), cpu.Mem[addr])
	})
}
