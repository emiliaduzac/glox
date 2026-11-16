package compiler

import (
	"math"
	"tp-lox/internal/chunk"
	"tp-lox/internal/errors"
	"tp-lox/internal/token"
)

func (c *Compiler) currentChunk() *chunk.Chunk {
	return &c.function.Chunk
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

func (c *Compiler) emit(bytes ...byte) {
	for _, b := range bytes {
		c.currentChunk().Write(b, c.cursor.GetPreviousLine())
	}
}

func (c *Compiler) emitReturn() {
	if c.funcType == TYPE_INITIALIZER {
		c.emit(byte(chunk.OP_GET_LOCAL), 0)
	} else {
		c.emit(byte(chunk.OP_NIL))
	}
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

func (c *Compiler) report(err error) {
	if c.panicMode {
		return
	}
	c.panicMode = true
	c.errs = append(c.errs, err)
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
