package compiler

import (
	"slices"
	"strings"
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

func compileExpr(t *testing.T, expr string) *c.Chunk {
	t.Helper()
	return compileSource(t, "print "+expr+";")
}

func compileExprError(t *testing.T, expr string) error {
	t.Helper()
	return compileSourceError(t, "print "+expr+";")
}

func compileErrors(t *testing.T, source string) []error {
	t.Helper()
	err := compileSourceError(t, source)
	joined, ok := err.(interface{ Unwrap() []error })
	if !ok {
		t.Fatalf("expected a list of errors, got %v", err)
	}
	return joined.Unwrap()
}

// assertBytes verifica que los bytes del chunk sean exactamente los esperados.
func assertBytes(t *testing.T, chunk *c.Chunk, expected ...byte) {
	t.Helper()
	if !slices.Equal(chunk.Bytes, expected) {
		t.Errorf("expected bytes %v, got %v", expected, chunk.Bytes)
	}
}

func assertExprBytes(t *testing.T, chunk *c.Chunk, expected ...byte) {
	t.Helper()
	assertBytes(t, chunk, slices.Concat(expected, []byte{byte(c.OP_PRINT), byte(c.OP_RETURN)})...)
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
			chunk := compileExpr(t, test.source)

			// opcode del literal y se cierra siempre con OP_PRINT y OP_RETURN
			assertExprBytes(t, chunk, byte(test.op))
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
			chunk := compileExpr(t, test.source)

			// NUM -> CONSTANT 0 (el indice del pool), PRINT, RETURN
			assertExprBytes(t, chunk, byte(c.OP_CONSTANT), 0)
			utils.AssertValue(t, len(chunk.Constants), 1)
			utils.AssertValue(t, chunk.Constants[0], test.expected)
		})
	}
}

func TestCompileStrings(t *testing.T) {
	chunk := compileExpr(t, `"hola"`)

	assertExprBytes(t, chunk, byte(c.OP_CONSTANT), 0)
	utils.AssertValue(t, len(chunk.Constants), 1)
	utils.AssertValue(t, chunk.Constants[0], "hola")
}

func TestCompileUnary(t *testing.T) {
	t.Run("Negación de un número", func(t *testing.T) {
		chunk := compileExpr(t, "-1")

		// -1 -> CONSTANT 0 (el indice del 1 en el pool), NEGATE, PRINT, RETURN
		assertExprBytes(t, chunk, byte(c.OP_CONSTANT), 0, byte(c.OP_NEGATE))
		utils.AssertValue(t, chunk.Constants[0], 1.0)
	})

	t.Run("Not de un booleano", func(t *testing.T) {
		chunk := compileExpr(t, "!true")

		assertExprBytes(t, chunk, byte(c.OP_TRUE), byte(c.OP_NOT))
	})

	t.Run("Operadores negacion anidados", func(t *testing.T) {
		chunk := compileExpr(t, "--1")

		assertExprBytes(t, chunk, byte(c.OP_CONSTANT), 0, byte(c.OP_NEGATE), byte(c.OP_NEGATE))
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
			chunk := compileExpr(t, test.source)

			// 1 OP 2 -> CONSTANT 0 (el primer numero), CONSTANT 1 (el 2do), OP, PRINT, RETURN
			assertExprBytes(t, chunk,
				byte(c.OP_CONSTANT), 0,
				byte(c.OP_CONSTANT), 1,
				byte(test.op),
			)
			utils.AssertValue(t, len(chunk.Constants), 2)
			utils.AssertValue(t, chunk.Constants[0], 1.0)
			utils.AssertValue(t, chunk.Constants[1], 2.0)
		})
	}

	t.Run("La multiplicación tiene precedencia sobre la suma", func(t *testing.T) {
		chunk := compileExpr(t, "1 + 2 * 3")

		// 1 + 2 * 3 -> CONSTANT 0, CONSTANT 1, CONSTANT 2, MULTIPLY, ADD, PRINT, RETURN
		assertExprBytes(t, chunk,
			byte(c.OP_CONSTANT), 0,
			byte(c.OP_CONSTANT), 1,
			byte(c.OP_CONSTANT), 2,
			byte(c.OP_MULTIPLY),
			byte(c.OP_ADD),
		)
	})

	t.Run("Los operadores del mismo nivel asocian a izquierda", func(t *testing.T) {
		chunk := compileExpr(t, "1 - 2 - 3")

		assertExprBytes(t, chunk,
			byte(c.OP_CONSTANT), 0,
			byte(c.OP_CONSTANT), 1,
			byte(c.OP_SUBTRACT),
			byte(c.OP_CONSTANT), 2,
			byte(c.OP_SUBTRACT),
		)
	})
}

func TestCompileGrouping(t *testing.T) {
	t.Run("Los paréntesis cambian el orden de evaluación", func(t *testing.T) {
		chunk := compileExpr(t, "(1 + 2) * 3")

		assertExprBytes(t, chunk,
			byte(c.OP_CONSTANT), 0,
			byte(c.OP_CONSTANT), 1,
			byte(c.OP_ADD),
			byte(c.OP_CONSTANT), 2,
			byte(c.OP_MULTIPLY),
		)
	})

	t.Run("Error si falta el paréntesis de cierre", func(t *testing.T) {
		err := compileExprError(t, "(1 + 2")

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
			chunk := compileExpr(t, test.source)

			expected := []byte{byte(c.OP_CONSTANT), 0, byte(c.OP_CONSTANT), 1}
			for _, op := range test.expected {
				expected = append(expected, byte(op))
			}
			assertExprBytes(t, chunk, expected...)
		})
	}

	t.Run("Se puede comparar con nil y booleanos", func(t *testing.T) {
		chunk := compileExpr(t, "nil == false")

		assertExprBytes(t, chunk, byte(c.OP_NIL), byte(c.OP_FALSE), byte(c.OP_EQUAL))
	})

	t.Run("La suma tiene precedencia sobre la comparación", func(t *testing.T) {
		chunk := compileExpr(t, "1 + 2 < 3")

		assertExprBytes(t, chunk,
			byte(c.OP_CONSTANT), 0,
			byte(c.OP_CONSTANT), 1,
			byte(c.OP_ADD),
			byte(c.OP_CONSTANT), 2,
			byte(c.OP_LESS),
		)
	})
}

func TestCompileErrors(t *testing.T) {
	tests := []struct {
		name          string
		source        string
		expectedError string
	}{
		{"falta la expresión", "", "expected expression"},
		{"falta el operando derecho", "1 +", "expected expression"},
		{"sobra un token", "1 2", "expected ';' after statement"},
	}

	for _, test := range tests {
		t.Run("Error si "+test.name, func(t *testing.T) {
			err := compileExprError(t, test.source)

			utils.AssertErrorMessage(t, err, test.expectedError)
		})
	}
}

func TestCompileConstantLimit(t *testing.T) {
	t.Run("Un chunk admite hasta 256 constantes", func(t *testing.T) {
		chunk := compileExpr(t, strings.Repeat("1 + ", 255)+"1")

		utils.AssertValue(t, len(chunk.Constants), 256)
	})

	t.Run("Error si un chunk tiene más de 256 constantes", func(t *testing.T) {
		err := compileExprError(t, strings.Repeat("1 + ", 256)+"1")

		utils.AssertErrorMessage(t, err, "too many constants in one chunk")
	})
}

func TestCompileProgram(t *testing.T) {
	t.Run("Un programa vacío solo tiene OP_RETURN", func(t *testing.T) {
		chunk := compileSource(t, "")

		assertBytes(t, chunk, byte(c.OP_RETURN))
	})

	t.Run("Compila varias declaraciones en orden", func(t *testing.T) {
		chunk := compileSource(t, "print 1; print 2;")

		assertBytes(t, chunk,
			byte(c.OP_CONSTANT), 0,
			byte(c.OP_PRINT),
			byte(c.OP_CONSTANT), 1,
			byte(c.OP_PRINT),
			byte(c.OP_RETURN),
		)
	})

	t.Run("Error si falta el punto y coma después de un print", func(t *testing.T) {
		err := compileSourceError(t, "print 1")

		utils.AssertErrorMessage(t, err, "';' after statement")
	})

	t.Run("Una expresión como statement se descarta con OP_POP", func(t *testing.T) {
		chunk := compileSource(t, "1;")

		assertBytes(t, chunk, byte(c.OP_CONSTANT), 0, byte(c.OP_POP), byte(c.OP_RETURN))
	})

	t.Run("Error si falta el punto y coma en una expresión", func(t *testing.T) {
		err := compileSourceError(t, "1 + 2")

		utils.AssertErrorMessage(t, err, "';' after statement")
	})
}

func TestCompileErrorSynchronization(t *testing.T) {
	t.Run("Levanta un error por cada statement que falla", func(t *testing.T) {
		errs := compileErrors(t, "print 1 +;\nprint 2 +;")

		utils.AssertValue(t, len(errs), 2)
		utils.AssertErrorMessage(t, errs[0], "[line 1] Error: expected expression")
		utils.AssertErrorMessage(t, errs[1], "[line 2] Error: expected expression")
	})

	t.Run("Sincroniza en una palabra clave aunque falte el punto y coma", func(t *testing.T) {
		errs := compileErrors(t, "print 1 2 3 print 4 +;")

		utils.AssertValue(t, len(errs), 2)
		utils.AssertErrorMessage(t, errs[0], "';' after statement")
		utils.AssertErrorMessage(t, errs[1], "expected expression")
	})

	t.Run("Saltea los tokens que sobran hasta el punto y coma sin generar errores en cascada", func(t *testing.T) {
		errs := compileErrors(t, "print 1 2 3;\nprint 4;")

		utils.AssertValue(t, len(errs), 1)
	})

	t.Run("Un programa con errores no se compila aunque tenga statements válidos", func(t *testing.T) {
		tokens, err := scanner.NewScanner("print 1 +; print 2;").Scan()
		utils.AssertNoError(t, err)

		chunk, err := NewCompiler(tokens).Compile()

		utils.AssertError(t, err)
		utils.AssertValue(t, chunk == nil, true)
	})
}

func TestCompileVarDeclaration(t *testing.T) {
	t.Run("Una variable declarada sin valor vale nil", func(t *testing.T) {
		chunk := compileSource(t, "var a;")

		assertBytes(t, chunk, byte(c.OP_NIL), byte(c.OP_DEFINE_GLOBAL), 0, byte(c.OP_RETURN))
		utils.AssertValue(t, len(chunk.Constants), 1)
		utils.AssertValue(t, chunk.Constants[0], "a")
	})

	t.Run("Una variable declarada con un valor guarda ese valor", func(t *testing.T) {
		chunk := compileSource(t, "var a = 1;")

		assertBytes(t, chunk, byte(c.OP_CONSTANT), 1, byte(c.OP_DEFINE_GLOBAL), 0, byte(c.OP_RETURN))
		utils.AssertValue(t, len(chunk.Constants), 2)
		utils.AssertValue(t, chunk.Constants[0], "a")
		utils.AssertValue(t, chunk.Constants[1], 1.0)
	})

	t.Run("Error si falta el nombre de la variable", func(t *testing.T) {
		err := compileSourceError(t, "var 1;")

		utils.AssertErrorMessage(t, err, "identifier after 'var'")
	})

	t.Run("Error si falta el punto y coma", func(t *testing.T) {
		err := compileSourceError(t, "var a")

		utils.AssertErrorMessage(t, err, "';' after variable declaration")
	})
}

func TestCompileVariableRead(t *testing.T) {
	t.Run("Leer una variable se compila a OP_GET_GLOBAL con su nombre", func(t *testing.T) {
		chunk := compileSource(t, "print a;")

		assertBytes(t, chunk, byte(c.OP_GET_GLOBAL), 0, byte(c.OP_PRINT), byte(c.OP_RETURN))
		utils.AssertValue(t, len(chunk.Constants), 1)
		utils.AssertValue(t, chunk.Constants[0], "a")
	})
}

func TestCompileAssignment(t *testing.T) {
	t.Run("Asignar una variable se compila a OP_SET_GLOBAL con su nombre", func(t *testing.T) {
		chunk := compileSource(t, "a = 1;")

		assertBytes(t, chunk, byte(c.OP_CONSTANT), 1, byte(c.OP_SET_GLOBAL), 0, byte(c.OP_POP), byte(c.OP_RETURN))
		utils.AssertValue(t, len(chunk.Constants), 2)
		utils.AssertValue(t, chunk.Constants[0], "a")
		utils.AssertValue(t, chunk.Constants[1], 1.0)
	})

	t.Run("Una asignación es una expresión y su valor se puede usar", func(t *testing.T) {
		chunk := compileSource(t, "print a = 2;")

		assertBytes(t, chunk, byte(c.OP_CONSTANT), 1, byte(c.OP_SET_GLOBAL), 0, byte(c.OP_PRINT), byte(c.OP_RETURN))
	})

	t.Run("La asignación asocia a derecha", func(t *testing.T) {
		chunk := compileSource(t, "a = b = 1;")

		assertBytes(t, chunk,
			byte(c.OP_CONSTANT), 2,
			byte(c.OP_SET_GLOBAL), 1,
			byte(c.OP_SET_GLOBAL), 0,
			byte(c.OP_POP),
			byte(c.OP_RETURN),
		)
		utils.AssertValue(t, chunk.Constants[0], "a")
		utils.AssertValue(t, chunk.Constants[1], "b")
	})

	t.Run("Error si falta el valor a asignar", func(t *testing.T) {
		err := compileSourceError(t, "a = ;")

		utils.AssertErrorMessage(t, err, "expected expression")
	})
}
