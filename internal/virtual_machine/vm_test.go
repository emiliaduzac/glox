package virtualmachine

import (
	"testing"
	c "tp-lox/internal/chunk"
	"tp-lox/internal/utils"
)

var line = 1

func setUpVM(t *testing.T) *VirtualMachine {
	t.Helper()
	vm := NewVirtualMachine()
	return vm
}

func writeConstant(chunk *c.Chunk, value any) {
	index := chunk.AddConstant(value)
	chunk.Write(byte(c.OP_CONSTANT), line)
	chunk.Write(byte(index), line)
}

func TestOpReturn(t *testing.T) {
	t.Run("Stack vacío no da error", func(t *testing.T) {
		vm := setUpVM(t)
		chunk := c.NewChunk()
		chunk.Write(byte(c.OP_RETURN), line)

		err := vm.Run(*chunk)

		utils.AssertNoError(t, err)
	})

	t.Run("Stack con elementos da error", func(t *testing.T) {
		vm := setUpVM(t)
		chunk := c.NewChunk()
		chunk.Write(byte(c.OP_NIL), line)
		chunk.Write(byte(c.OP_RETURN), line)

		err := vm.Run(*chunk)

		utils.AssertErrorMessage(t, err, "Inconsistent Stack Height")
	})

	t.Run("Corta la ejecución", func(t *testing.T) {
		vm := setUpVM(t)
		chunk := c.NewChunk()
		chunk.Write(byte(c.OP_RETURN), line)
		chunk.Write(byte(c.OP_NIL), line)

		err := vm.Run(*chunk)

		utils.AssertNoError(t, err)
		utils.AssertValue(t, len(vm.stack), 0)
	})
}

func TestOpConstant(t *testing.T) {
	line := 1
	tests := []struct {
		name  string
		value any
	}{
		{"número", 99.0},
		{"string", "Hello"},
		{"string vacío", ""},
		{"booleano", true},
		{"nil", nil},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			vm := setUpVM(t)
			chunk := c.NewChunk()
			writeConstant(chunk, test.value)

			err := vm.Run(*chunk)

			utils.AssertNoError(t, err)
			utils.AssertValue(t, len(vm.stack), 1)
			utils.AssertValue(t, vm.stack[0], test.value)
		})
	}

	t.Run("Dos constantes mantienen el orden", func(t *testing.T) {
		vm := setUpVM(t)
		chunk := c.NewChunk()
		writeConstant(chunk, 1.0)
		writeConstant(chunk, 2.0)

		err := vm.Run(*chunk)

		utils.AssertNoError(t, err)
		utils.AssertValue(t, len(vm.stack), 2)
		utils.AssertValue(t, vm.stack[0], 1.0)
		utils.AssertValue(t, vm.stack[1], 2.0)
	})

	t.Run("Usa el operando como índice para obtener constantes del pool, no el orden del pool", func(t *testing.T) {
		vm := setUpVM(t)
		chunk := c.NewChunk()

		// Pool de constantes: [10.0, 20.0]
		chunk.AddConstant(10.0)
		chunk.AddConstant(20.0)

		// Podemos cargar en el stack los índices en orden inverso: primero el 1 y después el 0
		chunk.Write(byte(c.OP_CONSTANT), line)
		chunk.Write(1, line)
		chunk.Write(byte(c.OP_CONSTANT), line)
		chunk.Write(0, line)

		err := vm.Run(*chunk)

		utils.AssertNoError(t, err)
		utils.AssertValue(t, len(vm.stack), 2)
		// El stack sigue el orden de los operandos (1, 0), no el del pool (0, 1)
		utils.AssertValue(t, vm.stack[0], 20.0)
		utils.AssertValue(t, vm.stack[1], 10.0)
	})

	t.Run("Error si el índice no existe en el pool", func(t *testing.T) {
		vm := setUpVM(t)
		chunk := c.NewChunk()
		chunk.AddConstant(10.0)
		chunk.Write(byte(c.OP_CONSTANT), line)
		chunk.Write(5, 1)

		err := vm.Run(*chunk)

		utils.AssertError(t, err)
	})

	t.Run("Error si falta el operando", func(t *testing.T) {
		vm := setUpVM(t)
		chunk := c.NewChunk()
		chunk.AddConstant(10.0)
		chunk.Write(byte(c.OP_CONSTANT), line)

		err := vm.Run(*chunk)

		utils.AssertError(t, err)
	})
}
