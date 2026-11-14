package virtualmachine

import (
	"bytes"
	"math"
	"testing"
	c "tp-lox/internal/chunk"
	"tp-lox/internal/utils"
)

var line = 1

func setUpVM(t *testing.T) (*VirtualMachine, *c.Chunk) {
	t.Helper()
	chunk := c.NewChunk()
	vm := NewVirtualMachine()
	return vm, chunk
}

func run(vm *VirtualMachine, chunk *c.Chunk) error {
	function := c.NewFunction()
	function.Chunk = *chunk
	return vm.Run(function)
}

func stackValues(vm *VirtualMachine) []any {
	if len(vm.stack) == 0 {
		return vm.stack
	}
	return vm.stack[1:]
}

func writeConstant(chunk *c.Chunk, value any) {
	index := chunk.AddConstant(value)
	chunk.Write(byte(c.OP_CONSTANT), line)
	chunk.Write(byte(index), line)
}

func writeCall(chunk *c.Chunk, argCount byte) {
	chunk.Write(byte(c.OP_CALL), line)
	chunk.Write(argCount, line)
}

func writeJump(chunk *c.Chunk, op c.OpCode, offset int) {
	chunk.Write(byte(op), line)
	chunk.Write(byte(offset>>8), line)
	chunk.Write(byte(offset), line)
}

func writeOp(op c.OpCode) func(chunk *c.Chunk) {
	return func(chunk *c.Chunk) { chunk.Write(byte(op), line) }
}

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
		utils.AssertValue(t, vm.stack[0], any(function))
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
		writeConstant(chunk, callee)
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
		writeConstant(chunk, callee)
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
		writeConstant(chunk, callee)
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
		writeConstant(chunk, callee)
		writeConstant(chunk, 1.0)
		writeCall(chunk, 1)

		err := run(vm, chunk)

		utils.AssertErrorMessage(t, err, "expected 2 arguments but got 1")
	})

	t.Run("Error si se llama a una función con más argumentos de los esperados", func(t *testing.T) {
		vm, chunk := setUpVM(t)
		callee := newFunction(1, func(body *c.Chunk) {})
		writeConstant(chunk, callee)
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
				writeConstant(body, next)
				writeCall(body, 0)
			})
		}
		return callee
	}

	t.Run("Se permiten llamadas anidadas hasta la profundidad máxima", func(t *testing.T) {
		vm, chunk := setUpVM(t)
		writeConstant(chunk, nestedCalls(framesMax-1))
		writeCall(chunk, 0)

		err := run(vm, chunk)

		utils.AssertNoError(t, err)
	})

	t.Run("Error de stack overflow si las llamadas anidadas superan la profundidad máxima", func(t *testing.T) {
		vm, chunk := setUpVM(t)
		writeConstant(chunk, nestedCalls(framesMax))
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
		writeConstant(chunk, returning(5.0))
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
		writeConstant(chunk, suma)
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
			writeConstant(chunk, callee)
			writeCall(chunk, 0)
			chunk.Write(byte(c.OP_POP), line)
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
		writeConstant(&outer.Chunk, returning(5.0))
		writeCall(&outer.Chunk, 0)
		outer.Chunk.Write(byte(c.OP_RETURN), line)
		writeConstant(chunk, outer)
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
		writeConstant(chunk, f)
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
