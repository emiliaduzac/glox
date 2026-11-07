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
	return c.chunk, nil
}
