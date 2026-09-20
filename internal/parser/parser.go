package parser

import (
	"fmt"
	"tp-lox/internal/errors"
	"tp-lox/internal/expression"
	"tp-lox/internal/token"
)

type Parser struct {
	tokens  []token.Token
	current int
}

// NewParser crea un nuevo parser, a partir de una lista de tokens generados por el scanner.
func NewParser(tokens []token.Token) *Parser {
	return &Parser{
		tokens:  tokens,
		current: 0,
	}
}

func (p *Parser) Parse() (expression.Expr, error) {
	expression, err := p.expression()
	if err != nil {
		return nil, err
	}
	return expression, nil
}

// expression es el punto de entrada para parsear una expresión. Comienza el proceso
// de llamadas recursivas para parsear la expresión completa, siguiendo la jerarquía de
// precedencia de operadores.
func (p *Parser) expression() (expression.Expr, error) {
	return p.equality()
}

func (p *Parser) equality() (expression.Expr, error) {
	var expr, err = p.comparison()
	if err != nil {
		return nil, err
	}

	for p.isInTokenTypes(token.BANG_EQUAL, token.EQUAL_EQUAL) {
		operator := p.getPrevious()
		right, err := p.comparison()
		if err != nil {
			return nil, err
		}
		expr = expression.NewBinary(expr, operator, right)
	}

	return expr, nil
}

func (p *Parser) comparison() (expression.Expr, error) {
	var expr, err = p.term()
	if err != nil {
		return nil, err
	}

	for p.isInTokenTypes(token.GREATER, token.GREATER_EQUAL, token.LESS, token.LESS_EQUAL) {
		operator := p.getPrevious()
		right, err := p.term()
		if err != nil {
			return nil, err
		}
		expr = expression.NewBinary(expr, operator, right)
	}

	return expr, nil
}

func (p *Parser) term() (expression.Expr, error) {
	var expr, err = p.factor()
	if err != nil {
		return nil, err
	}

	for p.isInTokenTypes(token.MINUS, token.PLUS) {
		operator := p.getPrevious()
		right, err := p.factor()
		if err != nil {
			return nil, err
		}
		expr = expression.NewBinary(expr, operator, right)
	}

	return expr, nil
}

func (p *Parser) factor() (expression.Expr, error) {
	var expr, err = p.unary()
	if err != nil {
		return nil, err
	}

	for p.isInTokenTypes(token.SLASH, token.STAR) {
		operator := p.getPrevious()
		right, err := p.unary()
		if err != nil {
			return nil, err
		}
		expr = expression.NewBinary(expr, operator, right)
	}

	return expr, nil
}

func (p *Parser) unary() (expression.Expr, error) {
	if p.isInTokenTypes(token.BANG, token.MINUS) {
		operator := p.getPrevious()
		right, err := p.unary()
		if err != nil {
			return nil, err
		}
		return expression.NewUnary(operator, right), nil
	}

	return p.primary()
}

func (p *Parser) primary() (expression.Expr, error) {
	if p.isInTokenTypes(token.FALSE) {
		return expression.NewLiteral(false), nil
	}
	if p.isInTokenTypes(token.TRUE) {
		return expression.NewLiteral(true), nil
	}
	if p.isInTokenTypes(token.NIL) {
		return expression.NewLiteral(nil), nil
	}
	if p.isInTokenTypes(token.NUMBER, token.STRING) {
		return expression.NewLiteral(p.getPrevious().Literal), nil
	}
	if p.isInTokenTypes(token.LEFT_PAREN) {
		expr, err := p.expression()
		if err != nil {
			return nil, err
		}
		if !p.isInTokenTypes(token.RIGHT_PAREN) {
			return nil, errors.NewLoxError(p.getCurrent().Line, "expected ')' after expression")
		}
		return expression.NewGrouping(expr), nil
	}

	return nil, errors.NewLoxError(p.getCurrent().Line, fmt.Sprintf("expected expression, got `%s`", p.getCurrent().Lexeme))
}

// isAtEnd indica si el token actual es el
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

// advance avanza al siguiente token y devuelve el token consumido
func (p *Parser) advance() token.Token {
	if !p.isAtEnd() {
		p.current++
	}
	return p.getPrevious()
}
