package compiler

import (
	"math"
	"tp-lox/internal/chunk"
	"tp-lox/internal/errors"
	"tp-lox/internal/token"
)

func (c *Compiler) expression() error {
	return c.parsePrecedence(PREC_ASSIGNMENT)
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

func (c *Compiler) grouping(_ bool) error {
	if err := c.expression(); err != nil {
		return err
	}
	if err := c.cursor.Consume(token.RIGHT_PAREN, "')' after expression"); err != nil {
		return err
	}
	return nil
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
	if c.cursor.Match(token.LEFT_PAREN) {
		argCount, err := c.argumentList()
		if err != nil {
			return err
		}
		c.emit(byte(chunk.OP_INVOKE), name, argCount)
		return nil
	}
	c.emit(byte(chunk.OP_GET_PROPERTY), name)
	return nil
}

func (c *Compiler) value(_ bool) error {
	value := c.cursor.Previous().Literal
	return c.emitConstant(value)
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

func (c *Compiler) variable(canAssign bool) error {
	return c.namedVariable(c.cursor.Previous(), canAssign)
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

func (c *Compiler) call(_ bool) error {
	argCount, err := c.argumentList()
	if err != nil {
		return err
	}
	c.emit(byte(chunk.OP_CALL), argCount)
	return nil
}

func (c *Compiler) this(_ bool) error {
	if c.currentClass == nil {
		c.report(errors.NewLoxError(c.cursor.GetPreviousLine(), "can't use 'this' outside of a class"))
		return nil
	}
	return c.variable(false)
}

func (c *Compiler) super(_ bool) error {
	if c.currentClass == nil {
		c.report(errors.NewLoxError(c.cursor.GetPreviousLine(), "can't use 'super' outside of a class"))
	} else if !c.currentClass.hasSuperclass {
		c.report(errors.NewLoxError(c.cursor.GetPreviousLine(), "can't use 'super' in a class with no superclass"))
	}

	if err := c.cursor.Consume(token.DOT, "'.' after 'super'"); err != nil {
		return err
	}
	if err := c.cursor.Consume(token.IDENTIFIER, "superclass method name"); err != nil {
		return err
	}
	name, err := c.identifierConstant(c.cursor.Previous())
	if err != nil {
		return err
	}

	if err := c.namedVariable(thisToken, false); err != nil {
		return err
	}
	if c.cursor.Match(token.LEFT_PAREN) {
		argCount, err := c.argumentList()
		if err != nil {
			return err
		}
		if err := c.namedVariable(superToken, false); err != nil {
			return err
		}
		c.emit(byte(chunk.OP_SUPER_INVOKE), name, argCount)
		return nil
	}
	if err := c.namedVariable(superToken, false); err != nil {
		return err
	}
	c.emit(byte(chunk.OP_GET_SUPER), name)
	return nil
}
