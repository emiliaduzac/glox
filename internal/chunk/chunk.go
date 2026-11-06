package chunk

import (
	"fmt"
)

type OpCode int

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
}

func NewChunk() *Chunk {
	return &Chunk{}
}

func (c *Chunk) Write(b byte) {
	c.Bytes = append(c.Bytes, b)
}

func (c *Chunk) AddConstant(value any) int {
	c.Constants = append(c.Constants, value)
	return len(c.Constants) - 1
}
