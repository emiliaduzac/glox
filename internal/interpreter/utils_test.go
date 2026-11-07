package interpreter

import (
	"io"
	"os"
	"testing"
	"tp-lox/internal/expression"
	"tp-lox/internal/statements"
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

// invalidExpr devuelve una expresión que falla al evaluarse.
func invalidExpr() expression.Expr {
	return expression.NewUnary(*token.NewToken(token.MINUS, "-", nil, 1), expression.NewLiteral("a"))
}

// printStmt devuelve un print statement que imprime el string dado.
func printStmt(msg string) statements.Stmt {
	return statements.NewPrintStmt(expression.NewLiteral(msg))
}
