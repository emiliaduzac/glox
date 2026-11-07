package parser

import (
	"tp-lox/internal/expression"
	"tp-lox/internal/statements"
	"tp-lox/internal/token"
)

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
	if err := p.cursor.Consume(token.SEMICOLON, "';' after for condition"); err != nil {
		return nil, err
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
	if err := p.cursor.Consume(token.RIGHT_PAREN, "')' after for clauses"); err != nil {
		return nil, err
	}
	return increment, nil
}
