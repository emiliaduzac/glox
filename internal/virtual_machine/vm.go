package virtualmachine

import (
	"fmt"
	"io"
	"math"
	"os"
	"sort"
	"strings"
	c "tp-lox/internal/chunk"
	"tp-lox/internal/errors"
	"tp-lox/internal/utils"
)

type VirtualMachine struct {
	ip      int
	stack   []any
	chunk   *c.Chunk
	out     io.Writer
	globals map[string]any
	debug   bool
}

func NewVirtualMachine() *VirtualMachine {
	return &VirtualMachine{
		stack:   []any{},
		out:     os.Stdout,
		globals: map[string]any{},
		debug:   false,
	}
}

func (vm *VirtualMachine) SetDebugMode() {
	vm.debug = true
}

func (vm *VirtualMachine) Run(chunk *c.Chunk) (err error) {
	vm.chunk = chunk
	vm.ip = 0
	defer func() {
		if err != nil {
			vm.resetStack()
		}
	}()

	if vm.debug {
		_, _ = fmt.Fprintln(vm.out, "\n== RUNTIME ==")
	}

	for vm.ip < len(vm.chunk.Bytes) {
		start := vm.ip
		b, err := vm.readByte()
		if err != nil {
			return err
		}
		op := c.OpCode(b)

		switch op {
		case c.OP_RETURN:
			if len(vm.stack) > 0 {
				return errors.NewLoxError(vm.currentLine(), "inconsistent Stack Height: should be empty at exit")
			}
			if vm.debug {
				_, _ = fmt.Fprintf(vm.out, "%04d| THE END\n", start)
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

		// Aritméticos
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
		case c.OP_MODULO:
			if err := vm.binaryOp(func(left, right float64) any { return math.Mod(left, right) }); err != nil {
				return err
			}

		// Comparación
		case c.OP_EQUAL:
			left, right, err := vm.peekOperands()
			if err != nil {
				return err
			}
			if err := vm.popTwo(); err != nil {
				return err
			}
			vm.push(utils.IsEqual(left, right))
		case c.OP_GREATER:
			if err := vm.binaryOp(func(left, right float64) any { return left > right }); err != nil {
				return err
			}
		case c.OP_LESS:
			if err := vm.binaryOp(func(left, right float64) any { return left < right }); err != nil {
				return err
			}

		case c.OP_NOT:
			value, err := vm.pop()
			if err != nil {
				return err
			}
			vm.push(!utils.IsTruthy(value))

		case c.OP_PRINT:
			value, err := vm.pop()
			if err != nil {
				return err
			}
			if vm.debug {
				if _, err := fmt.Fprintf(vm.out, "%04d| SCREEN OUTPUT %s\n", start, utils.Stringify(value)); err != nil {
					return err
				}
				continue
			}
			if _, err := fmt.Fprintln(vm.out, utils.Stringify(value)); err != nil {
				return err
			}

		case c.OP_POP:
			if _, err := vm.pop(); err != nil {
				return err
			}

		// Variables globales
		case c.OP_DEFINE_GLOBAL:
			name, err := vm.readString()
			if err != nil {
				return err
			}
			value, err := vm.pop()
			if err != nil {
				return err
			}
			vm.globals[name] = value
		case c.OP_GET_GLOBAL:
			name, err := vm.readString()
			if err != nil {
				return err
			}
			value, ok := vm.globals[name]
			if !ok {
				return errors.NewLoxError(vm.currentLine(), fmt.Sprintf("undefined variable '%s'", name))
			}
			vm.push(value)
		case c.OP_SET_GLOBAL:
			name, err := vm.readString()
			if err != nil {
				return err
			}
			if _, ok := vm.globals[name]; !ok {
				return errors.NewLoxError(vm.currentLine(), fmt.Sprintf("undefined variable '%s'", name))
			}
			value, err := vm.peek(0)
			if err != nil {
				return err
			}
			vm.globals[name] = value

		// Variables locales
		case c.OP_GET_LOCAL:
			slot, err := vm.readByte()
			if err != nil {
				return err
			}
			vm.push(vm.stack[slot])
		case c.OP_SET_LOCAL:
			slot, err := vm.readByte()
			if err != nil {
				return err
			}
			value, err := vm.peek(0)
			if err != nil {
				return err
			}
			vm.stack[slot] = value

		// Control de flujo
		case c.OP_JUMP_IF_FALSE:
			offset, err := vm.readShort()
			if err != nil {
				return err
			}
			value, err := vm.peek(0)
			if err != nil {
				return err
			}
			if !utils.IsTruthy(value) {
				vm.ip += offset
			}
		case c.OP_JUMP:
			offset, err := vm.readShort()
			if err != nil {
				return err
			}
			vm.ip += offset

		case c.OP_LOOP:
			offset, err := vm.readShort()
			if err != nil {
				return err
			}
			vm.ip -= offset

		default:
			return errors.NewLoxError(vm.currentLine(), fmt.Sprintf("unknown opcode %s", op))
		}

		if vm.debug {
			vm.printStackAndGlobal(start)
		}
	}
	return nil
}

func (vm *VirtualMachine) printStackAndGlobal(offset int) {
	values := make([]string, len(vm.stack))
	for i, v := range vm.stack {
		values[i] = utils.DebugString(v)
	}
	_, _ = fmt.Fprintf(vm.out, "%04d| STACK [%s]\n", offset, strings.Join(values, ", "))

	if len(vm.globals) == 0 {
		return
	}
	names := make([]string, 0, len(vm.globals))
	for name := range vm.globals {
		names = append(names, name)
	}
	sort.Strings(names)
	entries := make([]string, len(names))
	for i, name := range names {
		entries[i] = fmt.Sprintf("'%s': %s", name, utils.DebugString(vm.globals[name]))
	}
	_, _ = fmt.Fprintf(vm.out, "    | GLOBALS {%s}\n", strings.Join(entries, ", "))
}

// resetStack descarta todos los valores del stack
func (vm *VirtualMachine) resetStack() {
	vm.stack = vm.stack[:0]
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

func (vm *VirtualMachine) readString() (string, error) {
	constant, err := vm.readConstant()
	if err != nil {
		return "", err
	}
	name, ok := constant.(string)
	if !ok {
		return "", errors.NewLoxError(vm.currentLine(), "constant is not a string")
	}
	return name, nil
}

func (vm *VirtualMachine) readShort() (int, error) {
	high, err := vm.readByte()
	if err != nil {
		return 0, err
	}
	low, err := vm.readByte()
	if err != nil {
		return 0, err
	}
	return int(high)<<8 | int(low), nil
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
func (vm *VirtualMachine) peek(distance int) (any, error) {
	if len(vm.stack) <= distance {
		return nil, errors.NewLoxError(vm.currentLine(), "stack underflow")
	}
	return vm.stack[len(vm.stack)-1-distance], nil
}

// peekOperands devuelve los dos operandos del tope, sin sacarlos y en orden (izquierdo, derecho)
func (vm *VirtualMachine) peekOperands() (any, any, error) {
	right, err := vm.peek(0)
	if err != nil {
		return nil, nil, err
	}
	left, err := vm.peek(1)
	if err != nil {
		return nil, nil, err
	}
	return left, right, nil
}

// popNumberOperand verifica que el tope del stack sea un número. Si lo es, lo saca
func (vm *VirtualMachine) popNumberOperand() (float64, error) {
	operand, err := vm.peek(0)
	if err != nil {
		return 0, err
	}
	num, err := utils.CheckNumberOperand(vm.currentLine(), operand)
	if err != nil {
		return 0, err
	}
	if _, err := vm.pop(); err != nil {
		return 0, err
	}
	return num, nil
}

// popNumberOperands verifica que los dos operandos sean números. Si lo son, los saca
func (vm *VirtualMachine) popNumberOperands() (float64, float64, error) {
	left, right, err := vm.peekOperands()
	if err != nil {
		return 0, 0, err
	}
	l, r, err := utils.CheckNumberOperands(vm.currentLine(), left, right)
	if err != nil {
		return 0, 0, err
	}
	if err := vm.popTwo(); err != nil {
		return 0, 0, err
	}
	return l, r, nil
}

// popTwo saca los dos operandos del tope del stack
func (vm *VirtualMachine) popTwo() error {
	if _, err := vm.pop(); err != nil {
		return err
	}
	_, err := vm.pop()
	return err
}

// add suma dos números o concatena dos strings y deja el resultado en el stack
func (vm *VirtualMachine) add() error {
	left, right, err := vm.peekOperands()
	if err != nil {
		return err
	}

	var result any
	switch l := left.(type) {
	case float64:
		if r, ok := right.(float64); ok {
			result = l + r
		}
	case string:
		if r, ok := right.(string); ok {
			result = l + r
		}
	}
	if result == nil {
		return errors.NewLoxError(vm.currentLine(), "operands must be two numbers or two strings")
	}

	if err := vm.popTwo(); err != nil {
		return err
	}
	vm.push(result)
	return nil
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
