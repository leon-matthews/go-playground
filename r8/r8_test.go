package r8_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"local.dev/r8"
)

func TestNewInitialisesCPU(t *testing.T) {
	t.Parallel()
	cpu := r8.NewCPU()
	assert.Equal(t, 0, cpu.PC)
	assert.Equal(t, 0, cpu.Mem[0])
	assert.Equal(t, 0, cpu.A)
}

func TestStepIncrementsPC(t *testing.T) {
	t.Parallel()

	cpu := r8.NewCPU()
	assert.Equal(t, 0, cpu.PC)

	cpu.Step()
	assert.Equal(t, 1, cpu.PC)

	cpu.Step()
	assert.Equal(t, 2, cpu.PC)
}

func TestIncrement(t *testing.T) {
	t.Parallel()

	cpu := r8.NewCPU()
	assert.Equal(t, 0, cpu.A)

	cpu.Mem[0] = r8.Inc
	cpu.Step()
	assert.Equal(t, 1, cpu.A)
}
