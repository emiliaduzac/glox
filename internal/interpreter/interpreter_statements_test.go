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

func TestExecutePrintStatement(t *testing.T) {
	in := SetUpInterpreter(t)

	t.Run("Print a string literal", func(t *testing.T) {
		stmt := statements.NewPrintStmt(expression.NewLiteral("Hello"))
		out := CaptureOutput(t, func() {
			AssertNoError(t, in.Execute(stmt))
		})
		AssertValue(t, out, "Hello\n")
	})

	t.Run("Print a binary expression", func(t *testing.T) {
		op := *token.NewToken(token.PLUS, "+", nil, 1)
		stmt := statements.NewPrintStmt(expression.NewBinary(expression.NewLiteral(1.0), op, expression.NewLiteral(2.0)))
		out := CaptureOutput(t, func() {
			AssertNoError(t, in.Execute(stmt))
		})
		AssertValue(t, out, "3\n")
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

func TestExecuteVarStatement(t *testing.T) {
	in := SetUpInterpreter(t)

	t.Run("Declare variable with initializer", func(t *testing.T) {
		token := *token.NewToken(token.IDENTIFIER, "x", nil, 1)
		stmt := statements.NewVarStmt(token, expression.NewLiteral(18.0))
		AssertNoError(t, in.Execute(stmt))
		val, err := in.Evaluate(expression.NewVariable(token))
		AssertNoError(t, err)
		AssertValue(t, val, 18.0)
	})

	t.Run("Declare variable witout initializer defaults to nil", func(t *testing.T) {
		token := *token.NewToken(token.IDENTIFIER, "x", nil, 1)
		stmt := statements.NewVarStmt(token, nil)
		AssertNoError(t, in.Execute(stmt))
		val, err := in.Evaluate(expression.NewVariable(token))
		AssertNoError(t, err)
		AssertValue(t, val, nil)
	})

	t.Run("Redeclare variable with different type", func(t *testing.T) {
		token := *token.NewToken(token.IDENTIFIER, "a", nil, 1)
		AssertNoError(t, in.Execute(statements.NewVarStmt(token, expression.NewLiteral(10.0))))
		AssertNoError(t, in.Execute(statements.NewVarStmt(token, expression.NewLiteral("Hello"))))

		val, err := in.Evaluate(expression.NewVariable(token))
		AssertNoError(t, err)
		AssertValue(t, val, "Hello")
	})

	t.Run("Error when initializer expression is invalid", func(t *testing.T) {
		token := *token.NewToken(token.IDENTIFIER, "invalid", nil, 1)
		stmt := statements.NewVarStmt(token, invalidExpr())

		AssertError(t, in.Execute(stmt))
		_, err := in.Evaluate(expression.NewVariable(token))
		AssertError(t, err)
	})

	t.Run("Error when evaluating undeclared variable", func(t *testing.T) {
		token := *token.NewToken(token.IDENTIFIER, "undeclared", nil, 1)
		_, err := in.Evaluate(expression.NewVariable(token))
		AssertError(t, err)
	})
}
