package parser

import (
	"fmt"
	"tp-lox/internal/errors"
	"tp-lox/internal/expression"
	"tp-lox/internal/statements"
	"tp-lox/internal/token"
)

func (p *Parser) expectedError(expected string) error {
	got := fmt.Sprintf("`%s`", p.cursor.GetLexeme())
	if p.cursor.Check(token.EOF) {
		got = "EOF"
	}
	return errors.NewLoxError(p.cursor.GetLine(), fmt.Sprintf("expected %s, got %s instead", expected, got))
}

func (p *Parser) parseForInitializer() (statements.Stmt, error) {
	var initializer statements.Stmt
	var err error
	if p.cursor.Match(token.VAR) {
		initializer, err = p.varDeclaration()
	} else if !p.cursor.Match(token.SEMICOLON) {
		initializer, err = p.expressionStatement()
	}
	return initializer, err
}

func (p *Parser) parseForCondition() (expression.Expr, error) {
	var condition expression.Expr
	var err error
	if !p.cursor.Check(token.SEMICOLON) {
		condition, err = p.expression()
		if err != nil {
			return nil, err
		}
	}
	if !p.cursor.Match(token.SEMICOLON) {
		return nil, p.expectedError("';' after for condition")
	}
	return condition, nil
}

func (p *Parser) parseForIncrement() (expression.Expr, error) {
	var increment expression.Expr
	var err error
	if !p.cursor.Check(token.RIGHT_PAREN) {
		increment, err = p.expression()
		if err != nil {
			return nil, err
		}
	}
	if !p.cursor.Match(token.RIGHT_PAREN) {
		return nil, p.expectedError("')' after for clauses")
	}
	return increment, nil
}
