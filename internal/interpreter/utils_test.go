package interpreter

import (
	"io"
	"os"
	"testing"
	"tp-lox/internal/expression"
	"tp-lox/internal/token"
)

func CaptureOutput(t *testing.T, f func()) string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("could not create pipe: %v", err)
	}
	defer func() {
		if err := r.Close(); err != nil {
			t.Errorf("could not close read pipe: %v", err)
		}
	}()

	func() {
		defer func() {
			os.Stdout = old
			if err := w.Close(); err != nil {
				t.Errorf("could not close write pipe: %v", err)
			}
		}()
		os.Stdout = w
		f()
	}()

	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("could not read captured output: %v", err)
	}
	return string(out)
}

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
