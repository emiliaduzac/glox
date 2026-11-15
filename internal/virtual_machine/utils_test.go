package virtualmachine

import (
	"testing"
	c "tp-lox/internal/chunk"
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

func writeClosure(chunk *c.Chunk, function *c.Function, upvalues ...byte) {
	index := chunk.AddConstant(function)
	chunk.Write(byte(c.OP_CLOSURE), line)
	chunk.Write(byte(index), line)
	for _, b := range upvalues {
		chunk.Write(b, line)
	}
}

func writeGetGlobal(chunk *c.Chunk, name string) {
	index := chunk.AddConstant(name)
	chunk.Write(byte(c.OP_GET_GLOBAL), line)
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

func writeDefineGlobal(chunk *c.Chunk, name string) {
	index := chunk.AddConstant(name)
	chunk.Write(byte(c.OP_DEFINE_GLOBAL), line)
	chunk.Write(byte(index), line)
}

func writeCallStatement(chunk *c.Chunk) {
	writeCall(chunk, 0) // f();
	chunk.Write(byte(c.OP_POP), line)
}

func writeCallGlobal(chunk *c.Chunk, name string) {
	writeGetGlobal(chunk, name) // name();
	writeCallStatement(chunk)
}

func writeReturnNil(chunk *c.Chunk) {
	chunk.Write(byte(c.OP_NIL), line)
	chunk.Write(byte(c.OP_RETURN), line)
}

func newPrintUpvalue() *c.Function {
	// fun mostrar() { print x; }
	mostrar := c.NewFunction()
	mostrar.UpvalueCount = 1
	mostrar.Chunk.Write(byte(c.OP_GET_UPVALUE), line)
	mostrar.Chunk.Write(0, line)
	mostrar.Chunk.Write(byte(c.OP_PRINT), line)
	writeReturnNil(&mostrar.Chunk)
	return mostrar
}
