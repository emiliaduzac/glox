package virtualmachine

import (
	"fmt"
	c "tp-lox/internal/chunk"
	"tp-lox/internal/errors"
	"tp-lox/internal/utils"
)

type VirtualMachine struct {
	ip    int
	stack []any
	chunk *c.Chunk
}

func NewVirtualMachine(chunk *c.Chunk) *VirtualMachine {
	return &VirtualMachine{
		ip:    0,
		stack: []any{},
		chunk: chunk,
	}
}

func (vm *VirtualMachine) Run() error {
	for vm.ip < len(vm.chunk.Bytes) {
		b, err := vm.readByte()
		if err != nil {
			return err
		}
		op := c.OpCode(b)

		switch op {
		case c.OP_RETURN:
			if len(vm.stack) > 0 {
				return errors.NewLoxError(vm.currentLine(), "Inconsistent Stack Height: should be empty at exit")
			}
			return nil

		case c.OP_CONSTANT:
			constant, err := vm.readConstant()
			if err != nil {
				return err
			}
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
			if err := vm.add(); err != nil {
				return err
			}
		case c.OP_SUBTRACT:
			if err := vm.binaryOp(func(left, right float64) any { return left - right }); err != nil {
				return err
			}
		case c.OP_MULTIPLY:
			if err := vm.binaryOp(func(left, right float64) any { return left * right }); err != nil {
				return err
			}
		case c.OP_DIVIDE:
			if err := vm.binaryOp(func(left, right float64) any { return left / right }); err != nil {
				return err
			}
		case c.OP_GREATER:
			if err := vm.binaryOp(func(left, right float64) any { return left > right }); err != nil {
				return err
			}
		case c.OP_LESS:
			if err := vm.binaryOp(func(left, right float64) any { return left < right }); err != nil {
				return err
			}

		default:
			return errors.NewLoxError(vm.currentLine(), fmt.Sprintf("UNKOWN %b", op))
		}
	}
	return nil
}

// readByte devuelve el byte al que apunta el instruction pointer y lo avanza al siguiente
func (vm *VirtualMachine) readByte() (byte, error) {
	if vm.ip >= len(vm.chunk.Bytes) {
		return 0, errors.NewLoxError(vm.currentLine(), "unexpected end of bytecode")
	}
	b := vm.chunk.Bytes[vm.ip]
	vm.ip++
	return b, nil
}

// readConstant lee el operando y devuelve la constante del pool en ese índice
func (vm *VirtualMachine) readConstant() (any, error) {
	index, err := vm.readByte()
	if err != nil {
		return nil, err
	}
	if int(index) >= len(vm.chunk.Constants) {
		return nil, errors.NewLoxError(vm.currentLine(), "constant index out of range")
	}
	return vm.chunk.Constants[index], nil
}

// push agrega un elemento al tope del stack
func (vm *VirtualMachine) push(value any) {
	vm.stack = append(vm.stack, value)
}

// pop saca el elemento del tope del stack y lo devuelve
func (vm *VirtualMachine) pop() (any, error) {
	if len(vm.stack) == 0 {
		return nil, errors.NewLoxError(vm.currentLine(), "stack underflow")
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
	return utils.CheckNumberOperand(vm.currentLine(), operand)
}

// popNumberOperands saca dos números del stack y los devuelve en orden (izquierdo, derecho)
func (vm *VirtualMachine) popNumberOperands() (float64, float64, error) {
	left, right, err := vm.popOperands()
	if err != nil {
		return 0, 0, err
	}
	return utils.CheckNumberOperands(vm.currentLine(), left, right)
}

// popOperands saca los dos operandos del stack y los devuelve en orden (izquierdo, derecho)
func (vm *VirtualMachine) popOperands() (any, any, error) {
	right, err := vm.pop()
	if err != nil {
		return nil, nil, err
	}
	left, err := vm.pop()
	if err != nil {
		return nil, nil, err
	}
	return left, right, nil
}

// add suma dos números o concatena dos strings y deja el resultado en el stack
func (vm *VirtualMachine) add() error {
	left, right, err := vm.popOperands()
	if err != nil {
		return err
	}

	switch l := left.(type) {
	case float64:
		if r, ok := right.(float64); ok {
			vm.push(l + r)
			return nil
		}
	case string:
		if r, ok := right.(string); ok {
			vm.push(l + r)
			return nil
		}
	}
	return errors.NewLoxError(vm.currentLine(), "operands must be two numbers or two strings")
}

// binaryOp saca dos números del stack, les aplica op y deja el resultado en el stack
func (vm *VirtualMachine) binaryOp(op func(left, right float64) any) error {
	left, right, err := vm.popNumberOperands()
	if err != nil {
		return err
	}
	vm.push(op(left, right))
	return nil
}

func (vm *VirtualMachine) currentLine() int {
	return vm.chunk.Lines[vm.ip-1]
}
