package r8_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"local.dev/r8"
)

func TestNewCPU(t *testing.T) {
	t.Parallel()
	cpu := r8.NewCPU()
	assert.Equal(t, uint16(0), cpu.PC)
	assert.Equal(t, byte(0), cpu.Mem[0])
	assert.Equal(t, byte(0), cpu.A)
	assert.Equal(t, false, cpu.Zero)
}

func TestAdd(t *testing.T) {
	t.Parallel()
	t.Run("add to A", func(t *testing.T) {
		t.Parallel()
		cpu := r8.NewCPU()
		cpu.A = 10
		program := []byte{
			r8.AddA,
			7,
			r8.Halt,
		}
		cpu.RunProgram(program)
		assert.Equal(t, byte(17), cpu.A)
	})
}

func TestBranchAlways(t *testing.T) {
	t.Parallel()
	t.Run("jumps over halt", func(t *testing.T) {
		t.Parallel()
		cpu := r8.NewCPU()
		program := []byte{
			r8.Bra,
			1,
			r8.Halt,
			r8.Halt,
		}
		cpu.RunProgram(program)
		assert.Equal(t, uint16(4), cpu.PC)
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
		assert.False(t, cpu.Zero)
	})

	t.Run("wraps around to 255", func(t *testing.T) {
		t.Parallel()
		cpu := r8.NewCPU()
		cpu.A = 1
		cpu.Mem[0] = r8.DecA
		cpu.Mem[1] = r8.DecA

		cpu.Step()
		assert.Equal(t, byte(0), cpu.A)
		assert.True(t, cpu.Zero)
		cpu.Step()

		assert.Equal(t, byte(255), cpu.A)
		assert.False(t, cpu.Zero)
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

func TestIncrement(t *testing.T) {
	t.Parallel()
	t.Run("increments A", func(t *testing.T) {
		t.Parallel()
		cpu := r8.NewCPU()
		assert.Equal(t, byte(0), cpu.A)
		cpu.Mem[0] = r8.IncA
		cpu.Step()
		assert.Equal(t, byte(1), cpu.A)
		assert.False(t, cpu.Zero)
	})

	t.Run("wraps around to zero", func(t *testing.T) {
		t.Parallel()
		cpu := r8.NewCPU()
		cpu.A = 255
		cpu.Mem[0] = r8.IncA
		cpu.Step()
		assert.Equal(t, byte(0), cpu.A)
		assert.True(t, cpu.Zero)
	})
}

func TestJump(t *testing.T) {
	t.Parallel()
	t.Run("jump past halt", func(t *testing.T) {
		t.Parallel()
		cpu := r8.NewCPU()
		program := []byte{
			r8.Jmp,  // Jmp 4
			4,       // low byte
			0,       // high byte
			r8.Halt, // Halt
			r8.Halt, // Halt
		}
		cpu.RunProgram(program)

		// Program jumped over the first halt!
		assert.Equal(t, uint16(5), cpu.PC)
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

		cpu.Run()
		assert.Equal(t, byte(42), cpu.A)
		assert.Equal(t, uint16(3), cpu.PC)
		assert.False(t, cpu.Zero)
	})

	t.Run("loading zero into A sets zero flag", func(t *testing.T) {
		t.Parallel()
		cpu := r8.NewCPU()

		cpu.Mem[0] = r8.LdA
		cpu.Mem[1] = byte(0)
		cpu.Mem[2] = r8.Halt

		cpu.Run()
		assert.Equal(t, byte(0), cpu.A)
		assert.True(t, cpu.Zero)
		assert.Equal(t, uint16(3), cpu.PC)
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

func TestStep(t *testing.T) {
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
