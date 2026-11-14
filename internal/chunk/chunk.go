package chunk

import (
	"fmt"
	"strings"
)

type OpCode byte

const (
	OP_RETURN OpCode = iota
	OP_NIL
	OP_TRUE
	OP_FALSE
	OP_PRINT
	OP_POP

	OP_NEGATE
	OP_ADD
	OP_SUBTRACT
	OP_MULTIPLY
	OP_DIVIDE
	OP_MODULO
	OP_NOT
	OP_EQUAL
	OP_GREATER
	OP_LESS
	OP_PRINT_DEC

	OP_CONSTANT
	OP_DEFINE_GLOBAL
	OP_GET_GLOBAL
	OP_SET_GLOBAL

	OP_GET_LOCAL
	OP_SET_LOCAL

	OP_LOOP
	OP_JUMP_IF_FALSE
	OP_JUMP
)

var opCodeNames = map[OpCode]string{
	OP_RETURN: "OP_RETURN",
	OP_NIL:    "OP_NIL",
	OP_TRUE:   "OP_TRUE",
	OP_FALSE:  "OP_FALSE",
	OP_PRINT:  "OP_PRINT",
	OP_POP:    "OP_POP",

	OP_NEGATE:    "OP_NEGATE",
	OP_ADD:       "OP_ADD",
	OP_SUBTRACT:  "OP_SUBTRACT",
	OP_MULTIPLY:  "OP_MULTIPLY",
	OP_DIVIDE:    "OP_DIVIDE",
	OP_MODULO:    "OP_MODULO",
	OP_NOT:       "OP_NOT",
	OP_EQUAL:     "OP_EQUAL",
	OP_GREATER:   "OP_GREATER",
	OP_LESS:      "OP_LESS",
	OP_PRINT_DEC: "OP_PRINT_DEC",

	OP_CONSTANT:      "OP_CONSTANT",
	OP_DEFINE_GLOBAL: "OP_DEFINE_GLOBAL",
	OP_GET_GLOBAL:    "OP_GET_GLOBAL",
	OP_SET_GLOBAL:    "OP_SET_GLOBAL",

	OP_GET_LOCAL: "OP_GET_LOCAL",
	OP_SET_LOCAL: "OP_SET_LOCAL",

	OP_LOOP:          "OP_LOOP",
	OP_JUMP_IF_FALSE: "OP_JUMP_IF_FALSE",
	OP_JUMP:          "OP_JUMP",
}

func (op OpCode) String() string {
	if name, ok := opCodeNames[op]; ok {
		return name
	}
	return fmt.Sprintf("OP_UNKNOWN<%d>", byte(op))
}

type Chunk struct {
	Bytes     []byte // secuencia de instrucciones
	Constants []any  // pool de constantes
	Lines     []int
}

func NewChunk() *Chunk {
	return &Chunk{}
}

func (c *Chunk) Write(b byte, line int) {
	c.Lines = append(c.Lines, line)
	c.Bytes = append(c.Bytes, b)
}

func (c *Chunk) AddConstant(value any) int {
	c.Constants = append(c.Constants, value)
	return len(c.Constants) - 1
}

func (c *Chunk) Disassemble(name string) {
	fmt.Printf("== %s ==\n", name)
	fmt.Println("== RAW BYTES ==")
	var parts []string
	for _, b := range c.Bytes {
		parts = append(parts, fmt.Sprintf("%08b", b))
	}
	fmt.Println(strings.Join(parts, " "))
	fmt.Println()

	fmt.Println("== COMPILETIME ==")
	for i := 0; i < len(c.Bytes); {
		op := OpCode(c.Bytes[i])
		fmt.Printf("%04d| %4d ", i, c.Lines[i])

		switch op {
		case OP_CONSTANT, OP_DEFINE_GLOBAL, OP_GET_GLOBAL, OP_SET_GLOBAL:
			// Instrucciones con operandos: constantes
			constantIndex := c.Bytes[i+1]
			fmt.Printf("%s<%s>\n", op, c.constantString(constantIndex))
			i += 2

		case OP_GET_LOCAL, OP_SET_LOCAL:
			// Instrucciones con operandos: slots
			varIndex := c.Bytes[i+1]
			fmt.Printf("%s<@%d>\n", op, varIndex)
			i += 2

		case OP_JUMP, OP_JUMP_IF_FALSE, OP_LOOP:
			// Instrucciones con operandos: saltos
			highByte, lowByte := int(c.Bytes[i+1]), int(c.Bytes[i+2])
			offset := highByte<<8 | lowByte
			if op == OP_LOOP {
				offset = -offset
			}
			jump := offset + 3
			fmt.Printf("%s<%+d>\n", op, jump)
			i += 3

		default:
			// Instrucciones sin operandos
			fmt.Printf("%s\n", op)
			i++
		}
	}

	fmt.Printf("CONST %v\n", c.Constants)
}

// constantString muestra una constante del pool, con comillas si es string
func (c *Chunk) constantString(index byte) string {
	if int(index) >= len(c.Constants) {
		return fmt.Sprintf("índice %d inválido", index)
	}
	if s, ok := c.Constants[index].(string); ok {
		return fmt.Sprintf("'%s'", s)
	}
	return fmt.Sprintf("%v", c.Constants[index])
}
