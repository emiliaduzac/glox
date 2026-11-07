package compiler

import (
	"tp-lox/internal/chunk"
	"tp-lox/internal/token"
)

type Compiler struct {
	cursor *token.TokenCursor
	chunk  *chunk.Chunk
}

// NewCompiler crea un nuevo compiler, a partir de una lista de tokens generados por el scanner.
func NewCompiler(tokens []token.Token) *Compiler {
	return &Compiler{
		cursor: token.NewTokenCursor(tokens),
		chunk:  chunk.NewChunk(),
	}
}

func (c *Compiler) Compile() (*chunk.Chunk, error) {
	for !c.cursor.IsAtEnd() {

	}
	c.emitReturn()
	return c.chunk, nil
}

// emit escribe uno o más bytes en el chunk, con la línea del último token consumido
func (c *Compiler) emit(bytes ...byte) {
	for _, b := range bytes {
		c.chunk.Write(b, c.cursor.GetPreviousLine())
	}
}

func (c *Compiler) emitReturn() {
	c.chunk.Write(byte(chunk.OP_RETURN), c.cursor.GetPreviousLine())
}
