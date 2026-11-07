package compiler

import (
	"tp-lox/internal/chunk"
	"tp-lox/internal/errors"
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
	c.emit(byte(chunk.OP_RETURN))
}

func (c *Compiler) emitConstant(value any) error {
	index := c.chunk.AddConstant(value)
	if index > 255 {
		return errors.NewLoxError(c.cursor.GetPreviousLine(), "too many constants in one chunk")
	}
	c.emit(byte(chunk.OP_CONSTANT), byte(index))
	return nil
}

func (c *Compiler) parsePrecedence(precedence Precedence) error {
	prefix := getRule(c.cursor.Current().TokenType).prefix
	if prefix == nil {
		return c.cursor.ExpectedError("expression")
	}
	c.cursor.Advance()
	if err := prefix(c); err != nil {
		return err
	}

	for precedence <= getRule(c.cursor.Current().TokenType).precedence {
		tok := c.cursor.Advance()
		infix := getRule(tok.TokenType).infix
		if err := infix(c); err != nil {
			return err
		}
	}

	return nil
}

func (c *Compiler) expression() error {
	return c.parsePrecedence(PREC_ASSIGNMENT)
}

func (c *Compiler) value() error {
	value := c.cursor.Previous().Literal
	return c.emitConstant(value)
}

func (c *Compiler) grouping() error {
	if err := c.expression(); err != nil {
		return err
	}
	if err := c.cursor.Consume(token.RIGHT_PAREN, "')' after expression"); err != nil {
		return err
	}
	return nil
}

func (c *Compiler) unary() error {
	return nil
}
func (c *Compiler) binary() error {
	return nil
}
func (c *Compiler) literal() error {
	return nil
}
func (c *Compiler) logicAnd() error {
	return nil
}
func (c *Compiler) logicOr() error {
	return nil
}
func (c *Compiler) variable() error {
	return nil
}
