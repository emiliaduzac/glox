package compiler

import (
	stderrors "errors"
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
	var errs []error
	for !c.cursor.Match(token.EOF) {
		if err := c.declaration(); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return nil, stderrors.Join(errs...)
	}
	c.emitReturn()
	return c.chunk, nil
}

func (c *Compiler) expression() error {
	return c.parsePrecedence(PREC_ASSIGNMENT)
}

func (c *Compiler) declaration() error {
	var err error
	if c.cursor.Match(token.VAR) {
		err = c.varDeclaration()
	} else {
		err = c.statement()
	}
	if err != nil {
		c.synchronize()
	}
	return err
}

func (c *Compiler) varDeclaration() error {
	global, err := c.parseVariable("identifier after 'var'")
	if err != nil {
		return err
	}
	if c.cursor.Match(token.EQUAL) {
		if err := c.expression(); err != nil {
			return err
		}
	} else {
		c.emit(byte(chunk.OP_NIL))
	}
	if err := c.cursor.Consume(token.SEMICOLON, "';' after variable declaration"); err != nil {
		return err
	}
	c.defineVariable(global)
	return nil
}

func (c *Compiler) parseVariable(expected string) (byte, error) {
	if err := c.cursor.Consume(token.IDENTIFIER, expected); err != nil {
		return 0, err
	}
	return c.identifierConstant(c.cursor.Previous())
}

func (c *Compiler) identifierConstant(name token.Token) (byte, error) {
	return c.makeConstant(name.Lexeme)
}

func (c *Compiler) defineVariable(global byte) {
	c.emit(byte(chunk.OP_DEFINE_GLOBAL), global)
}

func (c *Compiler) synchronize() {
	for !c.cursor.IsAtEnd() {
		if c.cursor.GetPreviousType() == token.SEMICOLON {
			return
		}
		switch c.cursor.Current().TokenType {
		case token.CLASS, token.FUN, token.VAR, token.FOR, token.IF, token.WHILE, token.PRINT, token.RETURN:
			return
		}
		c.cursor.Advance()
	}
}

func (c *Compiler) statement() error {
	if c.cursor.Match(token.PRINT) {
		return c.printStatement()
	}
	return c.expressionStatement()
}

func (c *Compiler) printStatement() error {
	if err := c.expression(); err != nil {
		return err
	}
	if err := c.cursor.Consume(token.SEMICOLON, "';' after statement"); err != nil {
		return err
	}
	c.emit(byte(chunk.OP_PRINT))
	return nil
}

func (c *Compiler) expressionStatement() error {
	if err := c.expression(); err != nil {
		return err
	}
	if err := c.cursor.Consume(token.SEMICOLON, "';' after statement"); err != nil {
		return err
	}
	c.emit(byte(chunk.OP_POP))
	return nil
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
	index, err := c.makeConstant(value)
	if err != nil {
		return err
	}
	c.emit(byte(chunk.OP_CONSTANT), index)
	return nil
}

func (c *Compiler) makeConstant(value any) (byte, error) {
	index := c.chunk.AddConstant(value)
	if index > 255 {
		return 0, errors.NewLoxError(c.cursor.GetPreviousLine(), "too many constants in one chunk")
	}
	return byte(index), nil
}

func (c *Compiler) parsePrecedence(precedence Precedence) error {
	prefix := getRule(c.cursor.Current().TokenType).prefix
	if prefix == nil {
		err := c.cursor.ExpectedError("expression")
		c.cursor.Advance()
		return err
	}
	c.cursor.Advance()
	canAssign := precedence <= PREC_ASSIGNMENT
	if err := prefix(c, canAssign); err != nil {
		return err
	}

	for precedence <= getRule(c.cursor.Current().TokenType).precedence {
		tok := c.cursor.Advance()
		infix := getRule(tok.TokenType).infix
		if err := infix(c, canAssign); err != nil {
			return err
		}
	}

	return nil
}

func (c *Compiler) value(_ bool) error {
	value := c.cursor.Previous().Literal
	return c.emitConstant(value)
}

func (c *Compiler) grouping(_ bool) error {
	if err := c.expression(); err != nil {
		return err
	}
	if err := c.cursor.Consume(token.RIGHT_PAREN, "')' after expression"); err != nil {
		return err
	}
	return nil
}

func (c *Compiler) unary(_ bool) error {
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

func (c *Compiler) binary(_ bool) error {
	operatorType := c.cursor.GetPreviousType()
	rule := getRule(operatorType)
	if err := c.parsePrecedence(rule.precedence + 1); err != nil {
		return err
	}

	switch operatorType {
	// Aritméticos
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

	// Comparación
	case token.EQUAL_EQUAL:
		c.emit(byte(chunk.OP_EQUAL))
	case token.BANG_EQUAL:
		c.emit(byte(chunk.OP_EQUAL), byte(chunk.OP_NOT))
	case token.GREATER:
		c.emit(byte(chunk.OP_GREATER))
	case token.GREATER_EQUAL:
		c.emit(byte(chunk.OP_LESS), byte(chunk.OP_NOT))
	case token.LESS:
		c.emit(byte(chunk.OP_LESS))
	case token.LESS_EQUAL:
		c.emit(byte(chunk.OP_GREATER), byte(chunk.OP_NOT))

	default:
		return errors.NewLoxError(c.cursor.GetPreviousLine(), "unexpected binary operator")
	}
	return nil
}

func (c *Compiler) literal(_ bool) error {
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

func (c *Compiler) namedVariable(name token.Token) error {
	arg, err := c.identifierConstant(name)
	if err != nil {
		return err
	}
	if c.cursor.Match(token.EQUAL) {
		if err := c.expression(); err != nil {
			return err
		}
		c.emit(byte(chunk.OP_SET_GLOBAL), arg)
	} else {
		c.emit(byte(chunk.OP_GET_GLOBAL), arg)
	}
	return nil
}

func (c *Compiler) variable(_ bool) error {
	return c.namedVariable(c.cursor.Previous())
}

// TO DO:
func (c *Compiler) logicAnd(_ bool) error {
	return nil
}
func (c *Compiler) logicOr(_ bool) error {
	return nil
}
