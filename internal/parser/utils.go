package parser

import (
	"fmt"
	"tp-lox/internal/errors"
	"tp-lox/internal/expression"
	"tp-lox/internal/statements"
	"tp-lox/internal/token"
)

// isAtEnd indica si el token actual es el último.
func (p *Parser) isAtEnd() bool {
	return p.tokens[p.current].TokenType == token.EOF
}

// isInTokenTypes verifica si el token actual es de alguno de los tipos de token pasados como argumento
func (p *Parser) isInTokenTypes(types ...token.TokenType) bool {
	for _, t := range types {
		if p.isTokenType(t) {
			p.advance()
			return true
		}
	}
	return false
}

// isTokenType verifica si el token actual es del tipo de token pasado como argumento
func (p *Parser) isTokenType(t token.TokenType) bool {
	return p.getCurrent().TokenType == t
}

func (p *Parser) getCurrent() token.Token {
	return p.tokens[p.current]
}

func (p *Parser) getPrevious() token.Token {
	return p.tokens[p.current-1]
}

func (p *Parser) getLine() int {
	return p.getCurrent().Line
}

func (p *Parser) getLexeme() string {
	return p.getCurrent().Lexeme
}

func (p *Parser) expectedError(expected string) error {
	got := fmt.Sprintf("`%s`", p.getLexeme())
	if p.isTokenType(token.EOF) {
		got = "EOF"
	}
	return errors.NewLoxError(p.getLine(), fmt.Sprintf("expected %s, got %s instead", expected, got))
}

// advance avanza al siguiente token y devuelve el token consumido
func (p *Parser) advance() token.Token {
	if !p.isAtEnd() {
		p.current++
	}
	return p.getPrevious()
}

func (p *Parser) parseForInitializer() (statements.Stmt, error) {
	var initializer statements.Stmt
	var err error
	if p.isInTokenTypes(token.VAR) {
		initializer, err = p.varDeclaration()
	} else if !p.isInTokenTypes(token.SEMICOLON) {
		initializer, err = p.expressionStatement()
	}
	return initializer, err
}

func (p *Parser) parseForCondition() (expression.Expr, error) {
	var condition expression.Expr
	var err error
	if !p.isTokenType(token.SEMICOLON) {
		condition, err = p.expression()
		if err != nil {
			return nil, err
		}
	}
	if !p.isInTokenTypes(token.SEMICOLON) {
		return nil, p.expectedError("';' after for condition")
	}
	return condition, nil
}

func (p *Parser) parseForIncrement() (expression.Expr, error) {
	var increment expression.Expr
	var err error
	if !p.isTokenType(token.RIGHT_PAREN) {
		increment, err = p.expression()
		if err != nil {
			return nil, err
		}
	}
	if !p.isInTokenTypes(token.RIGHT_PAREN) {
		return nil, p.expectedError("')' after for clauses")
	}
	return increment, nil
}
