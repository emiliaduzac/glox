package compiler

import (
	stderrors "errors"
	"math"
	"tp-lox/internal/chunk"
	"tp-lox/internal/errors"
	"tp-lox/internal/token"
)

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
	if c.funcType == TYPE_INITIALIZER {
		c.report(errors.NewLoxError(c.cursor.GetPreviousLine(), "can't return a value from an initializer"))
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

	c.currentClass = &classCompiler{enclosing: c.currentClass}
	defer func() { c.currentClass = c.currentClass.enclosing }()

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
	funcType := TYPE_METHOD
	if c.cursor.GetPreviousLexeme() == "init" {
		funcType = TYPE_INITIALIZER
	}
	if err := c.compileFunction(funcType); err != nil {
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
