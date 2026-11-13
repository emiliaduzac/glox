package compiler

import (
	stderrors "errors"
	"math"
	"tp-lox/internal/chunk"
	"tp-lox/internal/errors"
	"tp-lox/internal/token"
)

type Local struct {
	Name  token.Token
	Depth int
}

type Compiler struct {
	cursor     *token.TokenCursor
	chunk      *chunk.Chunk
	locals     []Local
	scopeDepth int
}

// NewCompiler crea un nuevo compiler, a partir de una lista de tokens generados por el scanner.
func NewCompiler(tokens []token.Token) *Compiler {
	return &Compiler{
		cursor:     token.NewTokenCursor(tokens),
		chunk:      chunk.NewChunk(),
		locals:     []Local{},
		scopeDepth: 0,
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
	if err := c.declareLocalVariable(); err != nil {
		return 0, err
	}
	if c.scopeDepth > 0 {
		return 0, nil
	}
	return c.identifierConstant(c.cursor.Previous())
}

func (c *Compiler) identifierConstant(name token.Token) (byte, error) {
	return c.makeConstant(name.Lexeme)
}

func (c *Compiler) defineVariable(global byte) {
	if c.scopeDepth > 0 {
		c.markInitialized()
		return
	}
	c.emit(byte(chunk.OP_DEFINE_GLOBAL), global)
}

func (c *Compiler) declareLocalVariable() error {
	if c.scopeDepth == 0 {
		return nil
	}

	name := c.cursor.Previous()
	for i := len(c.locals) - 1; i >= 0; i-- {
		local := c.locals[i]
		if local.Depth != -1 && local.Depth < c.scopeDepth {
			break
		}
		if name.Lexeme == local.Name.Lexeme {
			return errors.NewLoxError(c.cursor.GetPreviousLine(), "already a variable with this name in this scope")
		}
	}
	return c.addLocal(name)
}

func (c *Compiler) addLocal(name token.Token) error {
	if len(c.locals) == math.MaxUint8+1 {
		return errors.NewLoxError(c.cursor.GetPreviousLine(), "too many local variables in scope")
	}
	local := Local{Name: name, Depth: -1} // no inicializada
	c.locals = append(c.locals, local)
	return nil
}

func (c *Compiler) markInitialized() {
	c.locals[len(c.locals)-1].Depth = c.scopeDepth
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
	} else if c.cursor.Match(token.IF) {
		return c.ifStatement()
	} else if c.cursor.Match(token.LEFT_BRACE) {
		c.beginScope()
		err := c.block()
		c.endScope() // siempre cerramos el scope, haya habido error o no
		return err
	}
	return c.expressionStatement()
}

func (c *Compiler) beginScope() {
	c.scopeDepth++
}

func (c *Compiler) endScope() {
	c.scopeDepth--

	for len(c.locals) > 0 && c.locals[len(c.locals)-1].Depth > c.scopeDepth {
		c.emit(byte(chunk.OP_POP))
		c.locals = c.locals[:len(c.locals)-1]
	}
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

func (c *Compiler) block() error {
	var errs []error
	for !c.cursor.Check(token.RIGHT_BRACE) && !c.cursor.Check(token.EOF) {
		if err := c.declaration(); err != nil {
			errs = append(errs, err)
		}
	}
	if err := c.cursor.Consume(token.RIGHT_BRACE, "'}' after block"); err != nil {
		errs = append(errs, err)
	}
	return stderrors.Join(errs...)
}

func (c *Compiler) ifStatement() error {
	if err := c.cursor.Consume(token.LEFT_PAREN, "'(' after if"); err != nil {
		return err
	}
	if err := c.expression(); err != nil {
		return err
	}
	if err := c.cursor.Consume(token.RIGHT_PAREN, "')' after condition"); err != nil {
		return err
	}
	thenJump := c.emitJump(chunk.OP_JUMP_IF_FALSE)
	c.emit(byte(chunk.OP_POP))
	if err := c.statement(); err != nil {
		return err
	}
	elseJump := c.emitJump(chunk.OP_JUMP)
	if err := c.patchJump(thenJump); err != nil {
		return err
	}
	c.emit(byte(chunk.OP_POP))

	if c.cursor.Match(token.ELSE) {
		if err := c.statement(); err != nil {
			return err
		}
	}
	if err := c.patchJump(elseJump); err != nil {
		return err
	}
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

func (c *Compiler) emitJump(instruction chunk.OpCode) int {
	c.emit(byte(instruction), 0xff, 0xff)
	return len(c.chunk.Bytes) - 2
}

func (c *Compiler) patchJump(offset int) error {
	jump := len(c.chunk.Bytes) - offset - 2
	if jump > math.MaxUint16 {
		return errors.NewLoxError(c.cursor.GetPreviousLine(), "too much code to jump over")
	}
	c.chunk.Bytes[offset] = byte(jump >> 8)
	c.chunk.Bytes[offset+1] = byte(jump)
	return nil
}

func (c *Compiler) makeConstant(value any) (byte, error) {
	index := c.chunk.AddConstant(value)
	if index > math.MaxUint8 {
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

	if canAssign && c.cursor.Match(token.EQUAL) {
		return errors.NewLoxError(c.cursor.GetPreviousLine(), "invalid assignment target")
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

func (c *Compiler) namedVariable(name token.Token, canAssign bool) error {
	var getOp, setOp chunk.OpCode
	var arg byte

	localIndex, err := c.resolveLocal(name)
	if err != nil {
		return err
	}
	if localIndex != -1 {
		getOp, setOp = chunk.OP_GET_LOCAL, chunk.OP_SET_LOCAL
		arg = byte(localIndex)
	} else {
		var err error
		arg, err = c.identifierConstant(name)
		if err != nil {
			return err
		}
		getOp, setOp = chunk.OP_GET_GLOBAL, chunk.OP_SET_GLOBAL
	}

	if canAssign && c.cursor.Match(token.EQUAL) {
		if err := c.expression(); err != nil {
			return err
		}
		c.emit(byte(setOp), arg)
	} else {
		c.emit(byte(getOp), arg)
	}
	return nil
}

func (c *Compiler) resolveLocal(name token.Token) (int, error) {
	for i := len(c.locals) - 1; i >= 0; i-- {
		local := c.locals[i]
		if name.Lexeme == local.Name.Lexeme {
			if local.Depth == -1 {
				return 0, errors.NewLoxError(c.cursor.GetPreviousLine(), "can't read local variable in its own initializer")
			}
			return i, nil
		}
	}
	return -1, nil
}

func (c *Compiler) variable(canAssign bool) error {
	return c.namedVariable(c.cursor.Previous(), canAssign)
}

// TO DO:
func (c *Compiler) logicAnd(_ bool) error {
	return nil
}
func (c *Compiler) logicOr(_ bool) error {
	return nil
}
