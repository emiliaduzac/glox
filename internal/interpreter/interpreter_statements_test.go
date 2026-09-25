package interpreter

import (
	"testing"
	"tp-lox/internal/expression"
	"tp-lox/internal/statements"
	"tp-lox/internal/token"
)

func TestExecuteUnsupportedStatement(t *testing.T) {
	in := SetUpInterpreter(t)

	t.Run("Error when statement is nil", func(t *testing.T) {
		err := in.Execute(nil)
		AssertError(t, err)
	})

	t.Run("Error when statement type is unknown", func(t *testing.T) {
		err := in.Execute(UnsupportedStmt{})
		AssertError(t, err)
	})
}

func TestExecuteExpressionStatement(t *testing.T) {
	in := SetUpInterpreter(t)

	t.Run("Execute expression statement with literal", func(t *testing.T) {
		stmt := statements.NewExpressionStmt(expression.NewLiteral(99.0))
		err := in.Execute(stmt)
		AssertNoError(t, err)
	})

	t.Run("Execute expression statement with binary expression", func(t *testing.T) {
		op := *token.NewToken(token.PLUS, "+", nil, 1)
		stmt := statements.NewExpressionStmt(expression.NewBinary(expression.NewLiteral(1.0), op, expression.NewLiteral(2.0)))
		err := in.Execute(stmt)
		AssertNoError(t, err)
	})

	t.Run("Error when expression is invalid", func(t *testing.T) {
		stmt := statements.NewExpressionStmt(invalidExpr())
		err := in.Execute(stmt)
		AssertError(t, err)
	})
}

func TestExecutePrintStatementErrors(t *testing.T) {
	in := SetUpInterpreter(t)

	t.Run("Error when expression is invalid", func(t *testing.T) {
		stmt := statements.NewPrintStmt(invalidExpr())
		err := in.Execute(stmt)
		AssertError(t, err)
	})
}

func TestInterpret(t *testing.T) {
	in := SetUpInterpreter(t)

	t.Run("No error with empty list of statements", func(t *testing.T) {
		err := in.Interpret([]statements.Stmt{})
		AssertNoError(t, err)
	})

	t.Run("No error with multiple valid statements", func(t *testing.T) {
		stmts := []statements.Stmt{
			statements.NewExpressionStmt(expression.NewLiteral(1.0)),
			statements.NewExpressionStmt(expression.NewLiteral("a")),
		}
		err := in.Interpret(stmts)
		AssertNoError(t, err)
	})

	t.Run("Error when one statement is invalid", func(t *testing.T) {
		stmts := []statements.Stmt{
			statements.NewExpressionStmt(expression.NewLiteral(1.0)),
			statements.NewExpressionStmt(invalidExpr()),
		}
		err := in.Interpret(stmts)
		AssertError(t, err)
	})

	t.Run("Error when statement type is unknown", func(t *testing.T) {
		err := in.Interpret([]statements.Stmt{UnsupportedStmt{}})
		AssertError(t, err)
	})
}
