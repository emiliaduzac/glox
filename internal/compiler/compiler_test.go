package compiler

import (
	"fmt"
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

	function, err := NewCompiler(tokens).Compile()
	utils.AssertNoError(t, err)
	return &function.Chunk
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

// assertBytes verifica que los bytes del chunk sean exactamente los esperados
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

func TestCompileScript(t *testing.T) {
	t.Run("Un programa se compila a una función sin nombre y sin parámetros, que es el script", func(t *testing.T) {
		tokens, err := scanner.NewScanner("print 1;").Scan()
		utils.AssertNoError(t, err)

		function, err := NewCompiler(tokens).Compile()

		utils.AssertNoError(t, err)
		utils.AssertValue(t, function.Name, "")
		utils.AssertValue(t, function.Arity, 0)
		utils.AssertValue(t, utils.Stringify(function), "<script>")
		assertBytes(t, &function.Chunk, byte(c.OP_CONSTANT), 0, byte(c.OP_PRINT), byte(c.OP_RETURN))
	})
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

func TestCompileVariableNameConstants(t *testing.T) {
	t.Run("Usar la misma variable varias veces guarda su nombre una sola vez", func(t *testing.T) {
		chunk := compileSource(t, "print a; print a;")

		assertBytes(t, chunk,
			byte(c.OP_GET_GLOBAL), 0,
			byte(c.OP_PRINT),
			byte(c.OP_GET_GLOBAL), 0,
			byte(c.OP_PRINT),
			byte(c.OP_RETURN),
		)
		utils.AssertValue(t, len(chunk.Constants), 1)
		utils.AssertValue(t, chunk.Constants[0], "a")
	})

	t.Run("Variables distintas guardan cada una su nombre", func(t *testing.T) {
		chunk := compileSource(t, "print a; print b;")

		assertBytes(t, chunk,
			byte(c.OP_GET_GLOBAL), 0,
			byte(c.OP_PRINT),
			byte(c.OP_GET_GLOBAL), 1,
			byte(c.OP_PRINT),
			byte(c.OP_RETURN),
		)
		utils.AssertValue(t, len(chunk.Constants), 2)
		utils.AssertValue(t, chunk.Constants[0], "a")
		utils.AssertValue(t, chunk.Constants[1], "b")
	})

	t.Run("Un programa que usa muchas veces la misma variable no llena las constantes", func(t *testing.T) {
		chunk := compileSource(t, "var a = 0;\n"+strings.Repeat("a = a + 1;\n", 130))

		utils.AssertValue(t, len(chunk.Constants), 132)
	})
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

		function, err := NewCompiler(tokens).Compile()

		utils.AssertError(t, err)
		utils.AssertValue(t, function == nil, true)
	})

	t.Run("Un error dentro de un bloque anidado se muestra una sola vez", func(t *testing.T) {
		errs := compileErrors(t, `{ var a = "x"; { var a = a; } }`)

		utils.AssertValue(t, len(errs), 1)
		utils.AssertErrorMessage(t, errs[0], "can't read local variable in its own initializer")
	})

	t.Run("Un error dentro de un bloque no oculta los errores de lo que sigue", func(t *testing.T) {
		errs := compileErrors(t, "{ print 1 +; }\na = ;")

		utils.AssertValue(t, len(errs), 2)
		utils.AssertErrorMessage(t, errs[0], "[line 1] Error: expected expression")
		utils.AssertErrorMessage(t, errs[1], "[line 2] Error: expected expression")
	})

	t.Run("Mientras hay un error sin sincronizar no se reportan otros del mismo statement", func(t *testing.T) {
		errs := compileErrors(t, "{ var a = a + ; }")

		utils.AssertValue(t, len(errs), 1)
		utils.AssertErrorMessage(t, errs[0], "can't read local variable in its own initializer")
	})

	t.Run("Leer una variable local en su propio inicializador no genera errores falsos más adelante", func(t *testing.T) {
		errs := compileErrors(t, "{ var a = a; }\n{ var a = 1; }")

		utils.AssertValue(t, len(errs), 1)
		utils.AssertErrorMessage(t, errs[0], "can't read local variable in its own initializer")
	})
}

func TestCompileGlobalVarDeclaration(t *testing.T) {
	t.Run("Una variable global declarada sin valor vale nil", func(t *testing.T) {
		chunk := compileSource(t, "var a;")

		assertBytes(t, chunk, byte(c.OP_NIL), byte(c.OP_DEFINE_GLOBAL), 0, byte(c.OP_RETURN))
		utils.AssertValue(t, len(chunk.Constants), 1)
		utils.AssertValue(t, chunk.Constants[0], "a")
	})

	t.Run("Una variable global declarada con un valor guarda ese valor", func(t *testing.T) {
		chunk := compileSource(t, "var a = 1;")

		assertBytes(t, chunk, byte(c.OP_CONSTANT), 1, byte(c.OP_DEFINE_GLOBAL), 0, byte(c.OP_RETURN))
		utils.AssertValue(t, len(chunk.Constants), 2)
		utils.AssertValue(t, chunk.Constants[0], "a")
		utils.AssertValue(t, chunk.Constants[1], 1.0)
	})

	t.Run("Error si falta el nombre de la variable global", func(t *testing.T) {
		err := compileSourceError(t, "var 1;")

		utils.AssertErrorMessage(t, err, "identifier after 'var'")
	})

	t.Run("Error si falta el punto y coma", func(t *testing.T) {
		err := compileSourceError(t, "var a")

		utils.AssertErrorMessage(t, err, "';' after variable declaration")
	})
}

func TestCompileGlobalVariableRead(t *testing.T) {
	t.Run("Leer una variable global se compila a OP_GET_GLOBAL con su nombre", func(t *testing.T) {
		chunk := compileSource(t, "print a;")

		assertBytes(t, chunk, byte(c.OP_GET_GLOBAL), 0, byte(c.OP_PRINT), byte(c.OP_RETURN))
		utils.AssertValue(t, len(chunk.Constants), 1)
		utils.AssertValue(t, chunk.Constants[0], "a")
	})
}

func TestCompileGlobalAssignment(t *testing.T) {
	t.Run("Asignar una variable global se compila a OP_SET_GLOBAL con su nombre", func(t *testing.T) {
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

	t.Run("Una variable dentro del valor asignado se lee", func(t *testing.T) {
		chunk := compileSource(t, `saludo = "hola " + Lionel;`)

		assertBytes(t, chunk,
			byte(c.OP_CONSTANT), 1,
			byte(c.OP_GET_GLOBAL), 2,
			byte(c.OP_ADD),
			byte(c.OP_SET_GLOBAL), 0,
			byte(c.OP_POP),
			byte(c.OP_RETURN),
		)
		utils.AssertValue(t, chunk.Constants[0], "saludo")
		utils.AssertValue(t, chunk.Constants[2], "Lionel")
	})

	invalidTargets := []struct {
		name   string
		source string
	}{
		{"una variable dentro de un operador binario", "a * b = c + d;"},
		{"una variable dentro de un operador unario", "-a = 1;"},
		{"algo que no es una variable", "1 = 2;"},
		{"una variable entre paréntesis", "(a) = 5;"},
	}

	for _, test := range invalidTargets {
		t.Run("Error si se asigna a "+test.name, func(t *testing.T) {
			err := compileSourceError(t, test.source)

			utils.AssertErrorMessage(t, err, "invalid assignment target")
		})
	}
}

func TestCompileBlock(t *testing.T) {
	t.Run("Los bloques se compilan correctamente", func(t *testing.T) {
		chunk := compileSource(t, "var a = 1; { print a; }")

		assertBytes(t, chunk, byte(c.OP_CONSTANT), 1, byte(c.OP_DEFINE_GLOBAL), 0, byte(c.OP_GET_GLOBAL), 0, byte(c.OP_PRINT), byte(c.OP_RETURN))
		utils.AssertValue(t, len(chunk.Constants), 2)
		utils.AssertValue(t, chunk.Constants[0], "a")
		utils.AssertValue(t, chunk.Constants[1], 1.0)
	})

	t.Run("Error si el bloque no se cierra", func(t *testing.T) {
		err := compileSourceError(t, "var a = 1; { print a; ")

		utils.AssertErrorMessage(t, err, "'}' after block")
	})
}

func TestCompileLocalVarDeclaration(t *testing.T) {
	t.Run("Declarar una variable local no emite una instrucción especial", func(t *testing.T) {
		chunk := compileSource(t, "{ var a = 1; }")

		assertBytes(t, chunk, byte(c.OP_CONSTANT), 0, byte(c.OP_POP), byte(c.OP_RETURN)) // el pop es por el cierre del scope
		utils.AssertValue(t, len(chunk.Constants), 1)
		utils.AssertValue(t, chunk.Constants[0], 1.0)
	})

	t.Run("Declarar una variable local sin valor es nil", func(t *testing.T) {
		chunk := compileSource(t, "{ var a; }")

		assertBytes(t, chunk, byte(c.OP_NIL), byte(c.OP_POP), byte(c.OP_RETURN))
		utils.AssertValue(t, len(chunk.Constants), 0)
	})

	t.Run("Error si falta el nombre de la variable local", func(t *testing.T) {
		err := compileSourceError(t, "{ var 1; }")

		utils.AssertErrorMessage(t, err, "identifier after 'var'")
	})

	t.Run("Error si falta el punto y coma", func(t *testing.T) {
		err := compileSourceError(t, "{ var a }")

		utils.AssertErrorMessage(t, err, "';' after variable declaration")
	})

	t.Run("Error si se redeclara la misma variable local", func(t *testing.T) {
		err := compileSourceError(t, "{ var a = 1; var a = 2; }")

		utils.AssertErrorMessage(t, err, "already a variable with this name in this scope")
	})
}

func TestCompileLocalLimit(t *testing.T) {
	localsSource := func(n int) string {
		var sb strings.Builder
		sb.WriteString("{ ")
		for i := 0; i < n; i++ {
			fmt.Fprintf(&sb, "var a%d; ", i)
		}
		sb.WriteString("}")
		return sb.String()
	}

	t.Run("Un scope admite hasta 255 variables locales, porque el slot cero está reservado", func(t *testing.T) {
		chunk := compileSource(t, localsSource(255))

		utils.AssertValue(t, len(chunk.Constants), 0)
	})

	t.Run("Error si un scope tiene más de 255 variables locales", func(t *testing.T) {
		err := compileSourceError(t, localsSource(256))

		utils.AssertErrorMessage(t, err, "too many local variables in scope")
	})
}

func TestCompileLocalVariableRead(t *testing.T) {
	t.Run("Leer una variable local se compila a OP_GET_LOCAL con su indice, no con su nombre", func(t *testing.T) {
		chunk := compileSource(t, "{ var a = 1; print a; }")

		assertBytes(t, chunk, byte(c.OP_CONSTANT), 0, byte(c.OP_GET_LOCAL), 1, byte(c.OP_PRINT), byte(c.OP_POP), byte(c.OP_RETURN))
		utils.AssertValue(t, len(chunk.Constants), 1)
		utils.AssertValue(t, chunk.Constants[0], 1.0)
	})

	t.Run("Usa el indice que corresponde a cada variable", func(t *testing.T) {
		chunk := compileSource(t, "{ var a = 1; var b = 2; print b; print a; }")

		assertBytes(t, chunk,
			byte(c.OP_CONSTANT), 0, // a: indice 1
			byte(c.OP_CONSTANT), 1, // b: indice 2
			byte(c.OP_GET_LOCAL), 2, // print b
			byte(c.OP_PRINT),
			byte(c.OP_GET_LOCAL), 1, // print a
			byte(c.OP_PRINT),
			byte(c.OP_POP),
			byte(c.OP_POP),
			byte(c.OP_RETURN),
		)
	})

	t.Run("Una variable local de un bloque interno oculta a la de afuera (shadowing)", func(t *testing.T) {
		chunk := compileSource(t, "{ var a = 1; { var a = 2; print a; } print a; }")

		assertBytes(t, chunk,
			byte(c.OP_CONSTANT), 0, // a externa, indice 1
			byte(c.OP_CONSTANT), 1, // a interna, indice 2
			byte(c.OP_GET_LOCAL), 2, // print a: lee la interna
			byte(c.OP_PRINT),
			byte(c.OP_POP),
			byte(c.OP_GET_LOCAL), 1, // print a fuera del bloque interno: lee la externa
			byte(c.OP_PRINT),
			byte(c.OP_POP),
			byte(c.OP_RETURN),
		)
	})

	t.Run("Error si se lee una variable local en su propio inicializador", func(t *testing.T) {
		err := compileSourceError(t, `{ var a = "outer"; { var a = a; } }`)
		utils.AssertErrorMessage(t, err, "can't read local variable in its own initializer")
	})
}

func TestCompileLocalAssignment(t *testing.T) {
	t.Run("Asignar una variable local se compila a OP_SET_LOCAL con su indice", func(t *testing.T) {
		chunk := compileSource(t, "{ var a = 1; a = 2; }")

		assertBytes(t, chunk,
			byte(c.OP_CONSTANT), 0, // el 1
			byte(c.OP_CONSTANT), 1, // el 2
			byte(c.OP_SET_LOCAL), 1, // a esta en la posicion 1 de locals (la 0 es el slot reservado)
			byte(c.OP_POP),
			byte(c.OP_POP),
			byte(c.OP_RETURN),
		)
		utils.AssertValue(t, len(chunk.Constants), 2) // en las constantes van los valores, no el nombre
	})

	t.Run("Una asignación a una local es una expresión y su valor se puede usar", func(t *testing.T) {
		chunk := compileSource(t, "{ var a = 1; print a = 2; }")

		assertBytes(t, chunk,
			byte(c.OP_CONSTANT), 0,
			byte(c.OP_CONSTANT), 1,
			byte(c.OP_SET_LOCAL), 1,
			byte(c.OP_PRINT),
			byte(c.OP_POP),
			byte(c.OP_RETURN),
		)
	})

	t.Run("La asignación entre locales asocia a derecha", func(t *testing.T) {
		chunk := compileSource(t, "{ var a; var b; a = b = 1; }")

		assertBytes(t, chunk,
			byte(c.OP_NIL),
			byte(c.OP_NIL),
			byte(c.OP_CONSTANT), 0,
			byte(c.OP_SET_LOCAL), 2,
			byte(c.OP_SET_LOCAL), 1,
			byte(c.OP_POP),
			byte(c.OP_POP),
			byte(c.OP_POP),
			byte(c.OP_RETURN),
		)
	})

	t.Run("Una local dentro del valor asignado se lee", func(t *testing.T) {
		chunk := compileSource(t, "{ var a = 1; a = a + 1; }")

		assertBytes(t, chunk,
			byte(c.OP_CONSTANT), 0,
			byte(c.OP_GET_LOCAL), 1,
			byte(c.OP_CONSTANT), 1,
			byte(c.OP_ADD),
			byte(c.OP_SET_LOCAL), 1,
			byte(c.OP_POP),
			byte(c.OP_POP),
			byte(c.OP_RETURN),
		)
	})
}

func TestCompileLocalAndGlobalVariables(t *testing.T) {
	t.Run("Distingue variables globales y locales", func(t *testing.T) {
		chunk := compileSource(t, "var x = 1; { var y = 2; print x; print y; }")

		assertBytes(t, chunk,
			byte(c.OP_CONSTANT), 1,
			byte(c.OP_DEFINE_GLOBAL), 0,
			byte(c.OP_CONSTANT), 2,
			byte(c.OP_GET_GLOBAL), 0,
			byte(c.OP_PRINT),
			byte(c.OP_GET_LOCAL), 1,
			byte(c.OP_PRINT),
			byte(c.OP_POP),
			byte(c.OP_RETURN),
		)
	})

	t.Run("Al terminar no quedan locales de usuario ni scopes abiertos", func(t *testing.T) {
		tokens, _ := scanner.NewScanner("{ var a = 1; print 1 +; } print 2;").Scan()
		compiler := NewCompiler(tokens)
		_, err := compiler.Compile()

		utils.AssertError(t, err)
		utils.AssertValue(t, len(compiler.locals), 1)
		utils.AssertValue(t, compiler.scopeDepth, 0)
	})
}

const (
	literalSize  = 1 // OP_TRUE - OP_FALSE - OP_NIL
	popSize      = 1
	printSize    = 1
	constantSize = 2 // OP_CONSTANT + indice
	jumpSize     = 3 // OP_JUMP - OP_JUMP_IF_FALSE + 2 bytes de offset
	loopSize     = 3 // OP_LOOP + 2 bytes de offset
	getLocalSize = 2 // OP_GET_LOCAL + índice

	// offsets de los saltos para un then o un else de la forma `print n;`
	skipThen = popSize + constantSize + printSize + jumpSize // el then y el JUMP que saltea el else
	skipElse = popSize + constantSize + printSize            // el POP de la condición y el else
)

func TestCompileIfWithoutElse(t *testing.T) {
	t.Run("If sin else", func(t *testing.T) {
		compiled := compileSource(t, "if (true) print 1;")

		assertBytes(t, compiled,
			byte(c.OP_TRUE),
			byte(c.OP_JUMP_IF_FALSE), 0, byte(skipThen),
			byte(c.OP_POP),
			byte(c.OP_CONSTANT), 0,
			byte(c.OP_PRINT),
			byte(c.OP_JUMP), 0, byte(popSize),
			byte(c.OP_POP),
			byte(c.OP_RETURN),
		)
	})

	t.Run("If anidado sin else: el salto externo incluye al interno completo", func(t *testing.T) {
		chunk := compileSource(t, "if (true) if (false) print 1;")
		innerIfSize := literalSize + jumpSize + skipThen + popSize
		skipOuterThen := popSize + innerIfSize + jumpSize

		assertBytes(t, chunk,
			byte(c.OP_TRUE),
			byte(c.OP_JUMP_IF_FALSE), 0, byte(skipOuterThen),
			byte(c.OP_POP),
			byte(c.OP_FALSE),
			byte(c.OP_JUMP_IF_FALSE), 0, byte(skipThen),
			byte(c.OP_POP),
			byte(c.OP_CONSTANT), 0,
			byte(c.OP_PRINT),
			byte(c.OP_JUMP), 0, byte(popSize),
			byte(c.OP_POP),
			byte(c.OP_JUMP), 0, byte(popSize),
			byte(c.OP_POP),
			byte(c.OP_RETURN),
		)
	})

	t.Run("Error si falta el paréntesis de apertura", func(t *testing.T) {
		err := compileSourceError(t, "if true) print 1;")
		utils.AssertErrorMessage(t, err, "'(' after if")
	})

	t.Run("Error si falta el paréntesis de cierre", func(t *testing.T) {
		err := compileSourceError(t, "if (true print 1;")
		utils.AssertErrorMessage(t, err, "')' after condition")
	})

	t.Run("Error si falta el cuerpo del if", func(t *testing.T) {
		err := compileSourceError(t, "if (true)")
		utils.AssertError(t, err)
		utils.AssertErrorMessage(t, err, "expected expression")
	})
}

func TestCompileIfWithElse(t *testing.T) {
	t.Run("If con else", func(t *testing.T) {
		compiled := compileSource(t, "if (true) print 1; else print 2;")

		assertBytes(t, compiled,
			byte(c.OP_TRUE),
			byte(c.OP_JUMP_IF_FALSE), 0, byte(skipThen), // si es falso, va al else
			byte(c.OP_POP),
			byte(c.OP_CONSTANT), 0,
			byte(c.OP_PRINT),
			byte(c.OP_JUMP), 0, byte(skipElse), // si se ejecutó el then, saltea el else
			byte(c.OP_POP),
			byte(c.OP_CONSTANT), 1,
			byte(c.OP_PRINT),
			byte(c.OP_RETURN),
		)
	})

	t.Run("Else anidado pertenece al if más cercano", func(t *testing.T) {
		compiled := compileSource(t, "if (true) if (false) print 1; else print 2;")
		innerIfSize := literalSize + jumpSize + skipThen + skipElse
		skipOuterThen := popSize + innerIfSize + jumpSize

		assertBytes(t, compiled,
			byte(c.OP_TRUE),
			byte(c.OP_JUMP_IF_FALSE), 0, byte(skipOuterThen), // externo: saltea todo el if interno
			byte(c.OP_POP),
			byte(c.OP_FALSE),
			byte(c.OP_JUMP_IF_FALSE), 0, byte(skipThen), // interno: va a su else
			byte(c.OP_POP),
			byte(c.OP_CONSTANT), 0,
			byte(c.OP_PRINT),
			byte(c.OP_JUMP), 0, byte(skipElse), // interno: saltea su else
			byte(c.OP_POP),
			byte(c.OP_CONSTANT), 1,
			byte(c.OP_PRINT),
			byte(c.OP_JUMP), 0, byte(popSize), // externo: no tiene else, solo saltea su POP
			byte(c.OP_POP),
			byte(c.OP_RETURN),
		)
	})

	t.Run("Else if encadenado: el else contiene a otro if", func(t *testing.T) {
		compiled := compileSource(t, "if (true) print 1; else if (false) print 2;")
		// el else del if externo contiene al if interno completo (sin else)
		innerIfSize := literalSize + jumpSize + skipThen + popSize
		skipElseIf := popSize + innerIfSize

		assertBytes(t, compiled,
			byte(c.OP_TRUE),
			byte(c.OP_JUMP_IF_FALSE), 0, byte(skipThen), // externo: si es falso, va al else
			byte(c.OP_POP),
			byte(c.OP_CONSTANT), 0,
			byte(c.OP_PRINT),
			byte(c.OP_JUMP), 0, byte(skipElseIf), // externo: saltea el else, que es el if interno
			byte(c.OP_POP),
			byte(c.OP_FALSE),
			byte(c.OP_JUMP_IF_FALSE), 0, byte(skipThen), // interno: saltea su then
			byte(c.OP_POP),
			byte(c.OP_CONSTANT), 1,
			byte(c.OP_PRINT),
			byte(c.OP_JUMP), 0, byte(popSize), // interno: saltea su POP
			byte(c.OP_POP),
			byte(c.OP_RETURN),
		)
	})

	t.Run("Error si falta el cuerpo del else", func(t *testing.T) {
		err := compileSourceError(t, "if (true) print 1; else")
		utils.AssertErrorMessage(t, err, "expected expression")
	})
}

func TestCompileAnd(t *testing.T) {
	t.Run("And se compila como un salto sobre el operando derecho", func(t *testing.T) {
		compiled := compileSource(t, "true and false;")
		skipRight := popSize + literalSize

		assertBytes(t, compiled,
			byte(c.OP_TRUE),
			byte(c.OP_JUMP_IF_FALSE), 0, byte(skipRight),
			byte(c.OP_POP),
			byte(c.OP_FALSE),
			byte(c.OP_POP),
			byte(c.OP_RETURN),
		)
	})

	t.Run("Error si falta el operando derecho", func(t *testing.T) {
		err := compileSourceError(t, "true and;")
		utils.AssertError(t, err)
	})

	t.Run("Error si falta el operando izquierdo", func(t *testing.T) {
		err := compileSourceError(t, "and true;")
		utils.AssertError(t, err)
	})
}

func TestCompileOr(t *testing.T) {
	t.Run("Or se compila como dos saltos: si el izquierdo es verdadero, saltea el derecho", func(t *testing.T) {
		compiled := compileSource(t, "true or false;")
		skipRight := popSize + literalSize

		assertBytes(t, compiled,
			byte(c.OP_TRUE),
			byte(c.OP_JUMP_IF_FALSE), 0, byte(jumpSize),
			byte(c.OP_JUMP), 0, byte(skipRight),
			byte(c.OP_POP),
			byte(c.OP_FALSE),
			byte(c.OP_POP),
			byte(c.OP_RETURN),
		)
	})

	t.Run("Error si falta el operando derecho", func(t *testing.T) {
		err := compileSourceError(t, "true or;")
		utils.AssertError(t, err)
	})

	t.Run("Error si falta el operando izquierdo", func(t *testing.T) {
		err := compileSourceError(t, "or ture;")
		utils.AssertError(t, err)
	})
}

func TestCompileWhile(t *testing.T) {
	t.Run("While: el cuerpo vuelve a la condición", func(t *testing.T) {
		compiled := compileSource(t, "while (false) print 1;")
		skipBody := popSize + constantSize + printSize + loopSize // el POP, el cuerpo y el LOOP
		backToCondition := literalSize + jumpSize + skipBody      // todo, desde la condición hasta el final del LOOP

		assertBytes(t, compiled,
			byte(c.OP_FALSE),
			byte(c.OP_JUMP_IF_FALSE), 0, byte(skipBody), // si es falsa, sale del ciclo
			byte(c.OP_POP),
			byte(c.OP_CONSTANT), 0,
			byte(c.OP_PRINT),
			byte(c.OP_LOOP), 0, byte(backToCondition), // vuelve a evaluar la condición
			byte(c.OP_POP), // salida: saca la condición
			byte(c.OP_RETURN),
		)
	})

	t.Run("Error si falta el paréntesis de apertura", func(t *testing.T) {
		err := compileSourceError(t, "while true) print 1;")
		utils.AssertErrorMessage(t, err, "'(' after 'while'")
	})

	t.Run("Error si falta el paréntesis de cierre", func(t *testing.T) {
		err := compileSourceError(t, "while (true print 1;")
		utils.AssertErrorMessage(t, err, "')' after condition")
	})
}

func TestCompileFor(t *testing.T) {
	t.Run("Sin nada: ciclo infinito, sin salto de salida", func(t *testing.T) {
		compiled := compileSource(t, "for (;;) print 1;")
		backToStart := constantSize + printSize + loopSize

		assertBytes(t, compiled,
			byte(c.OP_CONSTANT), 0,
			byte(c.OP_PRINT),
			byte(c.OP_LOOP), 0, byte(backToStart),
			byte(c.OP_RETURN),
		)
	})

	t.Run("Solo con condición: igual que un while", func(t *testing.T) {
		compiled := compileSource(t, "for (; false;) print 1;")
		skipBody := popSize + constantSize + printSize + loopSize
		backToCondition := literalSize + jumpSize + skipBody

		assertBytes(t, compiled,
			byte(c.OP_FALSE),
			byte(c.OP_JUMP_IF_FALSE), 0, byte(skipBody),
			byte(c.OP_POP),
			byte(c.OP_CONSTANT), 0,
			byte(c.OP_PRINT),
			byte(c.OP_LOOP), 0, byte(backToCondition),
			byte(c.OP_POP),
			byte(c.OP_RETURN),
		)
	})

	t.Run("Con inicializador: la variable es local al for", func(t *testing.T) {
		compiled := compileSource(t, "for (var i = 0;;) print i;")
		backToStart := getLocalSize + printSize + loopSize

		assertBytes(t, compiled,
			byte(c.OP_CONSTANT), 0, // var i = 0: se ejecuta una sola vez, antes del ciclo
			byte(c.OP_GET_LOCAL), 1,
			byte(c.OP_PRINT),
			byte(c.OP_LOOP), 0, byte(backToStart),
			byte(c.OP_POP), // endScope: saca i
			byte(c.OP_RETURN),
		)
	})

	t.Run("Con incremento: el cuerpo vuelve al incremento y el incremento a la condición", func(t *testing.T) {
		compiled := compileSource(t, "for (; false; 2) print 1;")
		incrementSize := constantSize + popSize
		bodySize := constantSize + printSize
		skipToExit := popSize + jumpSize + incrementSize + loopSize + bodySize + loopSize
		skipIncrement := incrementSize + loopSize
		backToCondition := literalSize + jumpSize + popSize + jumpSize + incrementSize + loopSize
		backToIncrement := incrementSize + loopSize + bodySize + loopSize

		assertBytes(t, compiled,
			byte(c.OP_FALSE),                              // condición: inicio del ciclo
			byte(c.OP_JUMP_IF_FALSE), 0, byte(skipToExit), // si es falsa, sale del ciclo
			byte(c.OP_POP),
			byte(c.OP_JUMP), 0, byte(skipIncrement), // saltea el incremento y va al cuerpo
			byte(c.OP_CONSTANT), 0, // incremento
			byte(c.OP_POP),
			byte(c.OP_LOOP), 0, byte(backToCondition), // después del incremento, vuelve a la condición
			byte(c.OP_CONSTANT), 1, // cuerpo
			byte(c.OP_PRINT),
			byte(c.OP_LOOP), 0, byte(backToIncrement), // después del cuerpo, vuelve al incremento
			byte(c.OP_POP), // salida: saca la condición
			byte(c.OP_RETURN),
		)
	})

	t.Run("Error si falta el paréntesis de apertura", func(t *testing.T) {
		err := compileSourceError(t, "for ;;) print 1;")
		utils.AssertErrorMessage(t, err, "'(' after 'for'")
	})

	t.Run("Error si falta el ';' después de la condición", func(t *testing.T) {
		err := compileSourceError(t, "for (; true) print 1;")
		utils.AssertErrorMessage(t, err, "';' after loop condition")
	})

	t.Run("Error si falta el paréntesis de cierre", func(t *testing.T) {
		err := compileSourceError(t, "for (;; 2 print 1;")
		utils.AssertErrorMessage(t, err, "')' after for clauses")
	})
}
