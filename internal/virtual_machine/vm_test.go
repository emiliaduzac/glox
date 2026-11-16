package virtualmachine

import (
	"bytes"
	"math"
	"testing"
	c "tp-lox/internal/chunk"
	"tp-lox/internal/utils"
)

func TestOpReturn(t *testing.T) {
	t.Run("El script termina sin error cuando solo queda su valor de retorno", func(t *testing.T) {
		vm, chunk := setUpVM(t)
		chunk.Write(byte(c.OP_NIL), line)
		chunk.Write(byte(c.OP_RETURN), line)

		err := run(vm, chunk)

		utils.AssertNoError(t, err)
	})

	t.Run("Error si al terminar el script queda algo además del valor de retorno", func(t *testing.T) {
		vm, chunk := setUpVM(t)
		chunk.Write(byte(c.OP_NIL), line)
		chunk.Write(byte(c.OP_NIL), line)
		chunk.Write(byte(c.OP_RETURN), line)

		err := run(vm, chunk)

		utils.AssertErrorMessage(t, err, "inconsistent Stack Height")
	})

	t.Run("El return del script corta la ejecución", func(t *testing.T) {
		vm, chunk := setUpVM(t)
		chunk.Write(byte(c.OP_NIL), line)
		chunk.Write(byte(c.OP_RETURN), line)
		chunk.Write(byte(c.OP_NIL), line)

		err := run(vm, chunk)

		utils.AssertNoError(t, err)
		utils.AssertValue(t, len(stackValues(vm)), 0)
	})
}

func TestScriptCallFrame(t *testing.T) {
	t.Run("La función del script ocupa el slot cero del stack mientras se ejecuta", func(t *testing.T) {
		vm, chunk := setUpVM(t)
		writeConstant(chunk, 1.0)
		function := c.NewFunction()
		function.Chunk = *chunk

		err := vm.Run(function)

		utils.AssertNoError(t, err)
		utils.AssertValue(t, vm.stack[0].(*closure).function, function)
		utils.AssertValue(t, vm.stack[1], 1.0)
	})

	t.Run("Al terminar el script no quedan ni valores ni frames", func(t *testing.T) {
		vm, chunk := setUpVM(t)
		chunk.Write(byte(c.OP_NIL), line)
		chunk.Write(byte(c.OP_RETURN), line)

		err := run(vm, chunk)

		utils.AssertNoError(t, err)
		utils.AssertValue(t, len(vm.stack), 0)
		utils.AssertValue(t, len(vm.frames), 0)
	})

	t.Run("Después de un error la VM se puede volver a usar", func(t *testing.T) {
		vm := NewVirtualMachine()
		var out bytes.Buffer
		vm.out = &out

		failing := c.NewChunk()
		name := failing.AddConstant("a")
		failing.Write(byte(c.OP_GET_GLOBAL), line)
		failing.Write(byte(name), line)
		failing.Write(byte(c.OP_NIL), line)
		failing.Write(byte(c.OP_RETURN), line)

		working := c.NewChunk()
		writeConstant(working, 1.0)
		working.Write(byte(c.OP_PRINT), line)
		working.Write(byte(c.OP_NIL), line)
		working.Write(byte(c.OP_RETURN), line)

		utils.AssertError(t, run(vm, failing))
		utils.AssertValue(t, len(vm.frames), 0)
		utils.AssertNoError(t, run(vm, working))
		utils.AssertValue(t, out.String(), "1\n")
	})
}

func TestOpCall(t *testing.T) {
	newFunction := func(arity int, body func(chunk *c.Chunk)) *c.Function {
		function := c.NewFunction()
		function.Arity = arity
		body(&function.Chunk)
		function.Chunk.Write(byte(c.OP_NIL), line)
		function.Chunk.Write(byte(c.OP_RETURN), line)
		return function
	}

	t.Run("Llamar a una función ejecuta el código de su chunk", func(t *testing.T) {
		vm, chunk := setUpVM(t)
		var out bytes.Buffer
		vm.out = &out
		callee := newFunction(0, func(body *c.Chunk) {
			writeConstant(body, "hola")
			body.Write(byte(c.OP_PRINT), line)
		})
		writeClosure(chunk, callee)
		writeCall(chunk, 0)

		err := run(vm, chunk)

		utils.AssertNoError(t, err)
		utils.AssertValue(t, out.String(), "hola\n")
	})

	t.Run("El argumento queda atado al parámetro de la función", func(t *testing.T) {
		vm, chunk := setUpVM(t)
		var out bytes.Buffer
		vm.out = &out
		callee := newFunction(1, func(body *c.Chunk) {
			body.Write(byte(c.OP_GET_LOCAL), line)
			body.Write(1, line)
			body.Write(byte(c.OP_PRINT), line)
		})
		writeClosure(chunk, callee)
		writeConstant(chunk, 4.0)
		writeCall(chunk, 1)

		err := run(vm, chunk)

		utils.AssertNoError(t, err)
		utils.AssertValue(t, out.String(), "4\n")
	})

	t.Run("Cada argumento queda atado al parámetro de su posición", func(t *testing.T) {
		vm, chunk := setUpVM(t)
		var out bytes.Buffer
		vm.out = &out
		callee := newFunction(2, func(body *c.Chunk) {
			body.Write(byte(c.OP_GET_LOCAL), line)
			body.Write(2, line)
			body.Write(byte(c.OP_PRINT), line)
			body.Write(byte(c.OP_GET_LOCAL), line)
			body.Write(1, line)
			body.Write(byte(c.OP_PRINT), line)
		})
		writeClosure(chunk, callee)
		writeConstant(chunk, 1.0)
		writeConstant(chunk, 2.0)
		writeCall(chunk, 2)

		err := run(vm, chunk)

		utils.AssertNoError(t, err)
		utils.AssertValue(t, out.String(), "2\n1\n")
	})

	tests := []struct {
		name  string
		value any
	}{
		{"número", 1.0},
		{"string", "hola"},
		{"nil", nil},
	}
	for _, test := range tests {
		t.Run("Error si se llama a un valor que no es una función: "+test.name, func(t *testing.T) {
			vm, chunk := setUpVM(t)
			writeConstant(chunk, test.value)
			writeCall(chunk, 0)

			err := run(vm, chunk)

			utils.AssertErrorMessage(t, err, "can only call functions and classes")
		})
	}

	t.Run("Error si se llama a una función con menos argumentos de los esperados", func(t *testing.T) {
		vm, chunk := setUpVM(t)
		callee := newFunction(2, func(body *c.Chunk) {})
		writeClosure(chunk, callee)
		writeConstant(chunk, 1.0)
		writeCall(chunk, 1)

		err := run(vm, chunk)

		utils.AssertErrorMessage(t, err, "expected 2 arguments but got 1")
	})

	t.Run("Error si se llama a una función con más argumentos de los esperados", func(t *testing.T) {
		vm, chunk := setUpVM(t)
		callee := newFunction(1, func(body *c.Chunk) {})
		writeClosure(chunk, callee)
		writeConstant(chunk, 1.0)
		writeConstant(chunk, 2.0)
		writeCall(chunk, 2)

		err := run(vm, chunk)

		utils.AssertErrorMessage(t, err, "expected 1 arguments but got 2")
	})

	nestedCalls := func(depth int) *c.Function {
		callee := newFunction(0, func(body *c.Chunk) {})
		for i := 1; i < depth; i++ {
			next := callee
			callee = newFunction(0, func(body *c.Chunk) {
				writeClosure(body, next)
				writeCall(body, 0)
			})
		}
		return callee
	}

	t.Run("Se permiten llamadas anidadas hasta la profundidad máxima", func(t *testing.T) {
		vm, chunk := setUpVM(t)
		writeClosure(chunk, nestedCalls(framesMax-1))
		writeCall(chunk, 0)

		err := run(vm, chunk)

		utils.AssertNoError(t, err)
	})

	t.Run("Error de stack overflow si las llamadas anidadas superan la profundidad máxima", func(t *testing.T) {
		vm, chunk := setUpVM(t)
		writeClosure(chunk, nestedCalls(framesMax))
		writeCall(chunk, 0)

		err := run(vm, chunk)

		utils.AssertErrorMessage(t, err, "stack overflow")
	})
}

func TestNativeCall(t *testing.T) {
	subtract := &native{arity: 2, fn: func(args []any) any {
		return args[0].(float64) - args[1].(float64)
	}}
	callSubtract := func(chunk *c.Chunk) {
		writeConstant(chunk, subtract)
		writeConstant(chunk, 33.0)
		writeConstant(chunk, 3.0)
		writeCall(chunk, 2)
	}

	t.Run("Una función nativa recibe los argumentos en orden y devuelve su resultado", func(t *testing.T) {
		vm, chunk := setUpVM(t)
		var out bytes.Buffer
		vm.out = &out
		callSubtract(chunk)
		chunk.Write(byte(c.OP_PRINT), line)
		chunk.Write(byte(c.OP_NIL), line)
		chunk.Write(byte(c.OP_RETURN), line)

		err := run(vm, chunk)

		utils.AssertNoError(t, err)
		utils.AssertValue(t, out.String(), "30\n")
	})

	t.Run("Después de llamar a una función nativa queda solo su resultado en el stack", func(t *testing.T) {
		vm, chunk := setUpVM(t)
		callSubtract(chunk)

		err := run(vm, chunk)

		utils.AssertNoError(t, err)
		utils.AssertValue(t, len(stackValues(vm)), 1)
		utils.AssertValue(t, stackValues(vm)[0], 30.0)
	})

	t.Run("Una función nativa se puede guardar en una variable y llamarla desde ahí", func(t *testing.T) {
		vm, chunk := setUpVM(t)
		writeConstant(chunk, subtract)
		alias := chunk.AddConstant("alias")
		chunk.Write(byte(c.OP_DEFINE_GLOBAL), line)
		chunk.Write(byte(alias), line)
		writeGetGlobal(chunk, "alias")
		writeConstant(chunk, 33.0)
		writeConstant(chunk, 3.0)
		writeCall(chunk, 2)

		err := run(vm, chunk)

		utils.AssertNoError(t, err)
		utils.AssertValue(t, stackValues(vm)[0], 30.0)
	})

	t.Run("Error si se llama a una función nativa con menos argumentos de los esperados", func(t *testing.T) {
		vm, chunk := setUpVM(t)
		writeConstant(chunk, subtract)
		writeConstant(chunk, 33.0)
		writeCall(chunk, 1)

		err := run(vm, chunk)

		utils.AssertErrorMessage(t, err, "expected 2 arguments but got 1")
	})

	t.Run("Error si se llama a una función nativa con más argumentos de los esperados", func(t *testing.T) {
		vm, chunk := setUpVM(t)
		writeConstant(chunk, subtract)
		writeConstant(chunk, 33.0)
		writeConstant(chunk, 3.0)
		writeConstant(chunk, 1.0)
		writeCall(chunk, 3)

		err := run(vm, chunk)

		utils.AssertErrorMessage(t, err, "expected 2 arguments but got 3")
	})

	t.Run("Una función nativa se imprime como <native fn>", func(t *testing.T) {
		vm, chunk := setUpVM(t)
		var out bytes.Buffer
		vm.out = &out
		writeConstant(chunk, subtract)
		chunk.Write(byte(c.OP_PRINT), line)

		err := run(vm, chunk)

		utils.AssertNoError(t, err)
		utils.AssertValue(t, out.String(), "<native fn>\n")
	})
}

func TestClock(t *testing.T) {
	t.Run("Una VM nueva ya tiene definida la función nativa clock", func(t *testing.T) {
		vm, chunk := setUpVM(t)
		writeGetGlobal(chunk, "clock")
		writeCall(chunk, 0)

		err := run(vm, chunk)

		utils.AssertNoError(t, err)
		_, isNumber := stackValues(vm)[0].(float64)
		utils.AssertValue(t, isNumber, true)
	})
}

func TestReturn(t *testing.T) {
	returning := func(value any) *c.Function {
		function := c.NewFunction()
		writeConstant(&function.Chunk, value)
		function.Chunk.Write(byte(c.OP_RETURN), line)
		return function
	}

	t.Run("Quien llamó recibe el valor que dejó la función y sigue ejecutando", func(t *testing.T) {
		vm, chunk := setUpVM(t)
		var out bytes.Buffer
		vm.out = &out
		writeClosure(chunk, returning(5.0))
		writeCall(chunk, 0)
		chunk.Write(byte(c.OP_PRINT), line)
		writeConstant(chunk, "listo")
		chunk.Write(byte(c.OP_PRINT), line)

		err := run(vm, chunk)

		utils.AssertNoError(t, err)
		utils.AssertValue(t, out.String(), "5\nlisto\n")
	})

	t.Run("Al volver se descartan la función y sus argumentos, y queda solo el resultado", func(t *testing.T) {
		vm, chunk := setUpVM(t)
		suma := c.NewFunction()
		suma.Arity = 2
		suma.Chunk.Write(byte(c.OP_GET_LOCAL), line)
		suma.Chunk.Write(1, line)
		suma.Chunk.Write(byte(c.OP_GET_LOCAL), line)
		suma.Chunk.Write(2, line)
		suma.Chunk.Write(byte(c.OP_ADD), line)
		suma.Chunk.Write(byte(c.OP_RETURN), line)
		writeClosure(chunk, suma)
		writeConstant(chunk, 2.0)
		writeConstant(chunk, 3.0)
		writeCall(chunk, 2)

		err := run(vm, chunk)

		utils.AssertNoError(t, err)
		utils.AssertValue(t, len(stackValues(vm)), 1)
		utils.AssertValue(t, stackValues(vm)[0], 5.0)
	})

	t.Run("Llamar a una función muchas veces seguidas no desborda los frames", func(t *testing.T) {
		vm, chunk := setUpVM(t)
		var out bytes.Buffer
		vm.out = &out
		callee := returning(1.0)
		for i := 0; i < framesMax+5; i++ {
			writeClosure(chunk, callee)
			writeCallStatement(chunk)
		}
		writeConstant(chunk, "listo")
		chunk.Write(byte(c.OP_PRINT), line)

		err := run(vm, chunk)

		utils.AssertNoError(t, err)
		utils.AssertValue(t, out.String(), "listo\n")
	})

	t.Run("El valor devuelto pasa por varias llamadas anidadas", func(t *testing.T) {
		vm, chunk := setUpVM(t)
		var out bytes.Buffer
		vm.out = &out
		outer := c.NewFunction()
		writeClosure(&outer.Chunk, returning(5.0))
		writeCall(&outer.Chunk, 0)
		outer.Chunk.Write(byte(c.OP_RETURN), line)
		writeClosure(chunk, outer)
		writeCall(chunk, 0)
		chunk.Write(byte(c.OP_PRINT), line)

		err := run(vm, chunk)

		utils.AssertNoError(t, err)
		utils.AssertValue(t, out.String(), "5\n")
	})
}

func TestStackTrace(t *testing.T) {
	callAt := func(chunk *c.Chunk, line int) {
		chunk.Write(byte(c.OP_CALL), line)
		chunk.Write(0, line)
	}
	failingCallAt := func(chunk *c.Chunk, line int) {
		chunk.Write(byte(c.OP_NIL), line)
		callAt(chunk, line)
	}

	t.Run("Un error en el script muestra la línea del script", func(t *testing.T) {
		vm, chunk := setUpVM(t)
		failingCallAt(chunk, 3)

		err := run(vm, chunk)

		utils.AssertValue(t, err.Error(), "[line 3] Error: can only call functions and classes\n"+
			"[line 3] in script")
	})

	t.Run("Un error dentro de una función muestra esa función y después el script", func(t *testing.T) {
		vm, chunk := setUpVM(t)
		f := c.NewFunction()
		f.Name = "f"
		failingCallAt(&f.Chunk, 2)
		writeClosure(chunk, f)
		callAt(chunk, 5)

		err := run(vm, chunk)

		utils.AssertValue(t, err.Error(), "[line 2] Error: can only call functions and classes\n"+
			"[line 2] in f()\n"+
			"[line 5] in script")
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

			err := run(vm, chunk)

			utils.AssertNoError(t, err)
			utils.AssertValue(t, len(stackValues(vm)), 1)
			utils.AssertValue(t, stackValues(vm)[0], test.value)
		})
	}

	t.Run("Dos constantes mantienen el orden", func(t *testing.T) {
		vm, chunk := setUpVM(t)
		writeConstant(chunk, 1.0)
		writeConstant(chunk, 2.0)

		err := run(vm, chunk)

		utils.AssertNoError(t, err)
		utils.AssertValue(t, len(stackValues(vm)), 2)
		utils.AssertValue(t, stackValues(vm)[0], 1.0)
		utils.AssertValue(t, stackValues(vm)[1], 2.0)
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

		err := run(vm, chunk)

		utils.AssertNoError(t, err)
		utils.AssertValue(t, len(stackValues(vm)), 2)
		// El stack sigue el orden de los operandos (1, 0), no el del pool (0, 1)
		utils.AssertValue(t, stackValues(vm)[0], 20.0)
		utils.AssertValue(t, stackValues(vm)[1], 10.0)
	})

	t.Run("Error si el índice no existe en el pool", func(t *testing.T) {
		vm, chunk := setUpVM(t)
		chunk.AddConstant(10.0)
		chunk.Write(byte(c.OP_CONSTANT), line)
		chunk.Write(5, 1)

		err := run(vm, chunk)

		utils.AssertError(t, err)
	})

	t.Run("Error si falta el operando", func(t *testing.T) {
		vm, chunk := setUpVM(t)
		chunk.AddConstant(10.0)
		chunk.Write(byte(c.OP_CONSTANT), line)

		err := run(vm, chunk)

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

			err := run(vm, chunk)

			utils.AssertNoError(t, err)
			utils.AssertValue(t, len(stackValues(vm)), 1)
			utils.AssertValue(t, stackValues(vm)[0], test.expected)
		})
	}
}

func TestOpPrint(t *testing.T) {
	namedFunction := c.NewFunction()
	namedFunction.Name = "f"

	tests := []struct {
		name     string
		value    any
		expected string
	}{
		{"número", 1.0, "1\n"},
		{"string", "hola", "hola\n"},
		{"booleano", true, "true\n"},
		{"nil", nil, "nil\n"},
		{"función", namedFunction, "<fn f>\n"},
	}

	for _, test := range tests {
		t.Run("Imprime "+test.name, func(t *testing.T) {
			vm, chunk := setUpVM(t)
			var out bytes.Buffer
			vm.out = &out
			writeConstant(chunk, test.value)
			chunk.Write(byte(c.OP_PRINT), line)

			err := run(vm, chunk)

			utils.AssertNoError(t, err)
			utils.AssertValue(t, out.String(), test.expected)
			utils.AssertValue(t, len(stackValues(vm)), 0)
		})
	}

	t.Run("Error si el stack está vacío", func(t *testing.T) {
		vm, chunk := setUpVM(t)
		var out bytes.Buffer
		vm.out = &out
		chunk.Write(byte(c.OP_PRINT), line)

		err := run(vm, chunk)

		utils.AssertErrorMessage(t, err, "stack underflow")
		utils.AssertValue(t, out.String(), "")
	})
}

func TestOpPop(t *testing.T) {
	t.Run("Saca solo el valor del tope", func(t *testing.T) {
		vm, chunk := setUpVM(t)
		writeConstant(chunk, 1.0)
		writeConstant(chunk, 2.0)
		chunk.Write(byte(c.OP_POP), line)

		err := run(vm, chunk)

		utils.AssertNoError(t, err)
		utils.AssertValue(t, len(stackValues(vm)), 1)
		utils.AssertValue(t, stackValues(vm)[0], 1.0)
	})

	t.Run("Error si el stack está vacío", func(t *testing.T) {
		vm, chunk := setUpVM(t)
		chunk.Write(byte(c.OP_POP), line)

		err := run(vm, chunk)

		utils.AssertErrorMessage(t, err, "stack underflow")
	})
}

func TestOpDefineGlobal(t *testing.T) {
	t.Run("Guarda el valor del tope con el nombre de la variable", func(t *testing.T) {
		vm, chunk := setUpVM(t)
		name := chunk.AddConstant("a")
		writeConstant(chunk, 1.0)
		chunk.Write(byte(c.OP_DEFINE_GLOBAL), line)
		chunk.Write(byte(name), line)

		err := run(vm, chunk)

		utils.AssertNoError(t, err)
		utils.AssertValue(t, vm.globals["a"], 1.0)
		utils.AssertValue(t, len(stackValues(vm)), 0)
	})

	t.Run("Redefinir una variable pisa su valor", func(t *testing.T) {
		vm, chunk := setUpVM(t)
		name := chunk.AddConstant("a")
		writeConstant(chunk, 1.0)
		chunk.Write(byte(c.OP_DEFINE_GLOBAL), line)
		chunk.Write(byte(name), line)
		writeConstant(chunk, 2.0)
		chunk.Write(byte(c.OP_DEFINE_GLOBAL), line)
		chunk.Write(byte(name), line)

		err := run(vm, chunk)

		utils.AssertNoError(t, err)
		utils.AssertValue(t, vm.globals["a"], 2.0)
	})

	t.Run("Error si el stack está vacío", func(t *testing.T) {
		vm, chunk := setUpVM(t)
		name := chunk.AddConstant("a")
		chunk.Write(byte(c.OP_DEFINE_GLOBAL), line)
		chunk.Write(byte(name), line)

		err := run(vm, chunk)

		utils.AssertErrorMessage(t, err, "stack underflow")
	})
}

func TestOpGetGlobal(t *testing.T) {
	t.Run("Apila el valor de la variable", func(t *testing.T) {
		vm, chunk := setUpVM(t)
		vm.globals["a"] = 1.0
		name := chunk.AddConstant("a")
		chunk.Write(byte(c.OP_GET_GLOBAL), line)
		chunk.Write(byte(name), line)

		err := run(vm, chunk)

		utils.AssertNoError(t, err)
		utils.AssertValue(t, len(stackValues(vm)), 1)
		utils.AssertValue(t, stackValues(vm)[0], 1.0)
	})

	t.Run("Una variable declarada sin valor se puede leer y vale nil", func(t *testing.T) {
		vm, chunk := setUpVM(t)
		vm.globals["a"] = nil
		name := chunk.AddConstant("a")
		chunk.Write(byte(c.OP_GET_GLOBAL), line)
		chunk.Write(byte(name), line)

		err := run(vm, chunk)

		utils.AssertNoError(t, err)
		utils.AssertValue(t, len(stackValues(vm)), 1)
		utils.AssertValue(t, stackValues(vm)[0], nil)
	})

	t.Run("Error si la variable no está definida", func(t *testing.T) {
		vm, chunk := setUpVM(t)
		name := chunk.AddConstant("a")
		chunk.Write(byte(c.OP_GET_GLOBAL), line)
		chunk.Write(byte(name), line)

		err := run(vm, chunk)

		utils.AssertErrorMessage(t, err, "undefined variable 'a'")
	})
}

func TestOpSetGlobal(t *testing.T) {
	t.Run("Asignar cambia el valor de la variable y lo deja en el stack", func(t *testing.T) {
		vm, chunk := setUpVM(t)
		vm.globals["a"] = 1.0
		name := chunk.AddConstant("a")
		writeConstant(chunk, 2.0)
		chunk.Write(byte(c.OP_SET_GLOBAL), line)
		chunk.Write(byte(name), line)

		err := run(vm, chunk)

		utils.AssertNoError(t, err)
		utils.AssertValue(t, vm.globals["a"], 2.0)
		utils.AssertValue(t, len(stackValues(vm)), 1)
		utils.AssertValue(t, stackValues(vm)[0], 2.0)
	})

	t.Run("Error si la variable no está definida y no se crea", func(t *testing.T) {
		vm, chunk := setUpVM(t)
		name := chunk.AddConstant("a")
		writeConstant(chunk, 2.0)
		chunk.Write(byte(c.OP_SET_GLOBAL), line)
		chunk.Write(byte(name), line)

		err := run(vm, chunk)

		utils.AssertErrorMessage(t, err, "undefined variable 'a'")
		_, defined := vm.globals["a"]
		utils.AssertValue(t, defined, false)
	})
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

			err := run(vm, chunk)

			utils.AssertNoError(t, err)
			utils.AssertValue(t, len(stackValues(vm)), 1)
			utils.AssertValue(t, stackValues(vm)[0], test.expected)
		})
	}

	t.Run("Negar dos veces devuelve el valor original", func(t *testing.T) {
		vm, chunk := setUpVM(t)
		writeConstant(chunk, 5.0)
		chunk.Write(byte(c.OP_NEGATE), line)
		chunk.Write(byte(c.OP_NEGATE), line)

		err := run(vm, chunk)

		utils.AssertNoError(t, err)
		utils.AssertValue(t, len(stackValues(vm)), 1)
		utils.AssertValue(t, stackValues(vm)[0], 5.0)
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

			err := run(vm, chunk)

			utils.AssertErrorMessage(t, err, "operand must be a number")
		})
	}

	t.Run("El error indica la línea de la instrucción", func(t *testing.T) {
		vm, chunk := setUpVM(t)
		index := chunk.AddConstant("a")
		chunk.Write(byte(c.OP_CONSTANT), 7)
		chunk.Write(byte(index), 7)
		chunk.Write(byte(c.OP_NEGATE), 7)

		err := run(vm, chunk)

		utils.AssertErrorMessage(t, err, "[line 7]")
	})

	t.Run("Error si el stack está vacío", func(t *testing.T) {
		vm, chunk := setUpVM(t)
		chunk.Write(byte(c.OP_NEGATE), line)

		err := run(vm, chunk)

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

			err := run(vm, chunk)

			utils.AssertNoError(t, err)
			utils.AssertValue(t, len(stackValues(vm)), 1)
			utils.AssertValue(t, stackValues(vm)[0], test.expected)
		})
	}

	t.Run("Concatena dos strings", func(t *testing.T) {
		vm, chunk := setUpVM(t)
		writeConstant(chunk, "Hello")
		writeConstant(chunk, " World")
		chunk.Write(byte(c.OP_ADD), line)

		err := run(vm, chunk)

		utils.AssertNoError(t, err)
		utils.AssertValue(t, len(stackValues(vm)), 1)
		utils.AssertValue(t, stackValues(vm)[0], "Hello World")
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

			err := run(vm, chunk)

			utils.AssertErrorMessage(t, err, "operands must be two numbers or two strings")
		})
	}

	t.Run("Error si el stack tiene un solo elemento", func(t *testing.T) {
		vm, chunk := setUpVM(t)
		writeConstant(chunk, 1.0)
		chunk.Write(byte(c.OP_ADD), line)

		err := run(vm, chunk)

		utils.AssertErrorMessage(t, err, "stack underflow")
	})

	t.Run("Error si el stack está vacío", func(t *testing.T) {
		vm, chunk := setUpVM(t)
		chunk.Write(byte(c.OP_ADD), line)

		err := run(vm, chunk)

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

			err := run(vm, chunk)

			utils.AssertNoError(t, err)
			utils.AssertValue(t, len(stackValues(vm)), 1)
			utils.AssertValue(t, stackValues(vm)[0], test.expected)
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

			err := run(vm, chunk)

			utils.AssertErrorMessage(t, err, "operands must be numbers")
		})
	}

	t.Run("Error si el stack tiene un solo elemento", func(t *testing.T) {
		vm, chunk := setUpVM(t)
		writeConstant(chunk, 1.0)
		chunk.Write(byte(c.OP_SUBTRACT), line)

		err := run(vm, chunk)

		utils.AssertErrorMessage(t, err, "stack underflow")
	})

	t.Run("Error si el stack está vacío", func(t *testing.T) {
		vm, chunk := setUpVM(t)
		chunk.Write(byte(c.OP_SUBTRACT), line)

		err := run(vm, chunk)

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

			err := run(vm, chunk)

			utils.AssertNoError(t, err)
			utils.AssertValue(t, len(stackValues(vm)), 1)
			utils.AssertValue(t, stackValues(vm)[0], test.expected)
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

			err := run(vm, chunk)

			utils.AssertErrorMessage(t, err, "operands must be numbers")
		})
	}

	t.Run("Error si el stack tiene un solo elemento", func(t *testing.T) {
		vm, chunk := setUpVM(t)
		writeConstant(chunk, 1.0)
		chunk.Write(byte(c.OP_MULTIPLY), line)

		err := run(vm, chunk)

		utils.AssertErrorMessage(t, err, "stack underflow")
	})

	t.Run("Error si el stack está vacío", func(t *testing.T) {
		vm, chunk := setUpVM(t)
		chunk.Write(byte(c.OP_MULTIPLY), line)

		err := run(vm, chunk)

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

			err := run(vm, chunk)

			utils.AssertNoError(t, err)
			utils.AssertValue(t, len(stackValues(vm)), 1)
			utils.AssertValue(t, stackValues(vm)[0], test.expected)
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

			err := run(vm, chunk)

			utils.AssertErrorMessage(t, err, "operands must be numbers")
		})
	}

	t.Run("Error si el stack tiene un solo elemento", func(t *testing.T) {
		vm, chunk := setUpVM(t)
		writeConstant(chunk, 1.0)
		chunk.Write(byte(c.OP_DIVIDE), line)

		err := run(vm, chunk)

		utils.AssertErrorMessage(t, err, "stack underflow")
	})

	t.Run("Error si el stack está vacío", func(t *testing.T) {
		vm, chunk := setUpVM(t)
		chunk.Write(byte(c.OP_DIVIDE), line)

		err := run(vm, chunk)

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

			err := run(vm, chunk)

			utils.AssertNoError(t, err)
			utils.AssertValue(t, len(stackValues(vm)), 1)
			utils.AssertValue(t, stackValues(vm)[0], test.expected)
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

			err := run(vm, chunk)

			utils.AssertNoError(t, err)
			utils.AssertValue(t, len(stackValues(vm)), 1)
			result, ok := stackValues(vm)[0].(float64)
			if !ok || !math.IsNaN(result) {
				t.Errorf("expected NaN, got %v", stackValues(vm)[0])
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

			err := run(vm, chunk)

			utils.AssertErrorMessage(t, err, "operands must be numbers")
		})
	}

	t.Run("Error si el stack tiene un solo elemento", func(t *testing.T) {
		vm, chunk := setUpVM(t)
		writeConstant(chunk, 1.0)
		chunk.Write(byte(c.OP_MODULO), line)

		err := run(vm, chunk)

		utils.AssertErrorMessage(t, err, "stack underflow")
	})

	t.Run("Error si el stack está vacío", func(t *testing.T) {
		vm, chunk := setUpVM(t)
		chunk.Write(byte(c.OP_MODULO), line)

		err := run(vm, chunk)

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

			err := run(vm, chunk)

			utils.AssertNoError(t, err)
			utils.AssertValue(t, len(stackValues(vm)), 1)
			utils.AssertValue(t, stackValues(vm)[0], test.expected)
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

			err := run(vm, chunk)

			utils.AssertErrorMessage(t, err, "operands must be numbers")
		})
	}

	t.Run("Error si el stack tiene un solo elemento", func(t *testing.T) {
		vm, chunk := setUpVM(t)
		writeConstant(chunk, 1.0)
		chunk.Write(byte(c.OP_GREATER), line)

		err := run(vm, chunk)

		utils.AssertErrorMessage(t, err, "stack underflow")
	})

	t.Run("Error si el stack está vacío", func(t *testing.T) {
		vm, chunk := setUpVM(t)
		chunk.Write(byte(c.OP_GREATER), line)

		err := run(vm, chunk)

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

			err := run(vm, chunk)

			utils.AssertNoError(t, err)
			utils.AssertValue(t, len(stackValues(vm)), 1)
			utils.AssertValue(t, stackValues(vm)[0], test.expected)
		})
	}

	t.Run("Error si el stack está vacío", func(t *testing.T) {
		vm, chunk := setUpVM(t)
		chunk.Write(byte(c.OP_NOT), line)

		err := run(vm, chunk)

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

			err := run(vm, chunk)

			utils.AssertNoError(t, err)
			utils.AssertValue(t, len(stackValues(vm)), 1)
			utils.AssertValue(t, stackValues(vm)[0], test.expected)
		})
	}

	t.Run("Error si el stack tiene un solo elemento", func(t *testing.T) {
		vm, chunk := setUpVM(t)
		writeConstant(chunk, 1.0)
		chunk.Write(byte(c.OP_EQUAL), line)

		err := run(vm, chunk)

		utils.AssertErrorMessage(t, err, "stack underflow")
	})

	t.Run("Error si el stack está vacío", func(t *testing.T) {
		vm, chunk := setUpVM(t)
		chunk.Write(byte(c.OP_EQUAL), line)

		err := run(vm, chunk)

		utils.AssertErrorMessage(t, err, "stack underflow")
	})

	t.Run("Un string armado en runtime es igual a un literal con el mismo contenido", func(t *testing.T) {
		vm, chunk := setUpVM(t)
		writeConstant(chunk, "st")
		writeConstant(chunk, "ri")
		chunk.Write(byte(c.OP_ADD), line)
		writeConstant(chunk, "stri")
		chunk.Write(byte(c.OP_EQUAL), line)

		err := run(vm, chunk)

		utils.AssertNoError(t, err)
		utils.AssertValue(t, len(stackValues(vm)), 1)
		utils.AssertValue(t, stackValues(vm)[0], true)
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

			err := run(vm, chunk)

			utils.AssertNoError(t, err)
			utils.AssertValue(t, len(stackValues(vm)), 1)
			utils.AssertValue(t, stackValues(vm)[0], test.expected)
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

			err := run(vm, chunk)

			utils.AssertErrorMessage(t, err, "operands must be numbers")
		})
	}

	t.Run("Error si el stack tiene un solo elemento", func(t *testing.T) {
		vm, chunk := setUpVM(t)
		writeConstant(chunk, 1.0)
		chunk.Write(byte(c.OP_LESS), line)

		err := run(vm, chunk)

		utils.AssertErrorMessage(t, err, "stack underflow")
	})

	t.Run("Error si el stack está vacío", func(t *testing.T) {
		vm, chunk := setUpVM(t)
		chunk.Write(byte(c.OP_LESS), line)

		err := run(vm, chunk)

		utils.AssertErrorMessage(t, err, "stack underflow")
	})
}

func TestRuntimeErrorResetsStack(t *testing.T) {
	t.Run("Un error en un operador unario vacía el stack", func(t *testing.T) {
		vm, chunk := setUpVM(t)
		chunk.Write(byte(c.OP_FALSE), line)
		chunk.Write(byte(c.OP_NEGATE), line)

		err := run(vm, chunk)

		utils.AssertErrorMessage(t, err, "operand must be a number")
		utils.AssertValue(t, len(stackValues(vm)), 0)
	})

	t.Run("Un error en un operador binario vacía el stack", func(t *testing.T) {
		vm, chunk := setUpVM(t)
		writeConstant(chunk, "a")
		writeConstant(chunk, 1.0)
		chunk.Write(byte(c.OP_ADD), line)

		err := run(vm, chunk)

		utils.AssertErrorMessage(t, err, "operands must be two numbers or two strings")
		utils.AssertValue(t, len(stackValues(vm)), 0)
	})

	t.Run("También se descartan los valores que había debajo de la operación que falla", func(t *testing.T) {
		vm, chunk := setUpVM(t)
		writeConstant(chunk, 10.0)
		chunk.Write(byte(c.OP_TRUE), line)
		chunk.Write(byte(c.OP_NEGATE), line)

		err := run(vm, chunk)

		utils.AssertErrorMessage(t, err, "operand must be a number")
		utils.AssertValue(t, len(stackValues(vm)), 0)
	})
}

func TestRunChunks(t *testing.T) {
	t.Run("Una misma VM ejecuta cada chunk nuevo desde el principio", func(t *testing.T) {
		vm := NewVirtualMachine()
		var out bytes.Buffer
		vm.out = &out

		first := c.NewChunk()
		writeConstant(first, 1.0)
		first.Write(byte(c.OP_PRINT), line)
		first.Write(byte(c.OP_NIL), line)
		first.Write(byte(c.OP_RETURN), line)
		second := c.NewChunk()
		writeConstant(second, 2.0)
		second.Write(byte(c.OP_PRINT), line)
		second.Write(byte(c.OP_NIL), line)
		second.Write(byte(c.OP_RETURN), line)

		utils.AssertNoError(t, run(vm, first))
		utils.AssertNoError(t, run(vm, second))

		utils.AssertValue(t, out.String(), "1\n2\n")
	})

	t.Run("Una variable definida en una ejecución se puede leer en la siguiente", func(t *testing.T) {
		vm := NewVirtualMachine()
		var out bytes.Buffer
		vm.out = &out

		define := c.NewChunk()
		name := define.AddConstant("a")
		writeConstant(define, 1.0)
		define.Write(byte(c.OP_DEFINE_GLOBAL), line)
		define.Write(byte(name), line)
		define.Write(byte(c.OP_NIL), line)
		define.Write(byte(c.OP_RETURN), line)

		read := c.NewChunk()
		name = read.AddConstant("a")
		read.Write(byte(c.OP_GET_GLOBAL), line)
		read.Write(byte(name), line)
		read.Write(byte(c.OP_PRINT), line)
		read.Write(byte(c.OP_NIL), line)
		read.Write(byte(c.OP_RETURN), line)

		utils.AssertNoError(t, run(vm, define))
		utils.AssertNoError(t, run(vm, read))

		utils.AssertValue(t, out.String(), "1\n")
	})
}

func TestOpGetLocal(t *testing.T) {
	t.Run("Copia el valor sin sacarlo de su lugar", func(t *testing.T) {
		vm, chunk := setUpVM(t)
		writeConstant(chunk, 10.0)
		chunk.Write(byte(c.OP_GET_LOCAL), line)
		chunk.Write(1, line)

		err := run(vm, chunk)

		utils.AssertNoError(t, err)
		utils.AssertValue(t, stackValues(vm)[0], 10.0)
		utils.AssertValue(t, stackValues(vm)[1], 10.0)
	})

	t.Run("Copia al tope el valor del índice indicado, no el ultimo", func(t *testing.T) {
		vm, chunk := setUpVM(t)
		writeConstant(chunk, 10.0)
		writeConstant(chunk, 20.0)
		chunk.Write(byte(c.OP_GET_LOCAL), line)
		chunk.Write(1, line) // el 10, en indice 1 (el 0 es la función del script)

		err := run(vm, chunk)

		utils.AssertNoError(t, err)
		utils.AssertValue(t, len(stackValues(vm)), 3)
		utils.AssertValue(t, stackValues(vm)[2], 10.0)
	})
}

func TestOpSetLocal(t *testing.T) {
	t.Run("Asignar cambia el valor de la variable y lo deja en el stack", func(t *testing.T) {
		vm, chunk := setUpVM(t)
		const variableIndex = 1
		initialValue, newValue := 1.0, 2.0

		writeConstant(chunk, initialValue)
		writeConstant(chunk, newValue)
		chunk.Write(byte(c.OP_SET_LOCAL), line)
		chunk.Write(variableIndex, line)

		err := run(vm, chunk)

		utils.AssertNoError(t, err)
		utils.AssertValue(t, vm.stack[variableIndex], newValue)
		utils.AssertValue(t, vm.stack[len(vm.stack)-1], newValue)
	})
}

const (
	literalSize  = 1
	popSize      = 1
	printSize    = 1
	constantSize = 2 // OP_CONSTANT + indice
	jumpSize     = 3 // OP_JUMP - OP_JUMP_IF_FALSE + 2 bytes de offset
	loopSize     = 3 // OP_LOOP + 2 bytes de offset
	localSize    = 2 // OP_GET_LOCAL / OP_SET_LOCAL + índice
	binarySize   = 1 // OP_GREATER / OP_SUBTRACT

	// offsets de los saltos para un then o un else de la forma `print n;`
	skipThen = popSize + constantSize + printSize + jumpSize
	skipElse = popSize + constantSize + printSize
)

func TestOpJumpIfFalse(t *testing.T) {
	tests := []struct {
		name      string
		condition func(chunk *c.Chunk)
		expected  string
	}{
		{"Con false salta el cuerpo", writeOp(c.OP_FALSE), ""},
		{"Con nil salta el cuerpo", writeOp(c.OP_NIL), ""},
		{"Con true ejecuta el cuerpo", writeOp(c.OP_TRUE), "1\n"},
		{"Con 0 ejecuta el cuerpo (es truthy)", func(chunk *c.Chunk) { writeConstant(chunk, 0.0) }, "1\n"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// if (<condición>) print 1;
			vm, chunk := setUpVM(t)
			var out bytes.Buffer
			vm.out = &out
			test.condition(chunk)
			writeJump(chunk, c.OP_JUMP_IF_FALSE, skipThen)
			chunk.Write(byte(c.OP_POP), line)
			writeConstant(chunk, 1.0)
			chunk.Write(byte(c.OP_PRINT), line)
			writeJump(chunk, c.OP_JUMP, popSize)
			chunk.Write(byte(c.OP_POP), line)
			chunk.Write(byte(c.OP_NIL), line)
			chunk.Write(byte(c.OP_RETURN), line)

			err := run(vm, chunk)

			utils.AssertNoError(t, err)
			utils.AssertValue(t, out.String(), test.expected)
		})
	}
}

func TestOpJump(t *testing.T) {
	tests := []struct {
		name      string
		condition c.OpCode
		expected  string
	}{
		{"Con true ejecuta el then y saltea el else", c.OP_TRUE, "1\n"},
		{"Con false saltea el then y ejecuta el else", c.OP_FALSE, "2\n"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// if (<condición>) print 1; else print 2;
			vm, chunk := setUpVM(t)
			var out bytes.Buffer
			vm.out = &out
			chunk.Write(byte(test.condition), line)
			writeJump(chunk, c.OP_JUMP_IF_FALSE, skipThen)
			chunk.Write(byte(c.OP_POP), line)
			writeConstant(chunk, 1.0)
			chunk.Write(byte(c.OP_PRINT), line)
			writeJump(chunk, c.OP_JUMP, skipElse)
			chunk.Write(byte(c.OP_POP), line)
			writeConstant(chunk, 2.0)
			chunk.Write(byte(c.OP_PRINT), line)
			chunk.Write(byte(c.OP_NIL), line)
			chunk.Write(byte(c.OP_RETURN), line)

			err := run(vm, chunk)

			utils.AssertNoError(t, err)
			utils.AssertValue(t, out.String(), test.expected)
		})
	}
}

func TestAnd(t *testing.T) {
	skipRight := popSize + literalSize // OP_POP + el literal derecho

	tests := []struct {
		name     string
		left     c.OpCode
		right    c.OpCode
		expected string
	}{
		{"true and true es true", c.OP_TRUE, c.OP_TRUE, "true\n"},
		{"true and false es false", c.OP_TRUE, c.OP_FALSE, "false\n"},
		{"false and true es false", c.OP_FALSE, c.OP_TRUE, "false\n"},
		{"false and false es false", c.OP_FALSE, c.OP_FALSE, "false\n"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// print <left> and <right>;
			vm, chunk := setUpVM(t)
			var out bytes.Buffer
			vm.out = &out
			chunk.Write(byte(test.left), line)
			writeJump(chunk, c.OP_JUMP_IF_FALSE, skipRight)
			chunk.Write(byte(c.OP_POP), line)
			chunk.Write(byte(test.right), line)
			chunk.Write(byte(c.OP_PRINT), line)
			chunk.Write(byte(c.OP_NIL), line)
			chunk.Write(byte(c.OP_RETURN), line)

			err := run(vm, chunk)

			utils.AssertNoError(t, err)
			utils.AssertValue(t, out.String(), test.expected)
		})
	}
}

func TestOr(t *testing.T) {
	skipRight := popSize + literalSize
	tests := []struct {
		name     string
		left     c.OpCode
		right    c.OpCode
		expected string
	}{
		{"true or true es true", c.OP_TRUE, c.OP_TRUE, "true\n"},
		{"true or false es true", c.OP_TRUE, c.OP_FALSE, "true\n"},
		{"false or true es true", c.OP_FALSE, c.OP_TRUE, "true\n"},
		{"false or false es false", c.OP_FALSE, c.OP_FALSE, "false\n"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// print <left> or <right>;
			vm, chunk := setUpVM(t)
			var out bytes.Buffer
			vm.out = &out
			chunk.Write(byte(test.left), line)
			writeJump(chunk, c.OP_JUMP_IF_FALSE, jumpSize)
			writeJump(chunk, c.OP_JUMP, skipRight)
			chunk.Write(byte(c.OP_POP), line)
			chunk.Write(byte(test.right), line)
			chunk.Write(byte(c.OP_PRINT), line)
			chunk.Write(byte(c.OP_NIL), line)
			chunk.Write(byte(c.OP_RETURN), line)

			err := run(vm, chunk)

			utils.AssertNoError(t, err)
			utils.AssertValue(t, out.String(), test.expected)
		})
	}
}

func TestOpLoop(t *testing.T) {
	t.Run("Vuelve a la condición y repite el cuerpo hasta que sea falsa", func(t *testing.T) {
		// { var i = 2; while (i > 0) { print i; i = i - 1; } }
		vm, chunk := setUpVM(t)
		var out bytes.Buffer
		vm.out = &out
		conditionSize := localSize + constantSize + binarySize                                          // i > 0
		bodySize := localSize + printSize + localSize + constantSize + binarySize + localSize + popSize // print i; i = i - 1;
		skipBody := popSize + bodySize + loopSize
		backToCondition := conditionSize + jumpSize + skipBody

		writeConstant(chunk, 2.0)
		// condición: inicio del ciclo
		chunk.Write(byte(c.OP_GET_LOCAL), line)
		chunk.Write(1, line)
		writeConstant(chunk, 0.0)
		chunk.Write(byte(c.OP_GREATER), line)
		writeJump(chunk, c.OP_JUMP_IF_FALSE, skipBody)
		chunk.Write(byte(c.OP_POP), line)
		// cuerpo: print i;
		chunk.Write(byte(c.OP_GET_LOCAL), line)
		chunk.Write(1, line)
		chunk.Write(byte(c.OP_PRINT), line)
		// cuerpo: i = i - 1;
		chunk.Write(byte(c.OP_GET_LOCAL), line)
		chunk.Write(1, line)
		writeConstant(chunk, 1.0)
		chunk.Write(byte(c.OP_SUBTRACT), line)
		chunk.Write(byte(c.OP_SET_LOCAL), line)
		chunk.Write(1, line)
		chunk.Write(byte(c.OP_POP), line)
		writeJump(chunk, c.OP_LOOP, backToCondition) // vuelve a la condición
		// salida
		chunk.Write(byte(c.OP_POP), line) // saca la condición
		chunk.Write(byte(c.OP_POP), line) // fin del bloque: saca i
		chunk.Write(byte(c.OP_NIL), line)
		chunk.Write(byte(c.OP_RETURN), line)

		err := run(vm, chunk)

		utils.AssertNoError(t, err)
		utils.AssertValue(t, out.String(), "2\n1\n")
	})

	t.Run("Con la condición falsa desde el inicio no entra al cuerpo", func(t *testing.T) {
		// while (false) print 1;
		vm, chunk := setUpVM(t)
		var out bytes.Buffer
		vm.out = &out
		skipBody := popSize + constantSize + printSize + loopSize
		backToCondition := literalSize + jumpSize + skipBody

		chunk.Write(byte(c.OP_FALSE), line)
		writeJump(chunk, c.OP_JUMP_IF_FALSE, skipBody)
		chunk.Write(byte(c.OP_POP), line)
		writeConstant(chunk, 1.0)
		chunk.Write(byte(c.OP_PRINT), line)
		writeJump(chunk, c.OP_LOOP, backToCondition)
		chunk.Write(byte(c.OP_POP), line)
		chunk.Write(byte(c.OP_NIL), line)
		chunk.Write(byte(c.OP_RETURN), line)

		err := run(vm, chunk)

		utils.AssertNoError(t, err)
		utils.AssertValue(t, out.String(), "")
	})
}

func TestUpvalues(t *testing.T) {
	t.Run("OP_GET_UPVALUE lee una local de la función de afuera", func(t *testing.T) {
		// fun afuera() { var x = "hola"; fun adentro() { print x; } adentro(); }  afuera();
		vm, chunk := setUpVM(t)
		var out bytes.Buffer
		vm.out = &out
		afuera := c.NewFunction()
		writeConstant(&afuera.Chunk, "hola")
		writeClosure(&afuera.Chunk, newPrintUpvalue(), 1, 1) // captura la local 1 de afuera: x
		writeCallStatement(&afuera.Chunk)
		writeReturnNil(&afuera.Chunk)
		writeClosure(chunk, afuera)
		writeCallStatement(chunk)

		err := run(vm, chunk)

		utils.AssertNoError(t, err)
		utils.AssertValue(t, out.String(), "hola\n")
	})

	t.Run("OP_SET_UPVALUE cambia la variable de la función de afuera", func(t *testing.T) {
		// fun afuera() { var x = 1; fun adentro() { x = 2; } adentro(); print x; }  afuera();
		vm, chunk := setUpVM(t)
		var out bytes.Buffer
		vm.out = &out
		adentro := c.NewFunction()
		adentro.UpvalueCount = 1
		writeConstant(&adentro.Chunk, 2.0)
		adentro.Chunk.Write(byte(c.OP_SET_UPVALUE), line)
		adentro.Chunk.Write(0, line)
		adentro.Chunk.Write(byte(c.OP_POP), line)
		writeReturnNil(&adentro.Chunk)
		afuera := c.NewFunction()
		writeConstant(&afuera.Chunk, 1.0)
		writeClosure(&afuera.Chunk, adentro, 1, 1)
		writeCallStatement(&afuera.Chunk)
		afuera.Chunk.Write(byte(c.OP_GET_LOCAL), line)
		afuera.Chunk.Write(1, line)
		afuera.Chunk.Write(byte(c.OP_PRINT), line)
		writeReturnNil(&afuera.Chunk)
		writeClosure(chunk, afuera)
		writeCallStatement(chunk)

		err := run(vm, chunk)

		utils.AssertNoError(t, err)
		utils.AssertValue(t, out.String(), "2\n") // afuera ve el cambio: es la misma variable
	})

	t.Run("Con tres niveles, la función de adentro usa el upvalue de la del medio", func(t *testing.T) {
		// fun a() { var x = "hola"; fun b() { fun c() { print x; } c(); } b(); }  a();
		vm, chunk := setUpVM(t)
		var out bytes.Buffer
		vm.out = &out
		b := c.NewFunction()
		b.UpvalueCount = 1
		writeClosure(&b.Chunk, newPrintUpvalue(), 0, 0) // c comparte el upvalue 0 de b
		writeCallStatement(&b.Chunk)
		writeReturnNil(&b.Chunk)
		a := c.NewFunction()
		writeConstant(&a.Chunk, "hola")
		writeClosure(&a.Chunk, b, 1, 1) // b captura la local 1 de a
		writeCallStatement(&a.Chunk)
		writeReturnNil(&a.Chunk)
		writeClosure(chunk, a)
		writeCallStatement(chunk)

		err := run(vm, chunk)

		utils.AssertNoError(t, err)
		utils.AssertValue(t, out.String(), "hola\n")
	})
}

func TestCloseUpvalues(t *testing.T) {
	t.Run("OP_CLOSE_UPVALUE mantiene la variable aunque salga del stack al cerrar un bloque", func(t *testing.T) {
		// { var x = "hola"; fun mostrar() { print x; } f = mostrar; }  f();
		vm, chunk := setUpVM(t)
		var out bytes.Buffer
		vm.out = &out
		writeConstant(chunk, "hola")
		writeClosure(chunk, newPrintUpvalue(), 1, 1)
		writeDefineGlobal(chunk, "f")
		chunk.Write(byte(c.OP_CLOSE_UPVALUE), line)
		writeCallGlobal(chunk, "f")

		err := run(vm, chunk)

		utils.AssertNoError(t, err)
		utils.AssertValue(t, out.String(), "hola\n")
	})

	t.Run("OP_RETURN cierra los upvalues de la función que termina", func(t *testing.T) {
		// fun crearContador() { var cuenta = 0; fun contar() { cuenta = cuenta + 1; print cuenta; } return contar; }
		// var contador = crearContador();  contador();  contador();
		vm, chunk := setUpVM(t)
		var out bytes.Buffer
		vm.out = &out
		contar := c.NewFunction()
		contar.UpvalueCount = 1
		contar.Chunk.Write(byte(c.OP_GET_UPVALUE), line)
		contar.Chunk.Write(0, line)
		writeConstant(&contar.Chunk, 1.0)
		contar.Chunk.Write(byte(c.OP_ADD), line)
		contar.Chunk.Write(byte(c.OP_SET_UPVALUE), line)
		contar.Chunk.Write(0, line)
		contar.Chunk.Write(byte(c.OP_PRINT), line)
		writeReturnNil(&contar.Chunk)
		crearContador := c.NewFunction()
		writeConstant(&crearContador.Chunk, 0.0)
		writeClosure(&crearContador.Chunk, contar, 1, 1)
		crearContador.Chunk.Write(byte(c.OP_RETURN), line)
		writeClosure(chunk, crearContador)
		writeCall(chunk, 0)
		writeDefineGlobal(chunk, "contador")
		writeCallGlobal(chunk, "contador")
		writeCallGlobal(chunk, "contador")

		err := run(vm, chunk)

		utils.AssertNoError(t, err)
		utils.AssertValue(t, out.String(), "1\n2\n")
	})

	t.Run("Dos closures que capturan la misma variable la siguen compartiendo después de cerrarse", func(t *testing.T) {
		// fun main() { var a = "initial"; fun set() { a = "updated"; } fun get() { print a; } globalSet = set; globalGet = get; }
		// main();  globalSet();  globalGet();
		vm, chunk := setUpVM(t)
		var out bytes.Buffer
		vm.out = &out
		set := c.NewFunction()
		set.UpvalueCount = 1
		writeConstant(&set.Chunk, "updated")
		set.Chunk.Write(byte(c.OP_SET_UPVALUE), line)
		set.Chunk.Write(0, line)
		set.Chunk.Write(byte(c.OP_POP), line)
		writeReturnNil(&set.Chunk)
		main := c.NewFunction()
		writeConstant(&main.Chunk, "initial")
		writeClosure(&main.Chunk, set, 1, 1)
		writeDefineGlobal(&main.Chunk, "globalSet")
		writeClosure(&main.Chunk, newPrintUpvalue(), 1, 1)
		writeDefineGlobal(&main.Chunk, "globalGet")
		writeReturnNil(&main.Chunk)
		writeClosure(chunk, main)
		writeCallStatement(chunk)
		writeCallGlobal(chunk, "globalSet")
		writeCallGlobal(chunk, "globalGet")

		err := run(vm, chunk)

		utils.AssertNoError(t, err)
		utils.AssertValue(t, out.String(), "updated\n")
	})
}

func TestOpClass(t *testing.T) {
	t.Run("Apila una clase nueva con el nombre de la constante", func(t *testing.T) {
		vm, chunk := setUpVM(t)
		name := chunk.AddConstant("MiClase")
		chunk.Write(byte(c.OP_CLASS), line)
		chunk.Write(byte(name), line)

		err := run(vm, chunk)

		utils.AssertNoError(t, err)
		utils.AssertValue(t, len(stackValues(vm)), 1)
		utils.AssertValue(t, stackValues(vm)[0].(*class).name, "MiClase")
	})
}

func TestOpMethod(t *testing.T) {
	writeMethod := func(chunk *c.Chunk, name string, method *c.Function) {
		writeClosure(chunk, method)
		index := chunk.AddConstant(name)
		chunk.Write(byte(c.OP_METHOD), line)
		chunk.Write(byte(index), line)
	}

	t.Run("Guarda el closure en los métodos de la clase con el nombre del operando", func(t *testing.T) {
		vm, chunk := setUpVM(t)
		auto := newClass("Auto")
		acelerar := c.NewFunction()
		writeConstant(chunk, auto)
		writeMethod(chunk, "acelerar", acelerar)

		err := run(vm, chunk)

		utils.AssertNoError(t, err)
		utils.AssertValue(t, len(auto.methods), 1)
		utils.AssertValue(t, auto.methods["acelerar"].function, acelerar)
	})

	t.Run("Saca el closure del stack y deja la clase en el tope", func(t *testing.T) {
		vm, chunk := setUpVM(t)
		auto := newClass("Auto")
		writeConstant(chunk, auto)
		writeMethod(chunk, "acelerar", c.NewFunction())

		err := run(vm, chunk)

		utils.AssertNoError(t, err)
		utils.AssertValue(t, len(stackValues(vm)), 1)
		utils.AssertValue(t, stackValues(vm)[0], auto)
	})

	t.Run("Varios OP_METHOD seguidos agregan cada método a la misma clase", func(t *testing.T) {
		vm, chunk := setUpVM(t)
		auto := newClass("Auto")
		acelerar, frenar := c.NewFunction(), c.NewFunction()
		writeConstant(chunk, auto)
		writeMethod(chunk, "acelerar", acelerar)
		writeMethod(chunk, "frenar", frenar)

		err := run(vm, chunk)

		utils.AssertNoError(t, err)
		utils.AssertValue(t, len(auto.methods), 2)
		utils.AssertValue(t, auto.methods["acelerar"].function, acelerar)
		utils.AssertValue(t, auto.methods["frenar"].function, frenar)
	})

	t.Run("Un método con el mismo nombre que otro lo pisa", func(t *testing.T) {
		vm, chunk := setUpVM(t)
		auto := newClass("Auto")
		primero, segundo := c.NewFunction(), c.NewFunction()
		writeConstant(chunk, auto)
		writeMethod(chunk, "acelerar", primero)
		writeMethod(chunk, "acelerar", segundo)

		err := run(vm, chunk)

		utils.AssertNoError(t, err)
		utils.AssertValue(t, len(auto.methods), 1)
		utils.AssertValue(t, auto.methods["acelerar"].function, segundo)
	})
}

func TestCallClass(t *testing.T) {
	callClass := func(chunk *c.Chunk, class *class) {
		writeConstant(chunk, class)
		writeCall(chunk, 0)
	}

	t.Run("Llamar a una clase deja en su lugar una instancia de esa clase", func(t *testing.T) {
		vm, chunk := setUpVM(t)
		auto := newClass("Auto")
		callClass(chunk, auto)

		err := run(vm, chunk)

		utils.AssertNoError(t, err)
		utils.AssertValue(t, len(stackValues(vm)), 1)
		created, isInstance := stackValues(vm)[0].(*instance)
		utils.AssertValue(t, isInstance, true)
		utils.AssertValue(t, created.class, auto)
	})

	t.Run("Cada llamada a una clase crea una instancia distinta", func(t *testing.T) {
		vm, chunk := setUpVM(t)
		auto := newClass("Auto")
		callClass(chunk, auto)
		callClass(chunk, auto)

		err := run(vm, chunk)

		utils.AssertNoError(t, err)
		first := stackValues(vm)[0].(*instance)
		second := stackValues(vm)[1].(*instance)
		utils.AssertValue(t, first != second, true)
	})

	t.Run("Una clase con init lo llama con la instancia nueva en el slot 0 y los argumentos después", func(t *testing.T) {
		vm, chunk := setUpVM(t)
		var out bytes.Buffer
		vm.out = &out
		auto := newClass("Auto")
		auto.methods["init"] = newMethod("init", 1, func(body *c.Chunk) {
			body.Write(byte(c.OP_GET_LOCAL), line)
			body.Write(0, line)
			body.Write(byte(c.OP_PRINT), line)
			body.Write(byte(c.OP_GET_LOCAL), line)
			body.Write(1, line)
			body.Write(byte(c.OP_PRINT), line)
		})
		writeConstant(chunk, auto)
		writeConstant(chunk, "rojo")
		writeCall(chunk, 1)

		err := run(vm, chunk)

		utils.AssertNoError(t, err)
		utils.AssertValue(t, out.String(), "Auto instance\nrojo\n")
	})

	t.Run("Error si se llama a una clase con init con otra cantidad de argumentos", func(t *testing.T) {
		vm, chunk := setUpVM(t)
		auto := newClass("Auto")
		auto.methods["init"] = newMethod("init", 1, func(body *c.Chunk) {})
		writeConstant(chunk, auto)
		writeCall(chunk, 0)

		err := run(vm, chunk)

		utils.AssertErrorMessage(t, err, "expected 1 arguments but got 0")
	})

	t.Run("Error si se le pasan argumentos a una clase sin init", func(t *testing.T) {
		vm, chunk := setUpVM(t)
		writeConstant(chunk, newClass("Auto"))
		writeConstant(chunk, "rojo")
		writeCall(chunk, 1)

		err := run(vm, chunk)

		utils.AssertErrorMessage(t, err, "expected 0 arguments but got 1")
	})
}

func TestCallBoundMethod(t *testing.T) {
	t.Run("Llamar a un bound method ejecuta el código del método", func(t *testing.T) {
		vm, chunk := setUpVM(t)
		var out bytes.Buffer
		vm.out = &out
		saludar := newMethod("saludar", 0, func(body *c.Chunk) {
			writeConstant(body, "hola")
			body.Write(byte(c.OP_PRINT), line)
		})
		writeConstant(chunk, newBoundMethod(newInstance(newClass("Persona")), saludar))
		writeCall(chunk, 0)

		err := run(vm, chunk)

		utils.AssertNoError(t, err)
		utils.AssertValue(t, out.String(), "hola\n")
	})

	t.Run("Los argumentos quedan atados a los parámetros del método", func(t *testing.T) {
		vm, chunk := setUpVM(t)
		var out bytes.Buffer
		vm.out = &out
		restar := newMethod("restar", 2, func(body *c.Chunk) {
			body.Write(byte(c.OP_GET_LOCAL), line)
			body.Write(1, line)
			body.Write(byte(c.OP_GET_LOCAL), line)
			body.Write(2, line)
			body.Write(byte(c.OP_SUBTRACT), line)
			body.Write(byte(c.OP_PRINT), line)
		})
		writeConstant(chunk, newBoundMethod(newInstance(newClass("Calculadora")), restar))
		writeConstant(chunk, 10.0)
		writeConstant(chunk, 4.0)
		writeCall(chunk, 2)

		err := run(vm, chunk)

		utils.AssertNoError(t, err)
		utils.AssertValue(t, out.String(), "6\n")
	})

	t.Run("Al llamar a un bound method el receptor queda en el slot 0 y los argumentos después", func(t *testing.T) {
		vm, chunk := setUpVM(t)
		var out bytes.Buffer
		vm.out = &out
		mostrar := newMethod("mostrar", 1, func(body *c.Chunk) {
			body.Write(byte(c.OP_GET_LOCAL), line)
			body.Write(0, line)
			body.Write(byte(c.OP_PRINT), line)
			body.Write(byte(c.OP_GET_LOCAL), line)
			body.Write(1, line)
			body.Write(byte(c.OP_PRINT), line)
		})
		writeConstant(chunk, newBoundMethod(newInstance(newClass("Auto")), mostrar))
		writeConstant(chunk, "rojo")
		writeCall(chunk, 1)

		err := run(vm, chunk)

		utils.AssertNoError(t, err)
		utils.AssertValue(t, out.String(), "Auto instance\nrojo\n")
	})

	t.Run("Error si se llama a un bound method con otra cantidad de argumentos", func(t *testing.T) {
		vm, chunk := setUpVM(t)
		acelerar := newMethod("acelerar", 1, func(body *c.Chunk) {})
		writeConstant(chunk, newBoundMethod(newInstance(newClass("Auto")), acelerar))
		writeCall(chunk, 0)

		err := run(vm, chunk)

		utils.AssertErrorMessage(t, err, "expected 1 arguments but got 0")
	})
}

func TestOpGetProperty(t *testing.T) {
	getProperty := func(chunk *c.Chunk, object any, name string) {
		writeConstant(chunk, object)
		index := chunk.AddConstant(name)
		chunk.Write(byte(c.OP_GET_PROPERTY), line)
		chunk.Write(byte(index), line)
	}

	t.Run("Reemplaza la instancia por el valor del campo", func(t *testing.T) {
		vm, chunk := setUpVM(t)
		auto := newInstance(newClass("Auto"))
		auto.fields["color"] = "rojo"
		getProperty(chunk, auto, "color")

		err := run(vm, chunk)

		utils.AssertNoError(t, err)
		utils.AssertValue(t, len(stackValues(vm)), 1)
		utils.AssertValue(t, stackValues(vm)[0], "rojo")
	})

	t.Run("Lee el campo con el nombre que indica el operando", func(t *testing.T) {
		vm, chunk := setUpVM(t)
		auto := newInstance(newClass("Auto"))
		auto.fields["color"] = "rojo"
		auto.fields["marca"] = "Fiat"
		getProperty(chunk, auto, "marca")

		err := run(vm, chunk)

		utils.AssertNoError(t, err)
		utils.AssertValue(t, stackValues(vm)[0], "Fiat")
	})

	t.Run("Un campo que vale nil existe y se puede leer", func(t *testing.T) {
		vm, chunk := setUpVM(t)
		auto := newInstance(newClass("Auto"))
		auto.fields["color"] = nil
		getProperty(chunk, auto, "color")

		err := run(vm, chunk)

		utils.AssertNoError(t, err)
		utils.AssertValue(t, len(stackValues(vm)), 1)
		utils.AssertValue(t, stackValues(vm)[0], nil)
	})

	t.Run("Si no hay un campo con ese nombre deja el método de la clase como bound method de la instancia", func(t *testing.T) {
		vm, chunk := setUpVM(t)
		auto := newClass("Auto")
		acelerar := c.NewFunction()
		auto.methods["acelerar"] = newClosure(acelerar)
		fiat := newInstance(auto)
		getProperty(chunk, fiat, "acelerar")

		err := run(vm, chunk)

		utils.AssertNoError(t, err)
		utils.AssertValue(t, len(stackValues(vm)), 1)
		bound, isBound := stackValues(vm)[0].(*boundMethod)
		utils.AssertValue(t, isBound, true)
		utils.AssertValue(t, bound.receiver, any(fiat))
		utils.AssertValue(t, bound.method, auto.methods["acelerar"])
	})

	t.Run("Un campo tapa a un método con el mismo nombre", func(t *testing.T) {
		vm, chunk := setUpVM(t)
		auto := newClass("Auto")
		auto.methods["color"] = newClosure(c.NewFunction())
		fiat := newInstance(auto)
		fiat.fields["color"] = "rojo"
		getProperty(chunk, fiat, "color")

		err := run(vm, chunk)

		utils.AssertNoError(t, err)
		utils.AssertValue(t, stackValues(vm)[0], "rojo")
	})

	t.Run("Leer dos veces el mismo método da dos bound methods distintos", func(t *testing.T) {
		vm, chunk := setUpVM(t)
		auto := newClass("Auto")
		auto.methods["acelerar"] = newClosure(c.NewFunction())
		fiat := newInstance(auto)
		getProperty(chunk, fiat, "acelerar")
		getProperty(chunk, fiat, "acelerar")

		err := run(vm, chunk)

		utils.AssertNoError(t, err)
		utils.AssertValue(t, stackValues(vm)[0] != stackValues(vm)[1], true)
	})

	t.Run("Error si la instancia no tiene ni un campo ni un método con ese nombre", func(t *testing.T) {
		vm, chunk := setUpVM(t)
		getProperty(chunk, newInstance(newClass("Auto")), "color")

		err := run(vm, chunk)

		utils.AssertErrorMessage(t, err, "undefined property 'color'")
	})

	t.Run("Error si se lee una propiedad de algo que no es una instancia", func(t *testing.T) {
		tests := []struct {
			name   string
			object any
		}{
			{"un número", 1.0},
			{"nil", nil},
			{"una clase", newClass("Auto")},
		}
		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				vm, chunk := setUpVM(t)
				getProperty(chunk, test.object, "color")

				err := run(vm, chunk)

				utils.AssertErrorMessage(t, err, "only instances have properties")
			})
		}
	})
}

func TestOpSetProperty(t *testing.T) {
	setProperty := func(chunk *c.Chunk, object any, name string, value any) {
		writeConstant(chunk, object)
		writeConstant(chunk, value)
		index := chunk.AddConstant(name)
		chunk.Write(byte(c.OP_SET_PROPERTY), line)
		chunk.Write(byte(index), line)
	}

	t.Run("Crea el campo si la instancia no lo tenía", func(t *testing.T) {
		vm, chunk := setUpVM(t)
		auto := newInstance(newClass("Auto"))
		setProperty(chunk, auto, "color", "rojo")

		err := run(vm, chunk)

		utils.AssertNoError(t, err)
		utils.AssertValue(t, auto.fields["color"], "rojo")
	})

	t.Run("Pisa el valor si el campo ya existía", func(t *testing.T) {
		vm, chunk := setUpVM(t)
		auto := newInstance(newClass("Auto"))
		auto.fields["color"] = "rojo"
		setProperty(chunk, auto, "color", "azul")

		err := run(vm, chunk)

		utils.AssertNoError(t, err)
		utils.AssertValue(t, auto.fields["color"], "azul")
		utils.AssertValue(t, len(auto.fields), 1)
	})

	t.Run("Asigna el campo con el nombre que indica el operando", func(t *testing.T) {
		vm, chunk := setUpVM(t)
		auto := newInstance(newClass("Auto"))
		auto.fields["color"] = "rojo"
		setProperty(chunk, auto, "marca", "Fiat")

		err := run(vm, chunk)

		utils.AssertNoError(t, err)
		utils.AssertValue(t, auto.fields["marca"], "Fiat")
		utils.AssertValue(t, auto.fields["color"], "rojo")
	})

	t.Run("Asignar nil deja el campo existiendo con nil", func(t *testing.T) {
		vm, chunk := setUpVM(t)
		auto := newInstance(newClass("Auto"))
		auto.fields["color"] = "rojo"
		setProperty(chunk, auto, "color", nil)

		err := run(vm, chunk)

		utils.AssertNoError(t, err)
		value, exists := auto.fields["color"]
		utils.AssertValue(t, exists, true)
		utils.AssertValue(t, value, nil)
	})

	t.Run("Deja en el stack solo el valor asignado donde estaba la instancia", func(t *testing.T) {
		vm, chunk := setUpVM(t)
		setProperty(chunk, newInstance(newClass("Auto")), "color", "rojo")

		err := run(vm, chunk)

		utils.AssertNoError(t, err)
		utils.AssertValue(t, len(stackValues(vm)), 1)
		utils.AssertValue(t, stackValues(vm)[0], "rojo")
	})

	t.Run("Error si se asigna un campo a algo que no es una instancia", func(t *testing.T) {
		tests := []struct {
			name   string
			object any
		}{
			{"un número", 1.0},
			{"nil", nil},
			{"una clase", newClass("Auto")},
		}
		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				vm, chunk := setUpVM(t)
				setProperty(chunk, test.object, "color", "rojo")

				err := run(vm, chunk)

				utils.AssertErrorMessage(t, err, "only instances have fields")
			})
		}
	})
}
