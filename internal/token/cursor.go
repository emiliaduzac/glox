package token

import (
	"fmt"
	"tp-lox/internal/errors"
)

// TokenCursor recorre una lista de tokens, manteniendo la posición actual.
// Lo comparten el parser y el compiler para no repetir la lógica de avanzar y consultar tokens.
type TokenCursor struct {
	tokens   []Token
	current  int
	previous Token
}

// NewTokenCursor crea un cursor posicionado en el primer token de la lista.
func NewTokenCursor(tokens []Token) *TokenCursor {
	return &TokenCursor{tokens: tokens, current: 0}
}

// IsAtEnd indica si el token actual es el EOF.
func (c *TokenCursor) IsAtEnd() bool {
	return c.Current().TokenType == EOF
}

// Current devuelve el token actual, sin consumirlo.
func (c *TokenCursor) Current() Token {
	return c.tokens[c.current]
}

// Previous devuelve el último token consumido.
func (c *TokenCursor) Previous() Token {
	return c.previous
}

// GetLine devuelve la línea del token actual.
func (c *TokenCursor) GetLine() int {
	return c.Current().Line
}

// GetLexeme devuelve el lexema del token actual.
func (c *TokenCursor) GetLexeme() string {
	return c.Current().Lexeme
}

// GetPreviousLexeme devuelve el lexema del último token consumido.
func (c *TokenCursor) GetPreviousLexeme() string {
	return c.Previous().Lexeme
}

// GetPreviousType devuelve el TokenType del último token consumido.
func (c *TokenCursor) GetPreviousType() TokenType {
	return c.Previous().TokenType
}

// GetPreviousLine devuelve la línea del último token consumido.
func (c *TokenCursor) GetPreviousLine() int {
	return c.Previous().Line
}

// Advance consume el token actual y devuelve el token consumido. No avanza más allá del EOF.
func (c *TokenCursor) Advance() Token {
	c.previous = c.Current()
	if !c.IsAtEnd() {
		c.current++
	}
	return c.previous
}

// Check indica si el token actual es del tipo pasado, sin consumirlo.
func (c *TokenCursor) Check(tokenType TokenType) bool {
	return c.Current().TokenType == tokenType
}

// Match consume el token actual si es de alguno de los tipos pasados, y devuelve si hubo coincidencia.
func (c *TokenCursor) Match(types ...TokenType) bool {
	for _, tokenType := range types {
		if c.Check(tokenType) {
			c.Advance()
			return true
		}
	}
	return false
}

// ExpectedError devuelve un error en la línea del token actual, indicando qué se esperaba y qué se encontró.
func (c *TokenCursor) ExpectedError(expected string) error {
	got := fmt.Sprintf("`%s`", c.GetLexeme())
	if c.Check(EOF) {
		got = "EOF"
	}
	return errors.NewLoxError(c.GetLine(), fmt.Sprintf("expected %s, got %s instead", expected, got))
}

// Consume avanza si el token actual es del tipo esperado. Si no, devuelve un error.
func (c *TokenCursor) Consume(tokenType TokenType, expected string) error {
	if c.Match(tokenType) {
		return nil
	}
	return c.ExpectedError(expected)
}
