package compiler

import (
	stderrors "errors"
	"math"
	"tp-lox/internal/chunk"
	"tp-lox/internal/errors"
	"tp-lox/internal/token"
)

type Local struct {
	Name       token.Token
	Depth      int
	IsCaptured bool
}

type FunctionType int

const (
	TYPE_FUNCTION FunctionType = iota
	TYPE_SCRIPT
	TYPE_METHOD
)

type parserState struct {
	cursor    *token.TokenCursor
	errs      []error
	panicMode bool
}

type Loop struct {
	loopStart  int
	scopeDepth int
	breakJumps []int
}

type Upvalue struct {
	index   byte
	isLocal bool
}

type Compiler struct {
	*parserState
	enclosing   *Compiler
	function    *chunk.Function
	funcType    FunctionType
	locals      []Local
	scopeDepth  int
	identifiers map[string]byte
	loops       []Loop
	upvalues    []Upvalue
}

// NewCompiler crea un nuevo compiler, a partir de una lista de tokens generados por el scanner.
func NewCompiler(tokens []token.Token) *Compiler {
	state := &parserState{cursor: token.NewTokenCursor(tokens)}
	return newFunctionCompiler(state, nil, TYPE_SCRIPT)
}

func newFunctionCompiler(state *parserState, enclosing *Compiler, funcType FunctionType) *Compiler {
	function := chunk.NewFunction()
	if funcType != TYPE_SCRIPT {
		function.Name = state.cursor.GetPreviousLexeme()
	}
	slotZero := Local{Depth: 0}
	if funcType != TYPE_FUNCTION {
		slotZero.Name = token.Token{TokenType: token.THIS, Lexeme: "this"}
	}
	return &Compiler{
		parserState: state,
		enclosing:   enclosing,
		function:    function,
		funcType:    funcType,
		locals:      []Local{slotZero},
		identifiers: map[string]byte{},
	}
}

func (c *Compiler) currentChunk() *chunk.Chunk {
	return &c.function.Chunk
}

func (c *Compiler) Compile() (*chunk.Function, error) {
	for !c.cursor.Match(token.EOF) {
		c.declaration()
	}
	function := c.endCompiler()
	if len(c.errs) > 0 {
		return nil, stderrors.Join(c.errs...)
	}
	return function, nil
}

func (c *Compiler) endCompiler() *chunk.Function {
	c.emitReturn()
	return c.function
}

func (c *Compiler) expression() error {
	return c.parsePrecedence(PREC_ASSIGNMENT)
}

func (c *Compiler) declaration() {
	var err error
	if c.cursor.Match(token.CLASS) {
		err = c.classDeclaration()
	} else if c.cursor.Match(token.FUN) {
		err = c.funDeclaration()
	} else if c.cursor.Match(token.VAR) {
		err = c.varDeclaration()
	} else {
		err = c.statement()
	}
	if err != nil {
		c.report(err)
	}
	if c.panicMode {
		c.synchronize()
	}
}

func (c *Compiler) report(err error) {
	if c.panicMode {
		return
	}
	c.panicMode = true
	c.errs = append(c.errs, err)
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
	c.declareLocalVariable()
	if c.scopeDepth > 0 {
		return 0, nil
	}
	return c.identifierConstant(c.cursor.Previous())
}

func (c *Compiler) identifierConstant(name token.Token) (byte, error) {
	if index, ok := c.identifiers[name.Lexeme]; ok {
		return index, nil
	}
	index, err := c.makeConstant(name.Lexeme)
	if err != nil {
		return 0, err
	}
	c.identifiers[name.Lexeme] = index
	return index, nil
}

func (c *Compiler) defineVariable(global byte) {
	if c.scopeDepth > 0 {
		c.markInitialized()
		return
	}
	c.emit(byte(chunk.OP_DEFINE_GLOBAL), global)
}

func (c *Compiler) declareLocalVariable() {
	if c.scopeDepth == 0 {
		return
	}

	name := c.cursor.Previous()
	for i := len(c.locals) - 1; i >= 0; i-- {
		local := c.locals[i]
		if local.Depth != -1 && local.Depth < c.scopeDepth {
			break
		}
		if name.Lexeme == local.Name.Lexeme {
			c.report(errors.NewLoxError(c.cursor.GetPreviousLine(), "already a variable with this name in this scope"))
		}
	}
	c.addLocal(name)
}

func (c *Compiler) addLocal(name token.Token) {
	if len(c.locals) == math.MaxUint8+1 {
		c.report(errors.NewLoxError(c.cursor.GetPreviousLine(), "too many local variables in scope"))
		return
	}
	local := Local{Name: name, Depth: -1, IsCaptured: false} // no inicializada
	c.locals = append(c.locals, local)
}

func (c *Compiler) markInitialized() {
	if c.scopeDepth == 0 {
		return
	}
	c.locals[len(c.locals)-1].Depth = c.scopeDepth
}

func (c *Compiler) synchronize() {
	c.panicMode = false
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
	} else if c.cursor.Match(token.WHILE) {
		return c.whileStatement()
	} else if c.cursor.Match(token.FOR) {
		return c.forStatement()
	} else if c.cursor.Match(token.RETURN) {
		return c.returnStatement()
	} else if c.cursor.Match(token.LEFT_BRACE) {
		c.beginScope()
		defer c.endScope()
		return c.block()
	} else if c.cursor.Match(token.CONTINUE) {
		return c.continueStatement()
	} else if c.cursor.Match(token.BREAK) {
		return c.breakStatement()
	}
	return c.expressionStatement()
}

func (c *Compiler) beginScope() {
	c.scopeDepth++
}

func (c *Compiler) endScope() {
	c.scopeDepth--

	for len(c.locals) > 0 && c.locals[len(c.locals)-1].Depth > c.scopeDepth {
		c.emitPopLocal(c.locals[len(c.locals)-1])
		c.locals = c.locals[:len(c.locals)-1]
	}
}

func (c *Compiler) beginLoop(loopStart int) {
	c.loops = append(c.loops, Loop{loopStart: loopStart, scopeDepth: c.scopeDepth})
}

func (c *Compiler) endLoop() {
	c.loops = c.loops[:len(c.loops)-1]
}

func (c *Compiler) getCurrentLoop() *Loop {
	return &c.loops[len(c.loops)-1]
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

func (c *Compiler) returnStatement() error {
	if c.funcType == TYPE_SCRIPT {
		c.report(errors.NewLoxError(c.cursor.GetPreviousLine(), "can't return from top-level code"))
	}
	if c.cursor.Match(token.SEMICOLON) {
		c.emitReturn()
		return nil
	}
	if err := c.expression(); err != nil {
		return err
	}
	if err := c.cursor.Consume(token.SEMICOLON, "';' after return value"); err != nil {
		return err
	}
	c.emit(byte(chunk.OP_RETURN))
	return nil
}

func (c *Compiler) classDeclaration() error {
	if err := c.cursor.Consume(token.IDENTIFIER, "class name"); err != nil {
		return err
	}
	className := c.cursor.Previous()
	nameConstant, err := c.identifierConstant(className)
	if err != nil {
		return err
	}
	c.declareLocalVariable()

	c.emit(byte(chunk.OP_CLASS), nameConstant)
	c.defineVariable(nameConstant)

	if err := c.namedVariable(className, false); err != nil {
		return err
	}
	if err := c.cursor.Consume(token.LEFT_BRACE, "'{' before class body"); err != nil {
		return err
	}
	for !c.cursor.Check(token.RIGHT_BRACE) && !c.cursor.Check(token.EOF) {
		if err := c.method(); err != nil {
			return err
		}
	}
	if err := c.cursor.Consume(token.RIGHT_BRACE, "'}' after class body"); err != nil {
		return err
	}
	c.emit(byte(chunk.OP_POP))
	return nil
}

func (c *Compiler) funDeclaration() error {
	global, err := c.parseVariable("function name")
	if err != nil {
		return err
	}
	c.markInitialized()
	if err := c.compileFunction(TYPE_FUNCTION); err != nil {
		return err
	}
	c.defineVariable(global)
	return nil
}

func (c *Compiler) compileFunction(funcType FunctionType) error {
	fc := newFunctionCompiler(c.parserState, c, funcType)
	fc.beginScope()

	if err := c.cursor.Consume(token.LEFT_PAREN, "'(' after function name"); err != nil {
		return err
	}
	if !c.cursor.Check(token.RIGHT_PAREN) {
		for {
			fc.function.Arity++
			if fc.function.Arity > math.MaxUint8 {
				c.report(errors.NewLoxError(c.cursor.GetLine(), "cannot have more than 255 parameters"))
			}
			param, err := fc.parseVariable("parameter name")
			if err != nil {
				return err
			}
			fc.defineVariable(param)
			if !c.cursor.Match(token.COMMA) {
				break
			}
		}
	}
	if err := c.cursor.Consume(token.RIGHT_PAREN, "')' after parameters"); err != nil {
		return err
	}
	if err := c.cursor.Consume(token.LEFT_BRACE, "'{' before function body"); err != nil {
		return err
	}
	if err := fc.block(); err != nil {
		return err
	}

	index, err := c.makeConstant(fc.endCompiler())
	if err != nil {
		return err
	}
	c.emit(byte(chunk.OP_CLOSURE), index)

	for _, upvalue := range fc.upvalues {
		isLocal := byte(0)
		if upvalue.isLocal {
			isLocal = 1
		}
		c.emit(isLocal, upvalue.index)
	}
	return nil
}

func (c *Compiler) method() error {
	if err := c.cursor.Consume(token.IDENTIFIER, "method name"); err != nil {
		return err
	}
	constant, err := c.identifierConstant(c.cursor.Previous())
	if err != nil {
		return err
	}
	if err := c.compileFunction(TYPE_METHOD); err != nil {
		return err
	}
	c.emit(byte(chunk.OP_METHOD), constant)
	return nil
}

func (c *Compiler) block() error {
	for !c.cursor.Check(token.RIGHT_BRACE) && !c.cursor.Check(token.EOF) {
		c.declaration()
	}
	return c.cursor.Consume(token.RIGHT_BRACE, "'}' after block")
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

func (c *Compiler) whileStatement() error {
	loopStart := len(c.currentChunk().Bytes)
	c.beginLoop(loopStart)
	defer c.endLoop()

	if err := c.cursor.Consume(token.LEFT_PAREN, "'(' after 'while'"); err != nil {
		return err
	}
	if err := c.expression(); err != nil {
		return err
	}
	if err := c.cursor.Consume(token.RIGHT_PAREN, "')' after condition"); err != nil {
		return err
	}

	exitJump := c.emitJump(chunk.OP_JUMP_IF_FALSE)
	c.emit(byte(chunk.OP_POP))
	if err := c.statement(); err != nil {
		return err
	}
	if err := c.emitLoop(loopStart); err != nil {
		return err
	}
	if err := c.patchJump(exitJump); err != nil {
		return err
	}
	c.emit(byte(chunk.OP_POP))

	return c.patchBreakJumps()
}

func (c *Compiler) forStatement() error {
	c.beginScope()
	defer c.endScope()

	if err := c.cursor.Consume(token.LEFT_PAREN, "'(' after 'for'"); err != nil {
		return err
	}
	if err := c.forInitializer(); err != nil {
		return err
	}

	loopStart := len(c.currentChunk().Bytes)
	exitJump, err := c.forCondition()
	if err != nil {
		return err
	}
	loopStart, err = c.forIncrement(loopStart)
	if err != nil {
		return err
	}
	c.beginLoop(loopStart)
	defer c.endLoop()

	if err := c.statement(); err != nil {
		return err
	}
	if err := c.emitLoop(loopStart); err != nil {
		return err
	}

	if exitJump != -1 {
		if err := c.patchJump(exitJump); err != nil {
			return err
		}
		c.emit(byte(chunk.OP_POP))
	}

	return c.patchBreakJumps()
}

func (c *Compiler) forInitializer() error {
	if c.cursor.Match(token.SEMICOLON) {
		return nil // sin inicializador
	}
	if c.cursor.Match(token.VAR) {
		return c.varDeclaration()
	}
	return c.expressionStatement()
}

func (c *Compiler) forCondition() (int, error) {
	if c.cursor.Match(token.SEMICOLON) {
		return -1, nil // sin condición
	}
	if err := c.expression(); err != nil {
		return -1, err
	}
	if err := c.cursor.Consume(token.SEMICOLON, "';' after loop condition"); err != nil {
		return -1, err
	}
	exitJump := c.emitJump(chunk.OP_JUMP_IF_FALSE)
	c.emit(byte(chunk.OP_POP))
	return exitJump, nil
}

func (c *Compiler) forIncrement(loopStart int) (int, error) {
	if c.cursor.Match(token.RIGHT_PAREN) {
		return loopStart, nil // sin incremento
	}

	bodyJump := c.emitJump(chunk.OP_JUMP)
	incrementStart := len(c.currentChunk().Bytes)
	if err := c.expression(); err != nil {
		return loopStart, err
	}
	c.emit(byte(chunk.OP_POP))
	if err := c.cursor.Consume(token.RIGHT_PAREN, "')' after for clauses"); err != nil {
		return loopStart, err
	}
	if err := c.emitLoop(loopStart); err != nil {
		return loopStart, err
	}
	if err := c.patchJump(bodyJump); err != nil {
		return loopStart, err
	}
	return incrementStart, nil
}

func (c *Compiler) continueStatement() error {
	if err := c.cursor.Consume(token.SEMICOLON, "';' after continue"); err != nil {
		return err
	}
	if len(c.loops) == 0 {
		c.report(errors.NewLoxError(c.cursor.GetPreviousLine(), "found 'continue' outside a loop"))
		return nil
	}

	currentLoop := c.getCurrentLoop()
	c.popLoopLocals(currentLoop)
	return c.emitLoop(currentLoop.loopStart)
}

func (c *Compiler) breakStatement() error {
	if err := c.cursor.Consume(token.SEMICOLON, "';' after break"); err != nil {
		return err
	}
	if len(c.loops) == 0 {
		c.report(errors.NewLoxError(c.cursor.GetPreviousLine(), "found 'break' outside a loop"))
		return nil
	}

	currentLoop := c.getCurrentLoop()
	c.popLoopLocals(currentLoop)
	breakJump := c.emitJump(chunk.OP_JUMP)
	currentLoop.breakJumps = append(currentLoop.breakJumps, breakJump)
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
		c.currentChunk().Write(b, c.cursor.GetPreviousLine())
	}
}

func (c *Compiler) emitReturn() {
	c.emit(byte(chunk.OP_NIL))
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
	return len(c.currentChunk().Bytes) - 2
}

func (c *Compiler) emitLoop(loopStart int) error {
	c.emit(byte(chunk.OP_LOOP))

	offset := len(c.currentChunk().Bytes) - loopStart + 2
	if offset > math.MaxUint16 {
		return errors.NewLoxError(c.cursor.GetPreviousLine(), "loop body too large")
	}

	c.emit(byte(offset>>8), byte(offset))
	return nil
}

func (c *Compiler) patchJump(offset int) error {
	jump := len(c.currentChunk().Bytes) - offset - 2
	if jump > math.MaxUint16 {
		return errors.NewLoxError(c.cursor.GetPreviousLine(), "too much code to jump over")
	}
	c.currentChunk().Bytes[offset] = byte(jump >> 8)
	c.currentChunk().Bytes[offset+1] = byte(jump)
	return nil
}

func (c *Compiler) patchBreakJumps() error {
	for _, breakJump := range c.getCurrentLoop().breakJumps {
		if err := c.patchJump(breakJump); err != nil {
			return err
		}
	}
	return nil
}

func (c *Compiler) popLoopLocals(currentLoop *Loop) {
	for i := len(c.locals) - 1; i >= 0 && c.locals[i].Depth > currentLoop.scopeDepth; i-- {
		c.emitPopLocal(c.locals[i])
	}
}

func (c *Compiler) emitPopLocal(local Local) {
	if local.IsCaptured {
		c.emit(byte(chunk.OP_CLOSE_UPVALUE))
	} else {
		c.emit(byte(chunk.OP_POP))
	}
}

func (c *Compiler) makeConstant(value any) (byte, error) {
	index := c.currentChunk().AddConstant(value)
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

func (c *Compiler) call(_ bool) error {
	argCount, err := c.argumentList()
	if err != nil {
		return err
	}
	c.emit(byte(chunk.OP_CALL), argCount)
	return nil
}

func (c *Compiler) argumentList() (byte, error) {
	argCount := 0
	if !c.cursor.Check(token.RIGHT_PAREN) {
		for {
			if err := c.expression(); err != nil {
				return 0, err
			}
			if argCount == math.MaxUint8 {
				c.report(errors.NewLoxError(c.cursor.GetPreviousLine(), "cannot have more than 255 arguments"))
			}
			argCount++
			if !c.cursor.Match(token.COMMA) {
				break
			}
		}
	}
	if err := c.cursor.Consume(token.RIGHT_PAREN, "')' after arguments"); err != nil {
		return 0, err
	}
	return byte(argCount), nil
}

func (c *Compiler) dot(canAssign bool) error {
	if err := c.cursor.Consume(token.IDENTIFIER, "property name after '.'"); err != nil {
		return err
	}
	name, err := c.identifierConstant(c.cursor.Previous())
	if err != nil {
		return err
	}

	if canAssign && c.cursor.Match(token.EQUAL) {
		if err := c.expression(); err != nil {
			return err
		}
		c.emit(byte(chunk.OP_SET_PROPERTY), name)
		return nil
	}
	c.emit(byte(chunk.OP_GET_PROPERTY), name)
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

	if localIndex := c.resolveLocal(name); localIndex != -1 {
		getOp, setOp = chunk.OP_GET_LOCAL, chunk.OP_SET_LOCAL
		arg = byte(localIndex)
	} else if upIndex := c.resolveUpvalue(name); upIndex != -1 {
		getOp, setOp = chunk.OP_GET_UPVALUE, chunk.OP_SET_UPVALUE
		arg = byte(upIndex)
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

func (c *Compiler) resolveLocal(name token.Token) int {
	for i := len(c.locals) - 1; i >= 0; i-- {
		local := c.locals[i]
		if name.Lexeme == local.Name.Lexeme {
			if local.Depth == -1 {
				c.report(errors.NewLoxError(c.cursor.GetPreviousLine(), "can't read local variable in its own initializer"))
			}
			return i
		}
	}
	return -1
}

func (c *Compiler) resolveUpvalue(name token.Token) int {
	if c.enclosing == nil {
		return -1
	}

	if local := c.enclosing.resolveLocal(name); local != -1 {
		c.enclosing.locals[local].IsCaptured = true
		return c.addUpvalue(byte(local), true)
	}
	if upvalue := c.enclosing.resolveUpvalue(name); upvalue != -1 {
		return c.addUpvalue(byte(upvalue), false)
	}
	return -1
}

func (c *Compiler) addUpvalue(index byte, isLocal bool) int {
	for i, upvalue := range c.upvalues {
		if upvalue.index == index && upvalue.isLocal == isLocal {
			return i
		}
	}

	if len(c.upvalues) == math.MaxUint8+1 {
		c.report(errors.NewLoxError(c.cursor.GetPreviousLine(), "too many closure variables in function"))
		return 0
	}

	c.upvalues = append(c.upvalues, Upvalue{index: index, isLocal: isLocal})
	c.function.UpvalueCount++
	return len(c.upvalues) - 1
}

func (c *Compiler) variable(canAssign bool) error {
	return c.namedVariable(c.cursor.Previous(), canAssign)
}

func (c *Compiler) this(_ bool) error {
	return c.variable(false)
}

func (c *Compiler) logicAnd(_ bool) error {
	endJump := c.emitJump(chunk.OP_JUMP_IF_FALSE)
	c.emit(byte(chunk.OP_POP))
	if err := c.parsePrecedence(PREC_AND); err != nil {
		return err
	}
	if err := c.patchJump(endJump); err != nil {
		return err
	}
	return nil
}

func (c *Compiler) logicOr(_ bool) error {
	elseJump := c.emitJump(chunk.OP_JUMP_IF_FALSE)
	endJump := c.emitJump(chunk.OP_JUMP)

	if err := c.patchJump(elseJump); err != nil {
		return err
	}
	c.emit(byte(chunk.OP_POP))

	if err := c.parsePrecedence(PREC_OR); err != nil {
		return err
	}
	if err := c.patchJump(endJump); err != nil {
		return err
	}
	return nil
}
