package virtualmachine

import (
	"math"
	"testing"
	c "tp-lox/internal/chunk"
	"tp-lox/internal/utils"
)

var line = 1

func setUpVM(t *testing.T) (*VirtualMachine, *c.Chunk) {
	t.Helper()
	chunk := c.NewChunk()
	vm := NewVirtualMachine(chunk)
	return vm, chunk
}

func writeConstant(chunk *c.Chunk, value any) {
	index := chunk.AddConstant(value)
	chunk.Write(byte(c.OP_CONSTANT), line)
	chunk.Write(byte(index), line)
}

func TestOpReturn(t *testing.T) {
	t.Run("Stack vacío no da error", func(t *testing.T) {
		vm, chunk := setUpVM(t)
		chunk.Write(byte(c.OP_RETURN), line)

		err := vm.Run()

		utils.AssertNoError(t, err)
	})

	t.Run("Stack con elementos da error", func(t *testing.T) {
		vm, chunk := setUpVM(t)
		chunk.Write(byte(c.OP_NIL), line)
		chunk.Write(byte(c.OP_RETURN), line)

		err := vm.Run()

		utils.AssertErrorMessage(t, err, "inconsistent Stack Height")
	})

	t.Run("Corta la ejecución", func(t *testing.T) {
		vm, chunk := setUpVM(t)
		chunk.Write(byte(c.OP_RETURN), line)
		chunk.Write(byte(c.OP_NIL), line)

		err := vm.Run()

		utils.AssertNoError(t, err)
		utils.AssertValue(t, len(vm.stack), 0)
	})
}

func TestOpConstant(t *testing.T) {
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
			vm, chunk := setUpVM(t)
			writeConstant(chunk, test.value)

			err := vm.Run()

			utils.AssertNoError(t, err)
			utils.AssertValue(t, len(vm.stack), 1)
			utils.AssertValue(t, vm.stack[0], test.value)
		})
	}

	t.Run("Dos constantes mantienen el orden", func(t *testing.T) {
		vm, chunk := setUpVM(t)
		writeConstant(chunk, 1.0)
		writeConstant(chunk, 2.0)

		err := vm.Run()

		utils.AssertNoError(t, err)
		utils.AssertValue(t, len(vm.stack), 2)
		utils.AssertValue(t, vm.stack[0], 1.0)
		utils.AssertValue(t, vm.stack[1], 2.0)
	})

	t.Run("Usa el operando como índice para obtener constantes del pool, no el orden del pool", func(t *testing.T) {
		vm, chunk := setUpVM(t)

		// Pool de constantes: [10.0, 20.0]
		chunk.AddConstant(10.0)
		chunk.AddConstant(20.0)

		// Podemos cargar en el stack los índices en orden inverso: primero el 1 y después el 0
		chunk.Write(byte(c.OP_CONSTANT), line)
		chunk.Write(1, line)
		chunk.Write(byte(c.OP_CONSTANT), line)
		chunk.Write(0, line)

		err := vm.Run()

		utils.AssertNoError(t, err)
		utils.AssertValue(t, len(vm.stack), 2)
		// El stack sigue el orden de los operandos (1, 0), no el del pool (0, 1)
		utils.AssertValue(t, vm.stack[0], 20.0)
		utils.AssertValue(t, vm.stack[1], 10.0)
	})

	t.Run("Error si el índice no existe en el pool", func(t *testing.T) {
		vm, chunk := setUpVM(t)
		chunk.AddConstant(10.0)
		chunk.Write(byte(c.OP_CONSTANT), line)
		chunk.Write(5, 1)

		err := vm.Run()

		utils.AssertError(t, err)
	})

	t.Run("Error si falta el operando", func(t *testing.T) {
		vm, chunk := setUpVM(t)
		chunk.AddConstant(10.0)
		chunk.Write(byte(c.OP_CONSTANT), line)

		err := vm.Run()

		utils.AssertError(t, err)
	})
}

func TestOpLiterals(t *testing.T) {
	tests := []struct {
		name     string
		op       c.OpCode
		expected any
	}{
		{"nil", c.OP_NIL, nil},
		{"true", c.OP_TRUE, true},
		{"false", c.OP_FALSE, false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			vm, chunk := setUpVM(t)
			chunk.Write(byte(test.op), line)

			err := vm.Run()

			utils.AssertNoError(t, err)
			utils.AssertValue(t, len(vm.stack), 1)
			utils.AssertValue(t, vm.stack[0], test.expected)
		})
	}
}

func TestOpNegate(t *testing.T) {
	tests := []struct {
		name     string
		operand  float64
		expected float64
	}{
		{"número positivo", 5.0, -5.0},
		{"número negativo", -5.0, 5.0},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			vm, chunk := setUpVM(t)
			writeConstant(chunk, test.operand)
			chunk.Write(byte(c.OP_NEGATE), line)

			err := vm.Run()

			utils.AssertNoError(t, err)
			utils.AssertValue(t, len(vm.stack), 1)
			utils.AssertValue(t, vm.stack[0], test.expected)
		})
	}

	t.Run("Negar dos veces devuelve el valor original", func(t *testing.T) {
		vm, chunk := setUpVM(t)
		writeConstant(chunk, 5.0)
		chunk.Write(byte(c.OP_NEGATE), line)
		chunk.Write(byte(c.OP_NEGATE), line)

		err := vm.Run()

		utils.AssertNoError(t, err)
		utils.AssertValue(t, len(vm.stack), 1)
		utils.AssertValue(t, vm.stack[0], 5.0)
	})

	errorTests := []struct {
		name    string
		operand any
	}{
		{"string", "not a number"},
		{"nil", nil},
		{"booleano", true},
	}

	for _, test := range errorTests {
		t.Run("Error si el operando es "+test.name, func(t *testing.T) {
			vm, chunk := setUpVM(t)
			writeConstant(chunk, test.operand)
			chunk.Write(byte(c.OP_NEGATE), line)

			err := vm.Run()

			utils.AssertErrorMessage(t, err, "operand must be a number")
		})
	}

	t.Run("El error indica la línea de la instrucción", func(t *testing.T) {
		vm, chunk := setUpVM(t)
		index := chunk.AddConstant("a")
		chunk.Write(byte(c.OP_CONSTANT), 7)
		chunk.Write(byte(index), 7)
		chunk.Write(byte(c.OP_NEGATE), 7)

		err := vm.Run()

		utils.AssertErrorMessage(t, err, "[line 7]")
	})

	t.Run("Error si el stack está vacío", func(t *testing.T) {
		vm, chunk := setUpVM(t)
		chunk.Write(byte(c.OP_NEGATE), line)

		err := vm.Run()

		utils.AssertErrorMessage(t, err, "stack underflow")
	})
}

func TestOpAdd(t *testing.T) {
	tests := []struct {
		name     string
		left     float64
		right    float64
		expected float64
	}{
		{"dos positivos", 1.0, 2.0, 3.0},
		{"positivo y negativo", 5.0, -3.0, 2.0},
		{"dos negativos", -1.0, -2.0, -3.0},
		{"decimales", 0.5, 0.25, 0.75},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			vm, chunk := setUpVM(t)
			writeConstant(chunk, test.left)
			writeConstant(chunk, test.right)
			chunk.Write(byte(c.OP_ADD), line)

			err := vm.Run()

			utils.AssertNoError(t, err)
			utils.AssertValue(t, len(vm.stack), 1)
			utils.AssertValue(t, vm.stack[0], test.expected)
		})
	}

	t.Run("Concatena dos strings", func(t *testing.T) {
		vm, chunk := setUpVM(t)
		writeConstant(chunk, "Hello")
		writeConstant(chunk, " World")
		chunk.Write(byte(c.OP_ADD), line)

		err := vm.Run()

		utils.AssertNoError(t, err)
		utils.AssertValue(t, len(vm.stack), 1)
		utils.AssertValue(t, vm.stack[0], "Hello World")
	})

	incompatibleTests := []struct {
		name  string
		left  any
		right any
	}{
		{"número y string", 1.0, "a"},
		{"string y número", "a", 1.0},
		{"número y otro", 1.0, nil},
	}

	for _, test := range incompatibleTests {
		t.Run("Error si se suma "+test.name, func(t *testing.T) {
			vm, chunk := setUpVM(t)
			writeConstant(chunk, test.left)
			writeConstant(chunk, test.right)
			chunk.Write(byte(c.OP_ADD), line)

			err := vm.Run()

			utils.AssertErrorMessage(t, err, "operands must be two numbers or two strings")
		})
	}

	t.Run("Error si el stack tiene un solo elemento", func(t *testing.T) {
		vm, chunk := setUpVM(t)
		writeConstant(chunk, 1.0)
		chunk.Write(byte(c.OP_ADD), line)

		err := vm.Run()

		utils.AssertErrorMessage(t, err, "stack underflow")
	})

	t.Run("Error si el stack está vacío", func(t *testing.T) {
		vm, chunk := setUpVM(t)
		chunk.Write(byte(c.OP_ADD), line)

		err := vm.Run()

		utils.AssertErrorMessage(t, err, "stack underflow")
	})
}

func TestOpSubtract(t *testing.T) {
	tests := []struct {
		name     string
		left     float64
		right    float64
		expected float64
	}{
		{"mayor menos menor", 5.0, 3.0, 2.0},
		{"menor menos mayor", 3.0, 5.0, -2.0},
		{"restar un negativo", 5.0, -3.0, 8.0},
		{"decimales", 0.5, 0.25, 0.25},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			vm, chunk := setUpVM(t)
			writeConstant(chunk, test.left)
			writeConstant(chunk, test.right)
			chunk.Write(byte(c.OP_SUBTRACT), line)

			err := vm.Run()

			utils.AssertNoError(t, err)
			utils.AssertValue(t, len(vm.stack), 1)
			utils.AssertValue(t, vm.stack[0], test.expected)
		})
	}

	incompatibleTests := []struct {
		name  string
		left  any
		right any
	}{
		{"número y string", 1.0, "a"},
		{"número y otro", 1.0, nil},
	}

	for _, test := range incompatibleTests {
		t.Run("Error si se resta "+test.name, func(t *testing.T) {
			vm, chunk := setUpVM(t)
			writeConstant(chunk, test.left)
			writeConstant(chunk, test.right)
			chunk.Write(byte(c.OP_SUBTRACT), line)

			err := vm.Run()

			utils.AssertErrorMessage(t, err, "operands must be numbers")
		})
	}

	t.Run("Error si el stack tiene un solo elemento", func(t *testing.T) {
		vm, chunk := setUpVM(t)
		writeConstant(chunk, 1.0)
		chunk.Write(byte(c.OP_SUBTRACT), line)

		err := vm.Run()

		utils.AssertErrorMessage(t, err, "stack underflow")
	})

	t.Run("Error si el stack está vacío", func(t *testing.T) {
		vm, chunk := setUpVM(t)
		chunk.Write(byte(c.OP_SUBTRACT), line)

		err := vm.Run()

		utils.AssertErrorMessage(t, err, "stack underflow")
	})
}

func TestOpMultiply(t *testing.T) {
	tests := []struct {
		name     string
		left     float64
		right    float64
		expected float64
	}{
		{"dos positivos", 2.0, 3.0, 6.0},
		{"positivo y negativo", 2.0, -3.0, -6.0},
		{"dos negativos", -2.0, -3.0, 6.0},
		{"por cero", 7.0, 0.0, 0.0},
		{"decimales", 0.5, 0.5, 0.25},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			vm, chunk := setUpVM(t)
			writeConstant(chunk, test.left)
			writeConstant(chunk, test.right)
			chunk.Write(byte(c.OP_MULTIPLY), line)

			err := vm.Run()

			utils.AssertNoError(t, err)
			utils.AssertValue(t, len(vm.stack), 1)
			utils.AssertValue(t, vm.stack[0], test.expected)
		})
	}

	incompatibleTests := []struct {
		name  string
		left  any
		right any
	}{
		{"número y string", 1.0, "a"},
		{"número y otro", 1.0, nil},
	}

	for _, test := range incompatibleTests {
		t.Run("Error si se multiplica "+test.name, func(t *testing.T) {
			vm, chunk := setUpVM(t)
			writeConstant(chunk, test.left)
			writeConstant(chunk, test.right)
			chunk.Write(byte(c.OP_MULTIPLY), line)

			err := vm.Run()

			utils.AssertErrorMessage(t, err, "operands must be numbers")
		})
	}

	t.Run("Error si el stack tiene un solo elemento", func(t *testing.T) {
		vm, chunk := setUpVM(t)
		writeConstant(chunk, 1.0)
		chunk.Write(byte(c.OP_MULTIPLY), line)

		err := vm.Run()

		utils.AssertErrorMessage(t, err, "stack underflow")
	})

	t.Run("Error si el stack está vacío", func(t *testing.T) {
		vm, chunk := setUpVM(t)
		chunk.Write(byte(c.OP_MULTIPLY), line)

		err := vm.Run()

		utils.AssertErrorMessage(t, err, "stack underflow")
	})
}

func TestOpDivide(t *testing.T) {
	tests := []struct {
		name     string
		left     float64
		right    float64
		expected float64
	}{
		{"división exacta", 6.0, 3.0, 2.0},
		{"menor dividido mayor", 3.0, 6.0, 0.5},
		{"positivo y negativo", -6.0, 3.0, -2.0},
		{"decimales", 1.0, 4.0, 0.25},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			vm, chunk := setUpVM(t)
			writeConstant(chunk, test.left)
			writeConstant(chunk, test.right)
			chunk.Write(byte(c.OP_DIVIDE), line)

			err := vm.Run()

			utils.AssertNoError(t, err)
			utils.AssertValue(t, len(vm.stack), 1)
			utils.AssertValue(t, vm.stack[0], test.expected)
		})
	}

	incompatibleTests := []struct {
		name  string
		left  any
		right any
	}{
		{"número y string", 1.0, "a"},
		{"número y otro", 1.0, nil},
	}

	for _, test := range incompatibleTests {
		t.Run("Error si se divide "+test.name, func(t *testing.T) {
			vm, chunk := setUpVM(t)
			writeConstant(chunk, test.left)
			writeConstant(chunk, test.right)
			chunk.Write(byte(c.OP_DIVIDE), line)

			err := vm.Run()

			utils.AssertErrorMessage(t, err, "operands must be numbers")
		})
	}

	t.Run("Error si el stack tiene un solo elemento", func(t *testing.T) {
		vm, chunk := setUpVM(t)
		writeConstant(chunk, 1.0)
		chunk.Write(byte(c.OP_DIVIDE), line)

		err := vm.Run()

		utils.AssertErrorMessage(t, err, "stack underflow")
	})

	t.Run("Error si el stack está vacío", func(t *testing.T) {
		vm, chunk := setUpVM(t)
		chunk.Write(byte(c.OP_DIVIDE), line)

		err := vm.Run()

		utils.AssertErrorMessage(t, err, "stack underflow")
	})
}

func TestOpModulo(t *testing.T) {
	tests := []struct {
		name     string
		left     float64
		right    float64
		expected float64
	}{
		{"positivos", 10.0, 3.0, 1.0},
		{"módulo exacto", 10.0, 5.0, 0.0},
		{"cero módulo positivo", 0.0, 5.0, 0.0},
		{"decimales", 5.5, 2.0, 1.5},
		{"dividendo negativo", -10.0, 3.0, -1.0},
		{"divisor negativo", 10.0, -3.0, 1.0},
		{"ambos negativos", -10.0, -3.0, -1.0},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			vm, chunk := setUpVM(t)
			writeConstant(chunk, test.left)
			writeConstant(chunk, test.right)
			chunk.Write(byte(c.OP_MODULO), line)

			err := vm.Run()

			utils.AssertNoError(t, err)
			utils.AssertValue(t, len(vm.stack), 1)
			utils.AssertValue(t, vm.stack[0], test.expected)
		})
	}

	nanTests := []struct {
		name  string
		left  float64
		right float64
	}{
		{"módulo cero", 10.0, 0.0},
		{"cero módulo cero", 0.0, 0.0},
	}

	for _, test := range nanTests {
		t.Run("Da NaN en "+test.name, func(t *testing.T) {
			vm, chunk := setUpVM(t)
			writeConstant(chunk, test.left)
			writeConstant(chunk, test.right)
			chunk.Write(byte(c.OP_MODULO), line)

			err := vm.Run()

			utils.AssertNoError(t, err)
			utils.AssertValue(t, len(vm.stack), 1)
			result, ok := vm.stack[0].(float64)
			if !ok || !math.IsNaN(result) {
				t.Errorf("expected NaN, got %v", vm.stack[0])
			}
		})
	}

	incompatibleTests := []struct {
		name  string
		left  any
		right any
	}{
		{"número y string", 1.0, "a"},
		{"número y otro", 1.0, nil},
	}

	for _, test := range incompatibleTests {
		t.Run("Error si se hace módulo de "+test.name, func(t *testing.T) {
			vm, chunk := setUpVM(t)
			writeConstant(chunk, test.left)
			writeConstant(chunk, test.right)
			chunk.Write(byte(c.OP_MODULO), line)

			err := vm.Run()

			utils.AssertErrorMessage(t, err, "operands must be numbers")
		})
	}

	t.Run("Error si el stack tiene un solo elemento", func(t *testing.T) {
		vm, chunk := setUpVM(t)
		writeConstant(chunk, 1.0)
		chunk.Write(byte(c.OP_MODULO), line)

		err := vm.Run()

		utils.AssertErrorMessage(t, err, "stack underflow")
	})

	t.Run("Error si el stack está vacío", func(t *testing.T) {
		vm, chunk := setUpVM(t)
		chunk.Write(byte(c.OP_MODULO), line)

		err := vm.Run()

		utils.AssertErrorMessage(t, err, "stack underflow")
	})
}

func TestOpGreater(t *testing.T) {
	tests := []struct {
		name     string
		left     float64
		right    float64
		expected bool
	}{
		{"izquierdo mayor", 5.0, 3.0, true},
		{"izquierdo menor", 3.0, 5.0, false},
		{"iguales", 3.0, 3.0, false},
		{"negativos", -1.0, -2.0, true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			vm, chunk := setUpVM(t)
			writeConstant(chunk, test.left)
			writeConstant(chunk, test.right)
			chunk.Write(byte(c.OP_GREATER), line)

			err := vm.Run()

			utils.AssertNoError(t, err)
			utils.AssertValue(t, len(vm.stack), 1)
			utils.AssertValue(t, vm.stack[0], test.expected)
		})
	}

	incompatibleTests := []struct {
		name  string
		left  any
		right any
	}{
		{"número y string", 1.0, "a"},
		{"número y otro", 1.0, nil},
	}

	for _, test := range incompatibleTests {
		t.Run("Error si se compara "+test.name, func(t *testing.T) {
			vm, chunk := setUpVM(t)
			writeConstant(chunk, test.left)
			writeConstant(chunk, test.right)
			chunk.Write(byte(c.OP_GREATER), line)

			err := vm.Run()

			utils.AssertErrorMessage(t, err, "operands must be numbers")
		})
	}

	t.Run("Error si el stack tiene un solo elemento", func(t *testing.T) {
		vm, chunk := setUpVM(t)
		writeConstant(chunk, 1.0)
		chunk.Write(byte(c.OP_GREATER), line)

		err := vm.Run()

		utils.AssertErrorMessage(t, err, "stack underflow")
	})

	t.Run("Error si el stack está vacío", func(t *testing.T) {
		vm, chunk := setUpVM(t)
		chunk.Write(byte(c.OP_GREATER), line)

		err := vm.Run()

		utils.AssertErrorMessage(t, err, "stack underflow")
	})
}

func TestOpNot(t *testing.T) {
	tests := []struct {
		name     string
		value    any
		expected bool
	}{
		{"nil", nil, true},
		{"false", false, true},
		{"true", true, false},
		{"número", 1.0, false},
		{"cero", 0.0, false},
		{"string", "a", false},
		{"string vacío", "", false},
	}

	for _, test := range tests {
		t.Run("Not de "+test.name, func(t *testing.T) {
			vm, chunk := setUpVM(t)
			writeConstant(chunk, test.value)
			chunk.Write(byte(c.OP_NOT), line)

			err := vm.Run()

			utils.AssertNoError(t, err)
			utils.AssertValue(t, len(vm.stack), 1)
			utils.AssertValue(t, vm.stack[0], test.expected)
		})
	}

	t.Run("Error si el stack está vacío", func(t *testing.T) {
		vm, chunk := setUpVM(t)
		chunk.Write(byte(c.OP_NOT), line)

		err := vm.Run()

		utils.AssertErrorMessage(t, err, "stack underflow")
	})
}

func TestOpEqual(t *testing.T) {
	tests := []struct {
		name     string
		left     any
		right    any
		expected bool
	}{
		{"números iguales", 1.0, 1.0, true},
		{"números distintos", 1.0, 2.0, false},
		{"strings iguales", "a", "a", true},
		{"strings distintos", "a", "b", false},
		{"booleanos iguales", true, true, true},
		{"booleanos distintos", true, false, false},
		{"nil con nil", nil, nil, true},
		{"número y string", 1.0, "1", false},
		{"nil y false", nil, false, false},
		{"número y booleano", 1.0, true, false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			vm, chunk := setUpVM(t)
			writeConstant(chunk, test.left)
			writeConstant(chunk, test.right)
			chunk.Write(byte(c.OP_EQUAL), line)

			err := vm.Run()

			utils.AssertNoError(t, err)
			utils.AssertValue(t, len(vm.stack), 1)
			utils.AssertValue(t, vm.stack[0], test.expected)
		})
	}

	t.Run("Error si el stack tiene un solo elemento", func(t *testing.T) {
		vm, chunk := setUpVM(t)
		writeConstant(chunk, 1.0)
		chunk.Write(byte(c.OP_EQUAL), line)

		err := vm.Run()

		utils.AssertErrorMessage(t, err, "stack underflow")
	})

	t.Run("Error si el stack está vacío", func(t *testing.T) {
		vm, chunk := setUpVM(t)
		chunk.Write(byte(c.OP_EQUAL), line)

		err := vm.Run()

		utils.AssertErrorMessage(t, err, "stack underflow")
	})

	t.Run("Un string armado en runtime es igual a un literal con el mismo contenido", func(t *testing.T) {
		vm, chunk := setUpVM(t)
		writeConstant(chunk, "st")
		writeConstant(chunk, "ri")
		chunk.Write(byte(c.OP_ADD), line)
		writeConstant(chunk, "stri")
		chunk.Write(byte(c.OP_EQUAL), line)

		err := vm.Run()

		utils.AssertNoError(t, err)
		utils.AssertValue(t, len(vm.stack), 1)
		utils.AssertValue(t, vm.stack[0], true)
	})
}

func TestOpLess(t *testing.T) {
	tests := []struct {
		name     string
		left     float64
		right    float64
		expected bool
	}{
		{"izquierdo menor", 3.0, 5.0, true},
		{"izquierdo mayor", 5.0, 3.0, false},
		{"iguales", 3.0, 3.0, false},
		{"negativos", -2.0, -1.0, true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			vm, chunk := setUpVM(t)
			writeConstant(chunk, test.left)
			writeConstant(chunk, test.right)
			chunk.Write(byte(c.OP_LESS), line)

			err := vm.Run()

			utils.AssertNoError(t, err)
			utils.AssertValue(t, len(vm.stack), 1)
			utils.AssertValue(t, vm.stack[0], test.expected)
		})
	}

	incompatibleTests := []struct {
		name  string
		left  any
		right any
	}{
		{"número y string", 1.0, "a"},
		{"número y otro", 1.0, nil},
	}

	for _, test := range incompatibleTests {
		t.Run("Error si se compara "+test.name, func(t *testing.T) {
			vm, chunk := setUpVM(t)
			writeConstant(chunk, test.left)
			writeConstant(chunk, test.right)
			chunk.Write(byte(c.OP_LESS), line)

			err := vm.Run()

			utils.AssertErrorMessage(t, err, "operands must be numbers")
		})
	}

	t.Run("Error si el stack tiene un solo elemento", func(t *testing.T) {
		vm, chunk := setUpVM(t)
		writeConstant(chunk, 1.0)
		chunk.Write(byte(c.OP_LESS), line)

		err := vm.Run()

		utils.AssertErrorMessage(t, err, "stack underflow")
	})

	t.Run("Error si el stack está vacío", func(t *testing.T) {
		vm, chunk := setUpVM(t)
		chunk.Write(byte(c.OP_LESS), line)

		err := vm.Run()

		utils.AssertErrorMessage(t, err, "stack underflow")
	})
}

func TestRuntimeErrorResetsStack(t *testing.T) {
	t.Run("Un error en un operador unario vacía el stack", func(t *testing.T) {
		vm, chunk := setUpVM(t)
		chunk.Write(byte(c.OP_FALSE), line)
		chunk.Write(byte(c.OP_NEGATE), line)

		err := vm.Run()

		utils.AssertErrorMessage(t, err, "operand must be a number")
		utils.AssertValue(t, len(vm.stack), 0)
	})

	t.Run("Un error en un operador binario vacía el stack", func(t *testing.T) {
		vm, chunk := setUpVM(t)
		writeConstant(chunk, "a")
		writeConstant(chunk, 1.0)
		chunk.Write(byte(c.OP_ADD), line)

		err := vm.Run()

		utils.AssertErrorMessage(t, err, "operands must be two numbers or two strings")
		utils.AssertValue(t, len(vm.stack), 0)
	})

	t.Run("También se descartan los valores que había debajo de la operación que falla", func(t *testing.T) {
		vm, chunk := setUpVM(t)
		writeConstant(chunk, 10.0)
		chunk.Write(byte(c.OP_TRUE), line)
		chunk.Write(byte(c.OP_NEGATE), line)

		err := vm.Run()

		utils.AssertErrorMessage(t, err, "operand must be a number")
		utils.AssertValue(t, len(vm.stack), 0)
	})
}
