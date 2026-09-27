package expression_test

import (
	"testing"
	"tp-lox/internal/expression"
	"tp-lox/internal/token"
)

func assertString(t *testing.T, expr expression.Expr, expected string) {
	t.Helper()
	if got := expression.String(expr); got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func assertValue(t *testing.T, got any, expected any) {
	t.Helper()
	if got != expected {
		t.Errorf("got %v, expected %v", got, expected)
	}
}

func TestCallExpression(t *testing.T) {
	paren := *token.NewToken(token.RIGHT_PAREN, ")", nil, 1)
	callee := expression.NewVariable(*token.NewToken(token.IDENTIFIER, "foo", nil, 1))

	t.Run("Instantiate call expression and verify fields", func(t *testing.T) {
		arg := expression.NewLiteral(19.0)
		call := expression.NewCall(callee, paren, []expression.Expr{arg})

		if call.Callee != callee {
			t.Errorf("expected callee %v, got %v", callee, call.Callee)
		}
		assertValue(t, call.Paren.TokenType, token.RIGHT_PAREN)
		assertValue(t, len(call.Arguments), 1)

		var expr expression.Expr = call
		expr.Expression()
	})

	t.Run("Format call without arguments", func(t *testing.T) {
		call := expression.NewCall(callee, paren, nil)
		assertString(t, call, "(call <foo>)")
	})

	t.Run("Format call with single argument", func(t *testing.T) {
		arg := expression.NewLiteral(10.0)
		call := expression.NewCall(callee, paren, []expression.Expr{arg})
		assertString(t, call, "(call <foo> <10>)")
	})

	t.Run("Format call with multiple arguments", func(t *testing.T) {
		arg1 := expression.NewLiteral(1.0)
		arg2 := expression.NewLiteral("bar")
		call := expression.NewCall(callee, paren, []expression.Expr{arg1, arg2})
		assertString(t, call, "(call <foo> <1> <\"bar\">)")
	})

	t.Run("Format chained nested calls", func(t *testing.T) {
		innerCall := expression.NewCall(callee, paren, nil)
		outerCall := expression.NewCall(innerCall, paren, nil)
		assertString(t, outerCall, "(call (call <foo>))")
	})
}
