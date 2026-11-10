package token_test

import (
	"testing"
	"tp-lox/internal/token"
	"tp-lox/internal/utils"
)

func TestCursorConsumesEOF(t *testing.T) {
	t.Run("Con solo EOF, consumirlo lo deja como anterior", func(t *testing.T) {
		cursor := token.NewTokenCursor([]token.Token{{TokenType: token.EOF, Line: 1}})

		utils.AssertValue(t, cursor.Match(token.EOF), true)
		utils.AssertValue(t, cursor.Previous().TokenType, token.EOF)
	})

	t.Run("Después de consumir el EOF, el anterior es el EOF y no el token previo", func(t *testing.T) {
		cursor := token.NewTokenCursor([]token.Token{
			{TokenType: token.PRINT, Line: 1},
			{TokenType: token.EOF, Line: 2},
		})

		cursor.Advance()
		cursor.Advance()

		utils.AssertValue(t, cursor.Previous().TokenType, token.EOF)
		utils.AssertValue(t, cursor.GetPreviousLine(), 2)
	})
}
