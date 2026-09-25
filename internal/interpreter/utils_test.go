package interpreter

import (
	"testing"
	"tp-lox/internal/expression"
	"tp-lox/internal/token"
)

func SetUpInterpreter(t *testing.T) *Interpreter {
	t.Helper()
	return NewInterpreter()
}

func AssertValue(t *testing.T, got any, expected any) {
	t.Helper()
	if got != expected {
		t.Errorf("got %v, expected %v", got, expected)
	}
}

func AssertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func AssertError(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

type UnsupportedExpr struct{}

func (u UnsupportedExpr) Expression() {}

type UnsupportedStmt struct{}

func (u UnsupportedStmt) Statement() {}

// invalidExpr devuelve una expresión que falla al evaluarse.
func invalidExpr() expression.Expr {
	return expression.NewUnary(*token.NewToken(token.MINUS, "-", nil, 1), expression.NewLiteral("a"))
}
