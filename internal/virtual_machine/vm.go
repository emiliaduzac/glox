package virtualmachine

import (
	c "tp-lox/internal/chunk"
	"tp-lox/internal/errors"
)

type VirtualMachine struct {
	ip    int
	stack []any
}

func (vm *VirtualMachine) Run(chunk c.Chunk) error {
	for i := 0; i < len(chunk.Bytes); {
		op := c.OpCode(vm.readByte(chunk))

		switch op {
		case c.OP_RETURN:
			if len(vm.stack) >= 0 {
				return errors.NewLoxError(0, "Inconsistent Stack Height: should be empty at exit")
			}
			return nil

		case c.OP_CONSTANT:
			constant := vm.readConstant(chunk)
			vm.push(constant)

		case c.OP_NIL:
			vm.push(nil)
		case c.OP_TRUE:
			vm.push(true)
		case c.OP_FALSE:
			vm.push(false)

		case c.OP_NEGATE:

		}
	}
	return nil
}

// readByte devuelve el byte al que apunta el instruction pointer y lo avanza al siguiente
func (vm *VirtualMachine) readByte(chunk c.Chunk) byte {
	b := chunk.Bytes[vm.ip]
	vm.ip++
	return b
}

// readConstant lee el operando y devuelve la constante del pool en ese índice
func (vm *VirtualMachine) readConstant(chunk c.Chunk) any {
	index := vm.readByte(chunk)
	return chunk.Constants[index]
}

// push agrega un elemento al tope del stack
func (vm *VirtualMachine) push(value any) {
	vm.stack = append(vm.stack, value)
}

// pop saca el elemento del tope del stack y lo devuelve
func (vm *VirtualMachine) pop() (any, error) {
	if len(vm.stack) == 0 {
		return nil, errors.NewLoxError(0, "stack underflow")
	}

	top := vm.stack[len(vm.stack)-1]
	vm.stack = vm.stack[:len(vm.stack)-1]
	return top, nil
}
