package utils

import (
	"fmt"
	"strings"
	"testing"
	"tp-lox/internal/errors"
	"tp-lox/internal/expression"
)

// -- INTERPRETER & VM --

func IsEqual(a, b any) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil {
		return false
	}
	return a == b
}

func IsTruthy(value any) bool {
	if value == nil {
		return false
	}
	if b, ok := value.(bool); ok {
		return b
	}
	return true
}

func Stringify(value any) string {
	if value == nil {
		return "nil"
	}
	return fmt.Sprintf("%v", value)
}

// DebugString muestra un valor para las trazas de debug, con comillas simples si es string
func DebugString(value any) string {
	if s, ok := value.(string); ok {
		return "'" + s + "'"
	}
	return Stringify(value)
}

func CheckNumberOperand(line int, operand any) (float64, error) {
	num, ok := operand.(float64)
	if !ok {
		return 0, errors.NewLoxError(line, "operand must be a number")
	}
	return num, nil
}

func CheckNumberOperands(line int, left, right any) (float64, float64, error) {
	leftNum, okLeft := left.(float64)
	rightNum, okRight := right.(float64)
	if !okLeft || !okRight {
		return 0, 0, errors.NewLoxError(line, "operands must be numbers")
	}
	return leftNum, rightNum, nil
}

// -- TESTS --

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

func AssertErrorMessage(t *testing.T, err error, expected string) {
	t.Helper()
	AssertError(t, err)
	if !strings.Contains(err.Error(), expected) {
		t.Fatalf("expected error containing %q, got %q", expected, err.Error())
	}
}

func AssertString(t *testing.T, expr expression.Expr, expected string) {
	t.Helper()

	if got := expression.String(expr); got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

type UnsupportedStmt struct{}

func (u UnsupportedStmt) Statement() {}

type UnsupportedExpr struct{}

func (u UnsupportedExpr) Expression() {}
