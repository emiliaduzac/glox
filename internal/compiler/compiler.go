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
	if err := c.expression(); err != nil {
		return nil, err
	}
	if err := c.cursor.Consume(token.EOF, "end of expression"); err != nil {
		return nil, err
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
	operatorType := c.cursor.GetPreviousType()
	if err := c.parsePrecedence(PREC_UNARY); err != nil {
		return err
	}

	switch operatorType {
	case token.MINUS:
		c.emit(byte(chunk.OP_NEGATE))
	case token.BANG:
		c.emit(byte(chunk.OP_NOT))
	default:
		return errors.NewLoxError(c.cursor.GetPreviousLine(), "unexpected unary operator")
	}
	return nil
}

func (c *Compiler) binary() error {
	operatorType := c.cursor.GetPreviousType()
	rule := getRule(operatorType)
	if err := c.parsePrecedence(rule.precedence + 1); err != nil {
		return err
	}

	switch operatorType {
	case token.PLUS:
		c.emit(byte(chunk.OP_ADD))
	case token.MINUS:
		c.emit(byte(chunk.OP_SUBTRACT))
	case token.STAR:
		c.emit(byte(chunk.OP_MULTIPLY))
	case token.SLASH:
		c.emit(byte(chunk.OP_DIVIDE))
	case token.PERCENT:
		c.emit(byte(chunk.OP_MODULO))
	default:
		return errors.NewLoxError(c.cursor.GetPreviousLine(), "unexpected binary operator")
	}
	return nil
}

func (c *Compiler) literal() error {
	operatorType := c.cursor.GetPreviousType()

	switch operatorType {
	case token.FALSE:
		c.emit(byte(chunk.OP_FALSE))
	case token.TRUE:
		c.emit(byte(chunk.OP_TRUE))
	case token.NIL:
		c.emit(byte(chunk.OP_NIL))
	default:
		return errors.NewLoxError(c.cursor.GetPreviousLine(), "unexpected literal")
	}
	return nil
}

// TO DO:
func (c *Compiler) logicAnd() error {
	return nil
}
func (c *Compiler) logicOr() error {
	return nil
}
func (c *Compiler) variable() error {
	return nil
}
