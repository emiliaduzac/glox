package compiler

import (
	"slices"
	"testing"
	c "tp-lox/internal/chunk"
	"tp-lox/internal/scanner"
	"tp-lox/internal/utils"
)

// compileSource escanea y compila el código fuente, y devuelve el chunk generado.
func compileSource(t *testing.T, source string) *c.Chunk {
	t.Helper()
	tokens, err := scanner.NewScanner(source).Scan()
	utils.AssertNoError(t, err)

	chunk, err := NewCompiler(tokens).Compile()
	utils.AssertNoError(t, err)
	return chunk
}

// compileSourceError escanea el código fuente (que debe ser válido para el scanner) y devuelve el error del compilador.
func compileSourceError(t *testing.T, source string) error {
	t.Helper()
	tokens, err := scanner.NewScanner(source).Scan()
	utils.AssertNoError(t, err)

	_, err = NewCompiler(tokens).Compile()
	return err
}

// assertBytes verifica que los bytes del chunk sean exactamente los esperados.
func assertBytes(t *testing.T, chunk *c.Chunk, expected ...byte) {
	t.Helper()
	if !slices.Equal(chunk.Bytes, expected) {
		t.Errorf("expected bytes %v, got %v", expected, chunk.Bytes)
	}
}

func TestCompileLiterals(t *testing.T) {
	tests := []struct {
		name   string
		source string
		op     c.OpCode
	}{
		{"nil", "nil", c.OP_NIL},
		{"true", "true", c.OP_TRUE},
		{"false", "false", c.OP_FALSE},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			chunk := compileSource(t, test.source)

			// opcode del literal y se cierra siempre con OP_RETURN
			assertBytes(t, chunk, byte(test.op), byte(c.OP_RETURN))
		})
	}
}

func TestCompileNumbers(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		expected float64
	}{
		{"entero", "1", 1.0},
		{"cero", "0", 0.0},
		{"decimal", "3.14", 3.14},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			chunk := compileSource(t, test.source)

			// NUM -> CONSTANT 0 (el indice del pool), RETURN
			assertBytes(t, chunk, byte(c.OP_CONSTANT), 0, byte(c.OP_RETURN))
			utils.AssertValue(t, len(chunk.Constants), 1)
			utils.AssertValue(t, chunk.Constants[0], test.expected)
		})
	}
}

func TestCompileUnary(t *testing.T) {
	t.Run("Negación de un número", func(t *testing.T) {
		chunk := compileSource(t, "-1")

		// -1 -> CONSTANT 0 (el indice del 1 en el pool), NEGATE, RETURN
		assertBytes(t, chunk, byte(c.OP_CONSTANT), 0, byte(c.OP_NEGATE), byte(c.OP_RETURN))
		utils.AssertValue(t, chunk.Constants[0], 1.0)
	})

	t.Run("Not de un booleano", func(t *testing.T) {
		chunk := compileSource(t, "!true")

		assertBytes(t, chunk, byte(c.OP_TRUE), byte(c.OP_NOT), byte(c.OP_RETURN))
	})

	t.Run("Operadores negacion anidados", func(t *testing.T) {
		chunk := compileSource(t, "--1")

		assertBytes(t, chunk, byte(c.OP_CONSTANT), 0, byte(c.OP_NEGATE), byte(c.OP_NEGATE), byte(c.OP_RETURN))
	})
}

func TestCompileBinary(t *testing.T) {
	tests := []struct {
		name   string
		source string
		op     c.OpCode
	}{
		{"suma", "1 + 2", c.OP_ADD},
		{"resta", "1 - 2", c.OP_SUBTRACT},
		{"multiplicación", "1 * 2", c.OP_MULTIPLY},
		{"división", "1 / 2", c.OP_DIVIDE},
		{"módulo", "1 % 2", c.OP_MODULO},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			chunk := compileSource(t, test.source)

			// 1 OP 2 -> CONSTANT 0 (el primer numero), CONSTANT 1 (el 2do), OP, RETURN
			assertBytes(t, chunk,
				byte(c.OP_CONSTANT), 0,
				byte(c.OP_CONSTANT), 1,
				byte(test.op),
				byte(c.OP_RETURN),
			)
			utils.AssertValue(t, len(chunk.Constants), 2)
			utils.AssertValue(t, chunk.Constants[0], 1.0)
			utils.AssertValue(t, chunk.Constants[1], 2.0)
		})
	}

	t.Run("La multiplicación tiene precedencia sobre la suma", func(t *testing.T) {
		chunk := compileSource(t, "1 + 2 * 3")

		// 1 + 2 * 3 -> CONSTANT 0, CONSTANT 1, CONSTANT 2, MULTIPLY, ADD, RETURN
		assertBytes(t, chunk,
			byte(c.OP_CONSTANT), 0,
			byte(c.OP_CONSTANT), 1,
			byte(c.OP_CONSTANT), 2,
			byte(c.OP_MULTIPLY),
			byte(c.OP_ADD),
			byte(c.OP_RETURN),
		)
	})

	t.Run("Los operadores del mismo nivel asocian a izquierda", func(t *testing.T) {
		chunk := compileSource(t, "1 - 2 - 3")

		assertBytes(t, chunk,
			byte(c.OP_CONSTANT), 0,
			byte(c.OP_CONSTANT), 1,
			byte(c.OP_SUBTRACT),
			byte(c.OP_CONSTANT), 2,
			byte(c.OP_SUBTRACT),
			byte(c.OP_RETURN),
		)
	})
}

func TestCompileGrouping(t *testing.T) {
	t.Run("Los paréntesis cambian el orden de evaluación", func(t *testing.T) {
		chunk := compileSource(t, "(1 + 2) * 3")

		assertBytes(t, chunk,
			byte(c.OP_CONSTANT), 0,
			byte(c.OP_CONSTANT), 1,
			byte(c.OP_ADD),
			byte(c.OP_CONSTANT), 2,
			byte(c.OP_MULTIPLY),
			byte(c.OP_RETURN),
		)
	})

	t.Run("Error si falta el paréntesis de cierre", func(t *testing.T) {
		err := compileSourceError(t, "(1 + 2")

		utils.AssertErrorMessage(t, err, "')' after expression")
	})
}

func TestCompileComparison(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		expected []c.OpCode
	}{
		{"igual", "1 == 2", []c.OpCode{c.OP_EQUAL}},
		{"mayor", "1 > 2", []c.OpCode{c.OP_GREATER}},
		{"menor", "1 < 2", []c.OpCode{c.OP_LESS}},
		{"distinto", "1 != 2", []c.OpCode{c.OP_EQUAL, c.OP_NOT}},
		{"mayor o igual", "1 >= 2", []c.OpCode{c.OP_LESS, c.OP_NOT}},
		{"menor o igual", "1 <= 2", []c.OpCode{c.OP_GREATER, c.OP_NOT}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			chunk := compileSource(t, test.source)

			expected := []byte{byte(c.OP_CONSTANT), 0, byte(c.OP_CONSTANT), 1}
			for _, op := range test.expected {
				expected = append(expected, byte(op))
			}
			expected = append(expected, byte(c.OP_RETURN))
			assertBytes(t, chunk, expected...)
		})
	}

	t.Run("Se puede comparar con nil y booleanos", func(t *testing.T) {
		chunk := compileSource(t, "nil == false")

		assertBytes(t, chunk, byte(c.OP_NIL), byte(c.OP_FALSE), byte(c.OP_EQUAL), byte(c.OP_RETURN))
	})

	t.Run("La suma tiene precedencia sobre la comparación", func(t *testing.T) {
		chunk := compileSource(t, "1 + 2 < 3")

		assertBytes(t, chunk,
			byte(c.OP_CONSTANT), 0,
			byte(c.OP_CONSTANT), 1,
			byte(c.OP_ADD),
			byte(c.OP_CONSTANT), 2,
			byte(c.OP_LESS),
			byte(c.OP_RETURN),
		)
	})
}

func TestCompileErrors(t *testing.T) {
	tests := []struct {
		name          string
		source        string
		expectedError string
	}{
		{"entrada vacía", "", "expected expression"},
		{"falta el operando derecho", "1 +", "expected expression"},
		{"sobra un token", "1 2", "expected end of expression"},
	}

	for _, test := range tests {
		t.Run("Error si "+test.name, func(t *testing.T) {
			err := compileSourceError(t, test.source)

			utils.AssertErrorMessage(t, err, test.expectedError)
		})
	}
}
