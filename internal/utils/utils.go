package utils

import (
	"fmt"
	"strings"
	"testing"
	"tp-lox/internal/expression"
)

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

// Funciones comunes para los tests

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
