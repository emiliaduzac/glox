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

const (
	framesMax = 64
	stackMax  = framesMax * (math.MaxUint8 + 1)
)

type callFrame struct {
	closure *closure
	ip      int
	slots   int
}

type VirtualMachine struct {
	frames       []callFrame
	stack        []any
	out          io.Writer
	globals      map[string]any
	debug        bool
	openUpvalues *upvalue
}

func NewVirtualMachine() *VirtualMachine {
	globals := map[string]any{}
	for name, function := range builtinNatives() {
		globals[name] = function
	}
	return &VirtualMachine{
		frames:  make([]callFrame, 0, framesMax),
		stack:   make([]any, 0, stackMax),
		out:     os.Stdout,
		globals: globals,
	}
}

func (vm *VirtualMachine) SetDebugMode() {
	vm.debug = true
}

func (vm *VirtualMachine) Run(function *c.Function) (err error) {
	script := newClosure(function)
	vm.push(script)
	defer func() {
		if err != nil {
			err = vm.withStackTrace(err)
			vm.resetStack()
		}
	}()
	if err := vm.call(script, 0); err != nil {
		return err
	}

	if vm.debug {
		_, _ = fmt.Fprintln(vm.out, "\n== RUNTIME ==")
	}

	frame := vm.currentFrame()
	for frame.ip < len(frame.closure.function.Chunk.Bytes) {
		start := frame.ip
		b, err := vm.readByte()
		if err != nil {
			return err
		}
		op := c.OpCode(b)

		switch op {
		case c.OP_RETURN:
			result, err := vm.pop()
			if err != nil {
				return err
			}
			if len(vm.frames) == 1 && len(vm.stack) > frame.slots+1 {
				return errors.NewLoxError(vm.currentLine(), "inconsistent Stack Height: should be empty at exit")
			}
			vm.closeUpvalues(frame.slots)
			vm.stack = vm.stack[:frame.slots]
			vm.frames = vm.frames[:len(vm.frames)-1]
			if len(vm.frames) == 0 {
				if vm.debug {
					_, _ = fmt.Fprintf(vm.out, "%04d| THE END\n", start)
				}
				return nil
			}
			vm.push(result)
			frame = vm.currentFrame()

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
			vm.push(vm.stack[frame.slots+int(slot)])
		case c.OP_SET_LOCAL:
			slot, err := vm.readByte()
			if err != nil {
				return err
			}
			value, err := vm.peek(0)
			if err != nil {
				return err
			}
			vm.stack[frame.slots+int(slot)] = value

		// UpValues
		case c.OP_GET_UPVALUE:
			slot, err := vm.readByte()
			if err != nil {
				return err
			}
			vm.push(*frame.closure.upvalues[slot].location)
		case c.OP_SET_UPVALUE:
			slot, err := vm.readByte()
			if err != nil {
				return err
			}
			value, err := vm.peek(0)
			if err != nil {
				return err
			}
			*frame.closure.upvalues[slot].location = value
		case c.OP_CLOSE_UPVALUE:
			vm.closeUpvalues(len(vm.stack) - 1)
			_, err := vm.pop()
			if err != nil {
				return err
			}

		case c.OP_CALL:
			argCount, err := vm.readByte()
			if err != nil {
				return err
			}
			callee, err := vm.peek(int(argCount))
			if err != nil {
				return err
			}
			if err := vm.callValue(callee, int(argCount)); err != nil {
				return err
			}
			frame = vm.currentFrame()

		case c.OP_CLOSURE:
			constant, err := vm.readConstant()
			if err != nil {
				return err
			}
			function := constant.(*c.Function)
			cl := newClosure(function)
			if err := vm.fillUpvalues(cl, frame); err != nil {
				return err
			}
			vm.push(cl)

		case c.OP_CLASS:
			name, err := vm.readString()
			if err != nil {
				return err
			}
			vm.push(newClass(name))

		case c.OP_METHOD:
			name, err := vm.readString()
			if err != nil {
				return err
			}
			if err := vm.defineMethod(name); err != nil {
				return err
			}

		case c.OP_GET_PROPERTY:
			object, err := vm.peek(0)
			if err != nil {
				return err
			}
			instance, ok := object.(*instance)
			if !ok {
				return errors.NewLoxError(vm.currentLine(), "only instances have properties")
			}
			name, err := vm.readString()
			if err != nil {
				return err
			}
			if value, ok := instance.fields[name]; ok {
				if _, err := vm.pop(); err != nil {
					return err
				}
				vm.push(value)
			} else if err := vm.bindMethod(instance.class, name); err != nil {
				return err
			}

		case c.OP_SET_PROPERTY:
			object, err := vm.peek(1)
			if err != nil {
				return err
			}
			instance, ok := object.(*instance)
			if !ok {
				return errors.NewLoxError(vm.currentLine(), "only instances have fields")
			}
			name, err := vm.readString()
			if err != nil {
				return err
			}
			value, err := vm.peek(0)
			if err != nil {
				return err
			}
			instance.fields[name] = value
			if err := vm.popTwo(); err != nil {
				return err
			}
			vm.push(value)

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
				frame.ip += offset
			}
		case c.OP_JUMP:
			offset, err := vm.readShort()
			if err != nil {
				return err
			}
			frame.ip += offset
		case c.OP_LOOP:
			offset, err := vm.readShort()
			if err != nil {
				return err
			}
			frame.ip -= offset

		default:
			return errors.NewLoxError(vm.currentLine(), fmt.Sprintf("unknown opcode %s", op))
		}

		if vm.debug {
			vm.printStackAndGlobal(start)
		}
	}
	return nil
}

func (vm *VirtualMachine) callValue(callee any, argCount int) error {
	switch function := callee.(type) {
	case *closure:
		return vm.call(function, argCount)
	case *native:
		return vm.callNative(function, argCount)
	case *class:
		vm.stack[len(vm.stack)-argCount-1] = newInstance(function)
		return nil
	default:
		return errors.NewLoxError(vm.currentLine(), "can only call functions and classes")
	}
}

func (vm *VirtualMachine) bindMethod(klass *class, name string) error {
	method, ok := klass.methods[name]
	if !ok {
		return errors.NewLoxError(vm.currentLine(), fmt.Sprintf("undefined property '%s'", name))
	}
	receiver, err := vm.peek(0)
	if err != nil {
		return err
	}
	bound := newBoundMethod(receiver, method)
	if _, err := vm.pop(); err != nil {
		return err
	}
	vm.push(bound)
	return nil
}

func (vm *VirtualMachine) call(cl *closure, argCount int) error {
	if argCount != cl.function.Arity {
		return errors.NewLoxError(vm.currentLine(), fmt.Sprintf("expected %d arguments but got %d", cl.function.Arity, argCount))
	}
	if len(vm.frames) == framesMax {
		return errors.NewLoxError(vm.currentLine(), "stack overflow")
	}
	vm.frames = append(vm.frames, callFrame{closure: cl, slots: len(vm.stack) - argCount - 1})
	return nil
}

func (vm *VirtualMachine) callNative(function *native, argCount int) error {
	if argCount != function.arity {
		return errors.NewLoxError(vm.currentLine(), fmt.Sprintf("expected %d arguments but got %d", function.arity, argCount))
	}
	args := vm.stack[len(vm.stack)-argCount:]
	result := function.fn(args)
	vm.stack = vm.stack[:len(vm.stack)-argCount-1]
	vm.push(result)
	return nil
}

func (vm *VirtualMachine) withStackTrace(err error) error {
	if len(vm.frames) == 0 {
		return err
	}
	var trace strings.Builder
	for i := len(vm.frames) - 1; i >= 0; i-- {
		frame := vm.frames[i]
		name := "script"
		function := frame.closure.function
		if function.Name != "" {
			name = function.Name + "()"
		}
		fmt.Fprintf(&trace, "\n[line %d] in %s", function.Chunk.Lines[frame.ip-1], name)
	}
	return fmt.Errorf("%w%s", err, trace.String())
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
	vm.frames = vm.frames[:0]
	vm.openUpvalues = nil
}

func (vm *VirtualMachine) stackBase() int {
	return vm.currentFrame().slots + 1
}

func (vm *VirtualMachine) currentFrame() *callFrame {
	return &vm.frames[len(vm.frames)-1]
}

// readByte devuelve el byte al que apunta el instruction pointer y lo avanza al siguiente
func (vm *VirtualMachine) readByte() (byte, error) {
	frame := vm.currentFrame()
	function := frame.closure.function
	if frame.ip >= len(function.Chunk.Bytes) {
		return 0, errors.NewLoxError(vm.currentLine(), "unexpected end of bytecode")
	}
	b := function.Chunk.Bytes[frame.ip]
	frame.ip++
	return b, nil
}

// readConstant lee el operando y devuelve la constante del pool en ese índice
func (vm *VirtualMachine) readConstant() (any, error) {
	index, err := vm.readByte()
	if err != nil {
		return nil, err
	}
	constants := vm.currentFrame().closure.function.Chunk.Constants
	if int(index) >= len(constants) {
		return nil, errors.NewLoxError(vm.currentLine(), "constant index out of range")
	}
	return constants[index], nil
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
	if len(vm.stack) <= vm.stackBase() {
		return nil, errors.NewLoxError(vm.currentLine(), "stack underflow")
	}

	top := vm.stack[len(vm.stack)-1]
	vm.stack = vm.stack[:len(vm.stack)-1]
	return top, nil
}

// peek devuelve el valor a distance posiciones del tope del stack, sin sacarlo
func (vm *VirtualMachine) peek(distance int) (any, error) {
	if len(vm.stack)-vm.stackBase() <= distance {
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

// popTwo saca los dos valores del tope del stack
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
	frame := vm.currentFrame()
	return frame.closure.function.Chunk.Lines[frame.ip-1]
}

func (vm *VirtualMachine) fillUpvalues(cl *closure, frame *callFrame) error {
	for i := range cl.upvalues {
		isLocal, err := vm.readByte()
		if err != nil {
			return err
		}
		index, err := vm.readByte()
		if err != nil {
			return err
		}
		if isLocal == 1 {
			cl.upvalues[i] = vm.captureUpvalue(frame.slots + int(index))
		} else {
			cl.upvalues[i] = frame.closure.upvalues[index]
		}
	}
	return nil
}

func (vm *VirtualMachine) captureUpvalue(stackIndex int) *upvalue {
	var prev *upvalue
	current := vm.openUpvalues
	for current != nil && current.stackIndex > stackIndex {
		prev = current
		current = current.next
	}

	if current != nil && current.stackIndex == stackIndex {
		return current
	}

	created := newUpvalue(&vm.stack[stackIndex], stackIndex)
	created.next = current
	if prev == nil {
		vm.openUpvalues = created
	} else {
		prev.next = created
	}
	return created
}

func (vm *VirtualMachine) closeUpvalues(last int) {
	for vm.openUpvalues != nil && vm.openUpvalues.stackIndex >= last {
		up := vm.openUpvalues
		up.closed = *up.location
		up.location = &up.closed
		vm.openUpvalues = up.next
	}
}

func (vm *VirtualMachine) defineMethod(name string) error {
	method, err := vm.peek(0)
	if err != nil {
		return err
	}
	klass, err := vm.peek(1)
	if err != nil {
		return err
	}
	klass.(*class).methods[name] = method.(*closure)
	_, err = vm.pop()
	return err
}
