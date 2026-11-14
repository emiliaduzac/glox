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
	assertBytes(t, chunk, slices.Concat(expected, []byte{byte(c.OP_PRINT), byte(c.OP_NIL), byte(c.OP_RETURN)})...)
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
		assertBytes(t, &function.Chunk, byte(c.OP_CONSTANT), 0, byte(c.OP_PRINT), byte(c.OP_NIL), byte(c.OP_RETURN))
	})
}

func TestCompileFunctionDeclaration(t *testing.T) {
	t.Run("Una función global se guarda como constante y se define como variable global", func(t *testing.T) {
		chunk := compileSource(t, "fun f() {}")

		assertBytes(t, chunk, byte(c.OP_CONSTANT), 1, byte(c.OP_DEFINE_GLOBAL), 0, byte(c.OP_NIL), byte(c.OP_RETURN))
		utils.AssertValue(t, chunk.Constants[0], "f")
		function := chunk.Constants[1].(*c.Function)
		utils.AssertValue(t, function.Arity, 0)
		assertBytes(t, &function.Chunk, byte(c.OP_NIL), byte(c.OP_RETURN))
	})

	t.Run("El cuerpo se compila en el chunk de la función y no en el del script", func(t *testing.T) {
		chunk := compileSource(t, "fun f() { print 1; }")

		function := chunk.Constants[1].(*c.Function)
		assertBytes(t, &function.Chunk, byte(c.OP_CONSTANT), 0, byte(c.OP_PRINT), byte(c.OP_NIL), byte(c.OP_RETURN))
		utils.AssertValue(t, len(function.Chunk.Constants), 1)
		utils.AssertValue(t, function.Chunk.Constants[0], 1.0)
		utils.AssertValue(t, len(chunk.Constants), 2)
	})

	t.Run("Dentro de un bloque la función es una variable local", func(t *testing.T) {
		chunk := compileSource(t, "{ fun f() {} }")

		assertBytes(t, chunk, byte(c.OP_CONSTANT), 0, byte(c.OP_POP), byte(c.OP_NIL), byte(c.OP_RETURN))
		_, isFunction := chunk.Constants[0].(*c.Function)
		utils.AssertValue(t, isFunction, true)
	})

	t.Run("Una función declarada dentro de otra queda en el chunk de la de afuera", func(t *testing.T) {
		chunk := compileSource(t, "fun a() { fun b() {} }")

		outer := chunk.Constants[1].(*c.Function)
		assertBytes(t, &outer.Chunk, byte(c.OP_CONSTANT), 0, byte(c.OP_NIL), byte(c.OP_RETURN))
		_, isFunction := outer.Chunk.Constants[0].(*c.Function)
		utils.AssertValue(t, isFunction, true)
	})

	t.Run("Las variables del cuerpo son locales de la función y no se sacan con OP_POP al terminar", func(t *testing.T) {
		chunk := compileSource(t, "fun f() { var x = 1; print x; }")

		function := chunk.Constants[1].(*c.Function)
		assertBytes(t, &function.Chunk, byte(c.OP_CONSTANT), 0, byte(c.OP_GET_LOCAL), 1, byte(c.OP_PRINT), byte(c.OP_NIL), byte(c.OP_RETURN))
	})

	t.Run("Al final del cuerpo de una función se emite OP_NIL antes de OP_RETURN", func(t *testing.T) {
		chunk := compileSource(t, "fun f() {}")

		function := chunk.Constants[1].(*c.Function)
		assertBytes(t, &function.Chunk, byte(c.OP_NIL), byte(c.OP_RETURN))
	})

	t.Run("Cada función tiene su propia tabla de constantes", func(t *testing.T) {
		chunk := compileSource(t, "var a; fun f() { print a; }")

		assertBytes(t, chunk,
			byte(c.OP_NIL),
			byte(c.OP_DEFINE_GLOBAL), 0,
			byte(c.OP_CONSTANT), 2,
			byte(c.OP_DEFINE_GLOBAL), 1,
			byte(c.OP_NIL),
			byte(c.OP_RETURN),
		)
		function := chunk.Constants[2].(*c.Function)
		assertBytes(t, &function.Chunk, byte(c.OP_GET_GLOBAL), 0, byte(c.OP_PRINT), byte(c.OP_NIL), byte(c.OP_RETURN))
		utils.AssertValue(t, function.Chunk.Constants[0], "a")
	})

	t.Run("Errores de sintaxis en la declaración", func(t *testing.T) {
		tests := []struct {
			name     string
			source   string
			expected string
		}{
			{"falta el nombre", "fun () {}", "function name"},
			{"falta el paréntesis que abre", "fun f {}", "'(' after function name"},
			{"falta el paréntesis que cierra", "fun f(a {}", "')' after parameters"},
			{"falta la llave que abre", "fun f() print 1;", "'{' before function body"},
		}
		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				err := compileSourceError(t, test.source)

				utils.AssertErrorMessage(t, err, test.expected)
			})
		}
	})

	t.Run("Un error dentro del cuerpo se suma a los demás y el compilador sigue", func(t *testing.T) {
		errs := compileErrors(t, "fun f() { print 1 +; } print 2 +;")

		utils.AssertValue(t, len(errs), 2)
	})

	t.Run("Una función no puede ser el cuerpo de un if sin bloque", func(t *testing.T) {
		err := compileSourceError(t, "if (true) fun f() {}")

		utils.AssertError(t, err)
	})
}

func TestCompileFunctionParameters(t *testing.T) {
	paramsSource := func(n int) string {
		names := make([]string, n)
		for i := range names {
			names[i] = fmt.Sprintf("a%d", i)
		}
		return "fun f(" + strings.Join(names, ", ") + ") {}"
	}

	arityTests := []struct {
		name   string
		source string
		arity  int
	}{
		{"sin parámetros", "fun f() {}", 0},
		{"un parámetro", "fun f(a) {}", 1},
		{"varios parámetros", "fun f(a, b) {}", 2},
	}
	for _, test := range arityTests {
		t.Run("La aridad de la función es igual a la cantidad de parámetros: "+test.name, func(t *testing.T) {
			chunk := compileSource(t, test.source)

			function := chunk.Constants[1].(*c.Function)
			utils.AssertValue(t, function.Arity, test.arity)
		})
	}

	t.Run("Los parámetros son variables locales de la función a continuación del slot cero", func(t *testing.T) {
		chunk := compileSource(t, "fun f(a, b) { print b; print a; }")

		function := chunk.Constants[1].(*c.Function)
		assertBytes(t, &function.Chunk,
			byte(c.OP_GET_LOCAL), 2,
			byte(c.OP_PRINT),
			byte(c.OP_GET_LOCAL), 1,
			byte(c.OP_PRINT),
			byte(c.OP_NIL),
			byte(c.OP_RETURN),
		)
	})

	t.Run("Una función admite hasta 255 parámetros", func(t *testing.T) {
		chunk := compileSource(t, paramsSource(255))

		function := chunk.Constants[1].(*c.Function)
		utils.AssertValue(t, function.Arity, 255)
	})

	t.Run("Error si una función tiene más de 255 parámetros", func(t *testing.T) {
		err := compileSourceError(t, paramsSource(256))

		utils.AssertErrorMessage(t, err, "cannot have more than 255 parameters")
	})

	t.Run("Errores de sintaxis en la lista de parámetros", func(t *testing.T) {
		tests := []struct {
			name     string
			source   string
			expected string
		}{
			{"falta el nombre del parámetro", "fun f(1) {}", "parameter name"},
			{"coma al final de la lista", "fun f(a,) {}", "parameter name"},
			{"falta la coma entre parámetros", "fun f(a b) {}", "')' after parameters"},
		}
		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				err := compileSourceError(t, test.source)

				utils.AssertErrorMessage(t, err, test.expected)
			})
		}
	})
}

func TestCompileFunctionName(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		index    int
		expected string
	}{
		{"función global", "fun suma() {}", 1, "suma"},
		{"función local de un bloque", "{ fun resta() {} }", 0, "resta"},
		{"función con parámetros", "fun multiplicar(a, b) {}", 1, "multiplicar"},
	}
	for _, test := range tests {
		t.Run("La función se llama como el identificador de su declaración: "+test.name, func(t *testing.T) {
			chunk := compileSource(t, test.source)

			function := chunk.Constants[test.index].(*c.Function)
			utils.AssertValue(t, function.Name, test.expected)
			utils.AssertValue(t, utils.Stringify(function), "<fn "+test.expected+">")
		})
	}

	t.Run("Cada función anidada tiene su propio nombre", func(t *testing.T) {
		chunk := compileSource(t, "fun outer() { fun inner() {} }")

		outer := chunk.Constants[1].(*c.Function)
		inner := outer.Chunk.Constants[0].(*c.Function)
		utils.AssertValue(t, outer.Name, "outer")
		utils.AssertValue(t, inner.Name, "inner")
	})
}

func TestCompileReturn(t *testing.T) {
	t.Run("Un return con valor compila la expresión y emite OP_RETURN", func(t *testing.T) {
		chunk := compileSource(t, "fun f() { return 5; }")

		function := chunk.Constants[1].(*c.Function)
		assertBytes(t, &function.Chunk,
			byte(c.OP_CONSTANT), 0,
			byte(c.OP_RETURN),
			byte(c.OP_NIL), byte(c.OP_RETURN),
		)
	})

	t.Run("Un return sin valor equivale a return nil", func(t *testing.T) {
		chunk := compileSource(t, "fun f() { return; }")

		function := chunk.Constants[1].(*c.Function)
		assertBytes(t, &function.Chunk,
			byte(c.OP_NIL), byte(c.OP_RETURN),
			byte(c.OP_NIL), byte(c.OP_RETURN),
		)
	})

	t.Run("El valor del return puede ser una expresión con los parámetros", func(t *testing.T) {
		chunk := compileSource(t, "fun square(a) { return a * a; }")

		function := chunk.Constants[1].(*c.Function)
		assertBytes(t, &function.Chunk,
			byte(c.OP_GET_LOCAL), 1,
			byte(c.OP_GET_LOCAL), 1,
			byte(c.OP_MULTIPLY),
			byte(c.OP_RETURN),
			byte(c.OP_NIL), byte(c.OP_RETURN),
		)
	})

	t.Run("Error si hay un return en el nivel superior", func(t *testing.T) {
		err := compileSourceError(t, "return 1;")

		utils.AssertErrorMessage(t, err, "can't return from top-level code")
	})

	t.Run("Error si falta el punto y coma después del valor del return", func(t *testing.T) {
		err := compileSourceError(t, "fun f() { return 1 }")

		utils.AssertErrorMessage(t, err, "';' after return value")
	})
}

func TestCompileCall(t *testing.T) {
	argsSource := func(n int) string {
		return "f(" + strings.Repeat("nil, ", n-1) + "nil);"
	}

	t.Run("Una llamada sin argumentos carga la función y emite OP_CALL con cero argumentos", func(t *testing.T) {
		chunk := compileSource(t, "f();")

		assertBytes(t, chunk, byte(c.OP_GET_GLOBAL), 0, byte(c.OP_CALL), 0, byte(c.OP_POP), byte(c.OP_NIL), byte(c.OP_RETURN))
	})

	t.Run("OP_CALL lleva la cantidad de argumentos que se cargan en orden después de la función", func(t *testing.T) {
		chunk := compileSource(t, "f(1, 2);")

		assertBytes(t, chunk,
			byte(c.OP_GET_GLOBAL), 0,
			byte(c.OP_CONSTANT), 1,
			byte(c.OP_CONSTANT), 2,
			byte(c.OP_CALL), 2,
			byte(c.OP_POP),
			byte(c.OP_NIL),
			byte(c.OP_RETURN),
		)
		utils.AssertValue(t, chunk.Constants[1], 1.0)
		utils.AssertValue(t, chunk.Constants[2], 2.0)
	})

	t.Run("Cada argumento es una expresión completa y la coma los separa", func(t *testing.T) {
		chunk := compileSource(t, "f(1 + 2, 3 * 4);")

		assertBytes(t, chunk,
			byte(c.OP_GET_GLOBAL), 0,
			byte(c.OP_CONSTANT), 1,
			byte(c.OP_CONSTANT), 2,
			byte(c.OP_ADD),
			byte(c.OP_CONSTANT), 3,
			byte(c.OP_CONSTANT), 4,
			byte(c.OP_MULTIPLY),
			byte(c.OP_CALL), 2,
			byte(c.OP_POP),
			byte(c.OP_NIL),
			byte(c.OP_RETURN),
		)
	})

	t.Run("La llamada tiene más precedencia que el operador unario", func(t *testing.T) {
		chunk := compileSource(t, "print -f();")

		assertBytes(t, chunk,
			byte(c.OP_GET_GLOBAL), 0,
			byte(c.OP_CALL), 0,
			byte(c.OP_NEGATE),
			byte(c.OP_PRINT),
			byte(c.OP_NIL),
			byte(c.OP_RETURN),
		)
	})

	t.Run("Se puede llamar al resultado de otra llamada", func(t *testing.T) {
		chunk := compileSource(t, "f()();")

		assertBytes(t, chunk,
			byte(c.OP_GET_GLOBAL), 0,
			byte(c.OP_CALL), 0,
			byte(c.OP_CALL), 0,
			byte(c.OP_POP),
			byte(c.OP_NIL),
			byte(c.OP_RETURN),
		)
	})

	t.Run("Una llamada admite hasta 255 argumentos", func(t *testing.T) {
		chunk := compileSource(t, argsSource(255))

		expected := []byte{byte(c.OP_GET_GLOBAL), 0}
		expected = append(expected, slices.Repeat([]byte{byte(c.OP_NIL)}, 255)...)
		expected = append(expected, byte(c.OP_CALL), 255, byte(c.OP_POP), byte(c.OP_NIL), byte(c.OP_RETURN))
		assertBytes(t, chunk, expected...)
	})

	t.Run("Error si una llamada tiene más de 255 argumentos", func(t *testing.T) {
		err := compileSourceError(t, argsSource(256))
		utils.AssertErrorMessage(t, err, "cannot have more than 255 arguments")
	})

	t.Run("Error si falta el paréntesis que cierra la lista de argumentos", func(t *testing.T) {
		err := compileSourceError(t, "f(1;")
		utils.AssertErrorMessage(t, err, "')' after arguments")
	})

	t.Run("Error si hay una coma al final de la lista de argumentos", func(t *testing.T) {
		err := compileSourceError(t, "f(1,);")
		utils.AssertErrorMessage(t, err, "expected expression")
	})
}

func TestCompileEnclosingFunction(t *testing.T) {
	t.Run("Al terminar una función anidada el compilador sigue escribiendo en la de afuera", func(t *testing.T) {
		chunk := compileSource(t, "fun a() { fun b() {} print 1; }")

		outer := chunk.Constants[1].(*c.Function)
		assertBytes(t, &outer.Chunk,
			byte(c.OP_CONSTANT), 0,
			byte(c.OP_CONSTANT), 1,
			byte(c.OP_PRINT),
			byte(c.OP_NIL),
			byte(c.OP_RETURN),
		)
		inner := outer.Chunk.Constants[0].(*c.Function)
		assertBytes(t, &inner.Chunk, byte(c.OP_NIL), byte(c.OP_RETURN))
	})

	t.Run("Al terminar una función el compilador sigue escribiendo en el script", func(t *testing.T) {
		chunk := compileSource(t, "fun a() {} print 2;")

		assertBytes(t, chunk,
			byte(c.OP_CONSTANT), 1,
			byte(c.OP_DEFINE_GLOBAL), 0,
			byte(c.OP_CONSTANT), 2,
			byte(c.OP_PRINT),
			byte(c.OP_NIL),
			byte(c.OP_RETURN),
		)
		function := chunk.Constants[1].(*c.Function)
		assertBytes(t, &function.Chunk, byte(c.OP_NIL), byte(c.OP_RETURN))
	})

	t.Run("Las funciones se pueden anidar en más de un nivel", func(t *testing.T) {
		chunk := compileSource(t, "fun a() { fun b() { fun c() {} } }")

		a := chunk.Constants[1].(*c.Function)
		assertBytes(t, &a.Chunk, byte(c.OP_CONSTANT), 0, byte(c.OP_NIL), byte(c.OP_RETURN))
		b := a.Chunk.Constants[0].(*c.Function)
		assertBytes(t, &b.Chunk, byte(c.OP_CONSTANT), 0, byte(c.OP_NIL), byte(c.OP_RETURN))
		innermost := b.Chunk.Constants[0].(*c.Function)
		assertBytes(t, &innermost.Chunk, byte(c.OP_NIL), byte(c.OP_RETURN))
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
			byte(c.OP_NIL),
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
			byte(c.OP_NIL),
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
	t.Run("Un programa vacío solo tiene OP_NIL y OP_RETURN", func(t *testing.T) {
		chunk := compileSource(t, "")

		assertBytes(t, chunk, byte(c.OP_NIL), byte(c.OP_RETURN))
	})

	t.Run("Compila varias declaraciones en orden", func(t *testing.T) {
		chunk := compileSource(t, "print 1; print 2;")

		assertBytes(t, chunk,
			byte(c.OP_CONSTANT), 0,
			byte(c.OP_PRINT),
			byte(c.OP_CONSTANT), 1,
			byte(c.OP_PRINT),
			byte(c.OP_NIL),
			byte(c.OP_RETURN),
		)
	})

	t.Run("Error si falta el punto y coma después de un print", func(t *testing.T) {
		err := compileSourceError(t, "print 1")

		utils.AssertErrorMessage(t, err, "';' after statement")
	})

	t.Run("Una expresión como statement se descarta con OP_POP", func(t *testing.T) {
		chunk := compileSource(t, "1;")

		assertBytes(t, chunk, byte(c.OP_CONSTANT), 0, byte(c.OP_POP), byte(c.OP_NIL), byte(c.OP_RETURN))
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

		assertBytes(t, chunk, byte(c.OP_NIL), byte(c.OP_DEFINE_GLOBAL), 0, byte(c.OP_NIL), byte(c.OP_RETURN))
		utils.AssertValue(t, len(chunk.Constants), 1)
		utils.AssertValue(t, chunk.Constants[0], "a")
	})

	t.Run("Una variable global declarada con un valor guarda ese valor", func(t *testing.T) {
		chunk := compileSource(t, "var a = 1;")

		assertBytes(t, chunk, byte(c.OP_CONSTANT), 1, byte(c.OP_DEFINE_GLOBAL), 0, byte(c.OP_NIL), byte(c.OP_RETURN))
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

		assertBytes(t, chunk, byte(c.OP_GET_GLOBAL), 0, byte(c.OP_PRINT), byte(c.OP_NIL), byte(c.OP_RETURN))
		utils.AssertValue(t, len(chunk.Constants), 1)
		utils.AssertValue(t, chunk.Constants[0], "a")
	})
}

func TestCompileGlobalAssignment(t *testing.T) {
	t.Run("Asignar una variable global se compila a OP_SET_GLOBAL con su nombre", func(t *testing.T) {
		chunk := compileSource(t, "a = 1;")

		assertBytes(t, chunk, byte(c.OP_CONSTANT), 1, byte(c.OP_SET_GLOBAL), 0, byte(c.OP_POP), byte(c.OP_NIL), byte(c.OP_RETURN))
		utils.AssertValue(t, len(chunk.Constants), 2)
		utils.AssertValue(t, chunk.Constants[0], "a")
		utils.AssertValue(t, chunk.Constants[1], 1.0)
	})

	t.Run("Una asignación es una expresión y su valor se puede usar", func(t *testing.T) {
		chunk := compileSource(t, "print a = 2;")

		assertBytes(t, chunk, byte(c.OP_CONSTANT), 1, byte(c.OP_SET_GLOBAL), 0, byte(c.OP_PRINT), byte(c.OP_NIL), byte(c.OP_RETURN))
	})

	t.Run("La asignación asocia a derecha", func(t *testing.T) {
		chunk := compileSource(t, "a = b = 1;")

		assertBytes(t, chunk,
			byte(c.OP_CONSTANT), 2,
			byte(c.OP_SET_GLOBAL), 1,
			byte(c.OP_SET_GLOBAL), 0,
			byte(c.OP_POP),
			byte(c.OP_NIL),
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
			byte(c.OP_NIL),
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

		assertBytes(t, chunk, byte(c.OP_CONSTANT), 1, byte(c.OP_DEFINE_GLOBAL), 0, byte(c.OP_GET_GLOBAL), 0, byte(c.OP_PRINT), byte(c.OP_NIL), byte(c.OP_RETURN))
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

		assertBytes(t, chunk, byte(c.OP_CONSTANT), 0, byte(c.OP_POP), byte(c.OP_NIL), byte(c.OP_RETURN)) // el pop es por el cierre del scope
		utils.AssertValue(t, len(chunk.Constants), 1)
		utils.AssertValue(t, chunk.Constants[0], 1.0)
	})

	t.Run("Declarar una variable local sin valor es nil", func(t *testing.T) {
		chunk := compileSource(t, "{ var a; }")

		assertBytes(t, chunk, byte(c.OP_NIL), byte(c.OP_POP), byte(c.OP_NIL), byte(c.OP_RETURN))
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

		assertBytes(t, chunk, byte(c.OP_CONSTANT), 0, byte(c.OP_GET_LOCAL), 1, byte(c.OP_PRINT), byte(c.OP_POP), byte(c.OP_NIL), byte(c.OP_RETURN))
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
			byte(c.OP_NIL),
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
			byte(c.OP_NIL),
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
			byte(c.OP_NIL),
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
			byte(c.OP_NIL),
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
			byte(c.OP_NIL),
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
			byte(c.OP_NIL),
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
			byte(c.OP_NIL),
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
			byte(c.OP_NIL),
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
			byte(c.OP_NIL),
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
			byte(c.OP_NIL),
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
			byte(c.OP_NIL),
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
			byte(c.OP_NIL),
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
			byte(c.OP_NIL),
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
			byte(c.OP_NIL),
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
			byte(c.OP_NIL),
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
			byte(c.OP_NIL),
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
			byte(c.OP_NIL),
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
			byte(c.OP_NIL),
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
			byte(c.OP_NIL),
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

func TestCompileLoopsAreClosed(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{"while", "while (false) print 1;"},
		{"for", "for (var i = 0; i < 1; i = i + 1) print i;"},
		{"ciclos anidados", "while (false) for (;;) print 1;"},
		{"error en el cuerpo del ciclo", "while (false) print ;"},
		{"error en el for antes de apilar, dentro de un while", "while (false) { for x;;) print 1; }"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tokens, _ := scanner.NewScanner(test.source).Scan()
			compiler := NewCompiler(tokens)
			_, _ = compiler.Compile()

			utils.AssertValue(t, len(compiler.loops), 0)
		})
	}
}

func TestCompileContinue(t *testing.T) {
	t.Run("En un while, vuelve a la condición", func(t *testing.T) {
		compiled := compileSource(t, "while (true) continue;")
		continueToCondition := literalSize + jumpSize + popSize + loopSize
		backToCondition := continueToCondition + loopSize
		skipBody := popSize + loopSize + loopSize // el POP, el continue y el LOOP del final

		assertBytes(t, compiled,
			byte(c.OP_TRUE),
			byte(c.OP_JUMP_IF_FALSE), 0, byte(skipBody),
			byte(c.OP_POP),
			byte(c.OP_LOOP), 0, byte(continueToCondition), // continue
			byte(c.OP_LOOP), 0, byte(backToCondition), // fin del cuerpo
			byte(c.OP_POP),
			byte(c.OP_NIL),
			byte(c.OP_RETURN),
		)
	})

	t.Run("En un for, vuelve al incremento", func(t *testing.T) {
		compiled := compileSource(t, "for (; true; 2) continue;")
		incrementSize := constantSize + popSize
		skipToExit := popSize + jumpSize + incrementSize + loopSize + loopSize + loopSize
		skipIncrement := incrementSize + loopSize
		backToCondition := literalSize + jumpSize + popSize + jumpSize + incrementSize + loopSize
		continueToIncrement := incrementSize + loopSize + loopSize
		backToIncrement := continueToIncrement + loopSize

		assertBytes(t, compiled,
			byte(c.OP_TRUE), // condición: inicio del ciclo
			byte(c.OP_JUMP_IF_FALSE), 0, byte(skipToExit),
			byte(c.OP_POP),
			byte(c.OP_JUMP), 0, byte(skipIncrement),
			byte(c.OP_CONSTANT), 0, // incremento
			byte(c.OP_POP),
			byte(c.OP_LOOP), 0, byte(backToCondition),
			byte(c.OP_LOOP), 0, byte(continueToIncrement), // continue: vuelve al incremento
			byte(c.OP_LOOP), 0, byte(backToIncrement), // fin del cuerpo también vuelve al incremento
			byte(c.OP_POP),
			byte(c.OP_NIL),
			byte(c.OP_RETURN),
		)
	})

	t.Run("Saca del stack las locales del ciclo antes de saltar", func(t *testing.T) {
		compiled := compileSource(t, "while (true) { var a = 1; continue; }")
		continueToCondition := literalSize + jumpSize + popSize + constantSize + popSize + loopSize
		backToCondition := continueToCondition + popSize + loopSize
		skipBody := popSize + constantSize + popSize + loopSize + popSize + loopSize

		assertBytes(t, compiled,
			byte(c.OP_TRUE), // inicio del ciclo
			byte(c.OP_JUMP_IF_FALSE), 0, byte(skipBody),
			byte(c.OP_POP),
			byte(c.OP_CONSTANT), 0, // var a = 1
			byte(c.OP_POP),                                // continue: saca a
			byte(c.OP_LOOP), 0, byte(continueToCondition), // continue: vuelve
			byte(c.OP_POP), // endScope: saca a
			byte(c.OP_LOOP), 0, byte(backToCondition),
			byte(c.OP_POP),
			byte(c.OP_NIL),
			byte(c.OP_RETURN),
		)
	})

	t.Run("No saca las locales de afuera del ciclo", func(t *testing.T) {
		compiled := compileSource(t, "{ var a = 1; while (true) continue; }")
		continueToCondition := literalSize + jumpSize + popSize + loopSize
		backToCondition := continueToCondition + loopSize
		skipBody := popSize + loopSize + loopSize

		assertBytes(t, compiled,
			byte(c.OP_CONSTANT), 0, // var a = 1: antes del ciclo
			byte(c.OP_TRUE), // inicio del ciclo
			byte(c.OP_JUMP_IF_FALSE), 0, byte(skipBody),
			byte(c.OP_POP),
			byte(c.OP_LOOP), 0, byte(continueToCondition),
			byte(c.OP_LOOP), 0, byte(backToCondition),
			byte(c.OP_POP),
			byte(c.OP_POP), // endScope final saca a
			byte(c.OP_NIL),
			byte(c.OP_RETURN),
		)
	})

	t.Run("No saca la variable del inicializador del for", func(t *testing.T) {
		compiled := compileSource(t, "for (var i = 0; true;) continue;")
		continueToCondition := literalSize + jumpSize + popSize + loopSize
		backToCondition := continueToCondition + loopSize
		skipBody := popSize + loopSize + loopSize

		assertBytes(t, compiled,
			byte(c.OP_CONSTANT), 0, // var i = 0
			byte(c.OP_TRUE), // inicio del ciclo
			byte(c.OP_JUMP_IF_FALSE), 0, byte(skipBody),
			byte(c.OP_POP),
			byte(c.OP_LOOP), 0, byte(continueToCondition), // continue: sin POP de i
			byte(c.OP_LOOP), 0, byte(backToCondition),
			byte(c.OP_POP),
			byte(c.OP_POP), // endScope del for: saca i
			byte(c.OP_NIL),
			byte(c.OP_RETURN),
		)
	})

	t.Run("En ciclos anidados, vuelve al más interno", func(t *testing.T) {
		compiled := compileSource(t, "while (true) while (false) continue;")
		continueToInner := literalSize + jumpSize + popSize + loopSize
		backToInner := continueToInner + loopSize
		skipInnerBody := popSize + loopSize + loopSize
		innerLoopSize := literalSize + jumpSize + skipInnerBody + popSize
		skipOuterBody := popSize + innerLoopSize + loopSize
		backToOuter := literalSize + jumpSize + skipOuterBody

		assertBytes(t, compiled,
			byte(c.OP_TRUE), // inicio del ciclo de afuera
			byte(c.OP_JUMP_IF_FALSE), 0, byte(skipOuterBody),
			byte(c.OP_POP),
			byte(c.OP_FALSE), // inicio del ciclo de adentro
			byte(c.OP_JUMP_IF_FALSE), 0, byte(skipInnerBody),
			byte(c.OP_POP),
			byte(c.OP_LOOP), 0, byte(continueToInner), // continue: vuelve al de adentro
			byte(c.OP_LOOP), 0, byte(backToInner),
			byte(c.OP_POP), // salida del de adentro
			byte(c.OP_LOOP), 0, byte(backToOuter),
			byte(c.OP_POP), // salida del de afuera
			byte(c.OP_NIL),
			byte(c.OP_RETURN),
		)
	})

	t.Run("Las locales siguen disponibles para el resto del cuerpo", func(t *testing.T) {
		compiled := compileSource(t, "while (true) { var a = 1; if (a) continue; print a; }")

		// si el continue hubiera sacado a de locals el print la compilaría como global
		if slices.Contains(compiled.Bytes, byte(c.OP_GET_GLOBAL)) {
			t.Errorf("expected a to be compiled as a local, got bytes %v", compiled.Bytes)
		}
	})

	t.Run("Error si está fuera de un ciclo", func(t *testing.T) {
		tests := []struct {
			name   string
			source string
		}{
			{"suelto", "continue;"},
			{"dentro de un if", "if (true) continue;"},
			{"después de que terminó el ciclo", "while (false) {} continue;"},
		}

		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				err := compileSourceError(t, test.source)
				utils.AssertErrorMessage(t, err, "found 'continue' outside a loop")
			})
		}
	})

	t.Run("Error si falta el ';'", func(t *testing.T) {
		err := compileSourceError(t, "while (true) continue")
		utils.AssertErrorMessage(t, err, "';' after continue")
	})
}

func TestCompileBreak(t *testing.T) {
	t.Run("En un while, salta después del POP de salida", func(t *testing.T) {
		compiled := compileSource(t, "while (true) break;")
		skipBody := popSize + jumpSize + loopSize
		backToCondition := literalSize + jumpSize + skipBody
		breakToEnd := loopSize + popSize

		assertBytes(t, compiled,
			byte(c.OP_TRUE),
			byte(c.OP_JUMP_IF_FALSE), 0, byte(skipBody),
			byte(c.OP_POP),
			byte(c.OP_JUMP), 0, byte(breakToEnd), // break
			byte(c.OP_LOOP), 0, byte(backToCondition),
			byte(c.OP_POP),
			byte(c.OP_NIL),
			byte(c.OP_RETURN),
		)
	})

	t.Run("En un for, salta después del POP de salida", func(t *testing.T) {
		compiled := compileSource(t, "for (; true; 2) break;")
		incrementSize := constantSize + popSize
		skipToExit := popSize + jumpSize + incrementSize + loopSize + jumpSize + loopSize
		skipIncrement := incrementSize + loopSize
		backToCondition := literalSize + jumpSize + popSize + jumpSize + incrementSize + loopSize
		breakToEnd := loopSize + popSize
		backToIncrement := incrementSize + loopSize + jumpSize + loopSize

		assertBytes(t, compiled,
			byte(c.OP_TRUE),
			byte(c.OP_JUMP_IF_FALSE), 0, byte(skipToExit),
			byte(c.OP_POP),
			byte(c.OP_JUMP), 0, byte(skipIncrement),
			byte(c.OP_CONSTANT), 0,
			byte(c.OP_POP),
			byte(c.OP_LOOP), 0, byte(backToCondition),
			byte(c.OP_JUMP), 0, byte(breakToEnd), // break: no pasa por el incremento
			byte(c.OP_LOOP), 0, byte(backToIncrement),
			byte(c.OP_POP),
			byte(c.OP_NIL),
			byte(c.OP_RETURN),
		)
	})

	t.Run("En un for sin condición, salta después del último LOOP", func(t *testing.T) {
		compiled := compileSource(t, "for (;;) break;")
		breakToEnd := loopSize
		backToStart := jumpSize + loopSize

		assertBytes(t, compiled,
			byte(c.OP_JUMP), 0, byte(breakToEnd), // break
			byte(c.OP_LOOP), 0, byte(backToStart),
			byte(c.OP_NIL),
			byte(c.OP_RETURN),
		)
	})

	t.Run("Saca del stack las locales del ciclo antes de saltar", func(t *testing.T) {
		compiled := compileSource(t, "while (true) { var a = 1; break; }")
		skipBody := popSize + constantSize + popSize + jumpSize + popSize + loopSize
		backToCondition := literalSize + jumpSize + skipBody
		breakToEnd := popSize + loopSize + popSize

		assertBytes(t, compiled,
			byte(c.OP_TRUE),
			byte(c.OP_JUMP_IF_FALSE), 0, byte(skipBody),
			byte(c.OP_POP),
			byte(c.OP_CONSTANT), 0,
			byte(c.OP_POP),                       // break: saca a
			byte(c.OP_JUMP), 0, byte(breakToEnd), // break: salta
			byte(c.OP_POP),
			byte(c.OP_LOOP), 0, byte(backToCondition),
			byte(c.OP_POP),
			byte(c.OP_NIL),
			byte(c.OP_RETURN),
		)
	})

	t.Run("No saca las locales de afuera del ciclo", func(t *testing.T) {
		compiled := compileSource(t, "{ var a = 1; while (true) break; }")
		skipBody := popSize + jumpSize + loopSize
		backToCondition := literalSize + jumpSize + skipBody
		breakToEnd := loopSize + popSize

		assertBytes(t, compiled,
			byte(c.OP_CONSTANT), 0,
			byte(c.OP_TRUE),
			byte(c.OP_JUMP_IF_FALSE), 0, byte(skipBody),
			byte(c.OP_POP),
			byte(c.OP_JUMP), 0, byte(breakToEnd), // break: sin POP de a
			byte(c.OP_LOOP), 0, byte(backToCondition),
			byte(c.OP_POP),
			byte(c.OP_POP),
			byte(c.OP_NIL),
			byte(c.OP_RETURN),
		)
	})

	t.Run("No saca la variable del inicializador del for", func(t *testing.T) {
		compiled := compileSource(t, "for (var i = 0; true;) break;")
		skipBody := popSize + jumpSize + loopSize
		backToCondition := literalSize + jumpSize + skipBody
		breakToEnd := loopSize + popSize

		assertBytes(t, compiled,
			byte(c.OP_CONSTANT), 0,
			byte(c.OP_TRUE),
			byte(c.OP_JUMP_IF_FALSE), 0, byte(skipBody),
			byte(c.OP_POP),
			byte(c.OP_JUMP), 0, byte(breakToEnd), // break: sin POP de i
			byte(c.OP_LOOP), 0, byte(backToCondition),
			byte(c.OP_POP),
			byte(c.OP_POP),
			byte(c.OP_NIL),
			byte(c.OP_RETURN),
		)
	})

	t.Run("En ciclos anidados, sale solo del más interno", func(t *testing.T) {
		compiled := compileSource(t, "while (true) while (false) break;")
		skipInnerBody := popSize + jumpSize + loopSize
		backToInner := literalSize + jumpSize + skipInnerBody
		breakToInnerEnd := loopSize + popSize
		innerLoopSize := literalSize + jumpSize + skipInnerBody + popSize
		skipOuterBody := popSize + innerLoopSize + loopSize
		backToOuter := literalSize + jumpSize + skipOuterBody

		assertBytes(t, compiled,
			byte(c.OP_TRUE),
			byte(c.OP_JUMP_IF_FALSE), 0, byte(skipOuterBody),
			byte(c.OP_POP),
			byte(c.OP_FALSE),
			byte(c.OP_JUMP_IF_FALSE), 0, byte(skipInnerBody),
			byte(c.OP_POP),
			byte(c.OP_JUMP), 0, byte(breakToInnerEnd), // break: cae en el LOOP del de afuera
			byte(c.OP_LOOP), 0, byte(backToInner),
			byte(c.OP_POP),
			byte(c.OP_LOOP), 0, byte(backToOuter),
			byte(c.OP_POP),
			byte(c.OP_NIL),
			byte(c.OP_RETURN),
		)
	})

	t.Run("Varios break en el mismo ciclo saltan al mismo lugar", func(t *testing.T) {
		compiled := compileSource(t, "while (true) { if (true) break; break; }")
		ifSkipThen := popSize + jumpSize + jumpSize
		ifSize := literalSize + jumpSize + ifSkipThen + popSize
		skipBody := popSize + ifSize + jumpSize + loopSize
		backToCondition := literalSize + jumpSize + skipBody
		firstBreakToEnd := jumpSize + popSize + jumpSize + loopSize + popSize
		secondBreakToEnd := loopSize + popSize

		assertBytes(t, compiled,
			byte(c.OP_TRUE),
			byte(c.OP_JUMP_IF_FALSE), 0, byte(skipBody),
			byte(c.OP_POP),
			byte(c.OP_TRUE),
			byte(c.OP_JUMP_IF_FALSE), 0, byte(ifSkipThen),
			byte(c.OP_POP),
			byte(c.OP_JUMP), 0, byte(firstBreakToEnd), // primer break
			byte(c.OP_JUMP), 0, byte(popSize),
			byte(c.OP_POP),
			byte(c.OP_JUMP), 0, byte(secondBreakToEnd), // segundo break
			byte(c.OP_LOOP), 0, byte(backToCondition),
			byte(c.OP_POP),
			byte(c.OP_NIL),
			byte(c.OP_RETURN),
		)
	})

	t.Run("Las locales siguen disponibles para el resto del cuerpo", func(t *testing.T) {
		compiled := compileSource(t, "while (true) { var a = 1; if (a) break; print a; }")

		if slices.Contains(compiled.Bytes, byte(c.OP_GET_GLOBAL)) {
			t.Errorf("expected a to be compiled as a local, got bytes %v", compiled.Bytes)
		}
	})

	t.Run("Error si está fuera de un ciclo", func(t *testing.T) {
		tests := []struct {
			name   string
			source string
		}{
			{"suelto", "break;"},
			{"dentro de un if", "if (true) break;"},
			{"después de que terminó el ciclo", "while (false) {} break;"},
		}

		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				err := compileSourceError(t, test.source)
				utils.AssertErrorMessage(t, err, "found 'break' outside a loop")
			})
		}
	})

	t.Run("Error si falta el ';'", func(t *testing.T) {
		err := compileSourceError(t, "while (true) break")
		utils.AssertErrorMessage(t, err, "';' after break")
	})
}
