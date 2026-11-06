package virtualmachine

import (
	c "tp-lox/internal/chunk"
	"tp-lox/internal/errors"
)

type VirtualMachine struct {
	ip    int
	stack []any
}

// Push agrega un elemento al tope del stack
func (vm *VirtualMachine) Push(value any) {
	vm.stack = append(vm.stack, value)
}

// Pop retorna el elemento del tope del stack
func (vm *VirtualMachine) Pop() (any, error) {
	if len(vm.stack) <= 0 {
		return nil, errors.NewLoxError(0, "STACK UNDERFLOW")
	}

	return vm.stack[len(vm.stack)-1], nil
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
			index := vm.readByte(chunk)
			constant := chunk.Constants[index]
			vm.Push(constant)

		case c.OP_NIL:
			vm.Push(nil)
		case c.OP_TRUE:
			vm.Push(true)
		case c.OP_FALSE:
			vm.Push(false)
		}
	}
	return nil
}

func (vm *VirtualMachine) readByte(chunk c.Chunk) byte {
	b := chunk.Bytes[vm.ip]
	vm.ip++
	return b
}
