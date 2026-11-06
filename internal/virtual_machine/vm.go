package virtualmachine

import (
	c "tp-lox/internal/chunk"
	"tp-lox/internal/errors"
	"tp-lox/internal/utils"
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
			num, err := vm.popNumberOperand()
			if err != nil {
				return err
			}
			vm.push(-num)

		case c.OP_ADD:
			if err := vm.binaryOp(func(a, b float64) any { return b + a }); err != nil {
				return err
			}
		case c.OP_SUBTRACT:
			if err := vm.binaryOp(func(a, b float64) any { return b - a }); err != nil {
				return err
			}
		case c.OP_MULTIPLY:
			if err := vm.binaryOp(func(a, b float64) any { return b * a }); err != nil {
				return err
			}
		case c.OP_DIVIDE:
			if err := vm.binaryOp(func(a, b float64) any { return b / a }); err != nil {
				return err
			}
		case c.OP_GREATER:
			if err := vm.binaryOp(func(a, b float64) any { return b > a }); err != nil {
				return err
			}
		case c.OP_LESS:
			if err := vm.binaryOp(func(a, b float64) any { return b < a }); err != nil {
				return err
			}
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

// peek devuelve el valor a distance posiciones del tope del stack, sin sacarlo
// func (vm *VirtualMachine) peek(distance int) any {
// 	if len(vm.stack) <= distance {
// 		return nil
// 	}
// 	return vm.stack[len(vm.stack)-1-distance]
// }

// popNumberOperand saca el tope del stack y verifica que sea un número
func (vm *VirtualMachine) popNumberOperand() (float64, error) {
	operand, err := vm.pop()
	if err != nil {
		return 0, err
	}
	return utils.CheckNumberOperand(0, operand)
}

func (vm *VirtualMachine) popNumberOperands() (float64, float64, error) {
	a, err := vm.popNumberOperand()
	if err != nil {
		return 0, 0, err
	}
	b, err := vm.popNumberOperand()
	if err != nil {
		return 0, 0, err
	}
	return a, b, nil
}

// binaryOp saca dos números del stack, les aplica op y deja el resultado en el stack
func (vm *VirtualMachine) binaryOp(op func(a, b float64) any) error {
	a, b, err := vm.popNumberOperands()
	if err != nil {
		return err
	}
	vm.push(op(a, b))
	return nil
}
