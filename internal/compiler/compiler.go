package compiler

import "go/token"

type Compiler struct {
	tokens  []token.Token
	current int
}

// NewCompiler crea un nuevo compiler, a partir de una lista de tokens generados por el scanner.
func NewCompiler(tokens []token.Token) *Compiler {
	return &Compiler{
		tokens:  tokens,
		current: 0,
	}
}
