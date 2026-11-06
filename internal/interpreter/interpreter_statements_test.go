package interpreter

import (
	"testing"
	"tp-lox/internal/expression"
	"tp-lox/internal/statements"
	"tp-lox/internal/token"
	"tp-lox/internal/utils"
)

// Tests de Execute para cada statement (Print, Expression, Var, Block, If)

func TestExecuteUnsupportedStatement(t *testing.T) {
	in := SetUpInterpreter(t)

	t.Run("Error when statement is nil", func(t *testing.T) {
		err := in.Execute(nil)
		utils.AssertError(t, err)
	})

	t.Run("Error when statement type is unknown", func(t *testing.T) {
		err := in.Execute(utils.UnsupportedStmt{})
		utils.AssertError(t, err)
	})
}

func TestExecutePrintStatement(t *testing.T) {
	in := SetUpInterpreter(t)

	t.Run("Print a string literal", func(t *testing.T) {
		stmt := statements.NewPrintStmt(expression.NewLiteral("Hello"))
		out := CaptureOutput(t, func() {
			utils.AssertNoError(t, in.Execute(stmt))
		})
		utils.AssertValue(t, out, "Hello\n")
	})

	t.Run("Print a binary expression", func(t *testing.T) {
		op := *token.NewToken(token.PLUS, "+", nil, 1)
		stmt := statements.NewPrintStmt(expression.NewBinary(expression.NewLiteral(1.0), op, expression.NewLiteral(2.0)))
		out := CaptureOutput(t, func() {
			utils.AssertNoError(t, in.Execute(stmt))
		})
		utils.AssertValue(t, out, "3\n")
	})
}

func TestExecutePrintStatementErrors(t *testing.T) {
	in := SetUpInterpreter(t)

	t.Run("Error when expression is invalid", func(t *testing.T) {
		stmt := statements.NewPrintStmt(invalidExpr())
		err := in.Execute(stmt)
		utils.AssertError(t, err)
	})
}

func TestExecuteExpressionStatement(t *testing.T) {
	in := SetUpInterpreter(t)

	t.Run("Execute expression statement with literal", func(t *testing.T) {
		stmt := statements.NewExpressionStmt(expression.NewLiteral(99.0))
		err := in.Execute(stmt)
		utils.AssertNoError(t, err)
	})

	t.Run("Execute expression statement with binary expression", func(t *testing.T) {
		op := *token.NewToken(token.PLUS, "+", nil, 1)
		stmt := statements.NewExpressionStmt(expression.NewBinary(expression.NewLiteral(1.0), op, expression.NewLiteral(2.0)))
		err := in.Execute(stmt)
		utils.AssertNoError(t, err)
	})

	t.Run("Error when expression is invalid", func(t *testing.T) {
		stmt := statements.NewExpressionStmt(invalidExpr())
		err := in.Execute(stmt)
		utils.AssertError(t, err)
	})
}

func TestExecuteVarStatement(t *testing.T) {
	in := SetUpInterpreter(t)

	t.Run("Declare variable with initializer", func(t *testing.T) {
		token := *token.NewToken(token.IDENTIFIER, "x", nil, 1)
		stmt := statements.NewVarStmt(token, expression.NewLiteral(18.0))
		utils.AssertNoError(t, in.Execute(stmt))
		val, err := in.Evaluate(expression.NewVariable(token))
		utils.AssertNoError(t, err)
		utils.AssertValue(t, val, 18.0)
	})

	t.Run("Declare variable witout initializer defaults to nil", func(t *testing.T) {
		token := *token.NewToken(token.IDENTIFIER, "x", nil, 1)
		stmt := statements.NewVarStmt(token, nil)
		utils.AssertNoError(t, in.Execute(stmt))
		val, err := in.Evaluate(expression.NewVariable(token))
		utils.AssertNoError(t, err)
		utils.AssertValue(t, val, nil)
	})

	t.Run("Redeclare variable with different type", func(t *testing.T) {
		token := *token.NewToken(token.IDENTIFIER, "a", nil, 1)
		utils.AssertNoError(t, in.Execute(statements.NewVarStmt(token, expression.NewLiteral(10.0))))
		utils.AssertNoError(t, in.Execute(statements.NewVarStmt(token, expression.NewLiteral("Hello"))))

		val, err := in.Evaluate(expression.NewVariable(token))
		utils.AssertNoError(t, err)
		utils.AssertValue(t, val, "Hello")
	})

	t.Run("Error when initializer expression is invalid", func(t *testing.T) {
		token := *token.NewToken(token.IDENTIFIER, "invalid", nil, 1)
		stmt := statements.NewVarStmt(token, invalidExpr())

		utils.AssertError(t, in.Execute(stmt))
		_, err := in.Evaluate(expression.NewVariable(token))
		utils.AssertError(t, err)
	})
}

func TestExecuteBlockStatement(t *testing.T) {
	t.Run("Empty block executes successfully", func(t *testing.T) {
		in := SetUpInterpreter(t)
		stmt := statements.NewBlockStmt([]statements.Stmt{})
		utils.AssertNoError(t, in.Execute(stmt))
	})

	t.Run("Block local variable is inaccessible outside block", func(t *testing.T) {
		in := SetUpInterpreter(t)
		tokLocal := *token.NewToken(token.IDENTIFIER, "localOnly", nil, 1)
		innerVar := statements.NewVarStmt(tokLocal, expression.NewLiteral(42.0))
		blockStmt := statements.NewBlockStmt([]statements.Stmt{innerVar})

		utils.AssertNoError(t, in.Execute(blockStmt))

		_, err := in.Evaluate(expression.NewVariable(tokLocal))
		utils.AssertErrorMessage(t, err, "undefined variable 'localOnly'")
	})

	t.Run("Block mutates outer variable and persists outside", func(t *testing.T) {
		in := SetUpInterpreter(t)
		tokCount := *token.NewToken(token.IDENTIFIER, "count", nil, 1)
		utils.AssertNoError(t, in.Execute(statements.NewVarStmt(tokCount, expression.NewLiteral(1.0))))

		assignStmt := statements.NewExpressionStmt(expression.NewAssign(tokCount, expression.NewLiteral(2.0)))
		blockStmt := statements.NewBlockStmt([]statements.Stmt{assignStmt})

		utils.AssertNoError(t, in.Execute(blockStmt))

		val, err := in.Evaluate(expression.NewVariable(tokCount))
		utils.AssertNoError(t, err)
		utils.AssertValue(t, val, 2.0)
	})

	t.Run("Runtime error inside block restores previous environment", func(t *testing.T) {
		in := SetUpInterpreter(t)
		tokG := *token.NewToken(token.IDENTIFIER, "globalVar", nil, 1)
		utils.AssertNoError(t, in.Execute(statements.NewVarStmt(tokG, expression.NewLiteral("survives"))))

		tokL := *token.NewToken(token.IDENTIFIER, "temp", nil, 1)
		localStmt := statements.NewVarStmt(tokL, expression.NewLiteral(99.0))
		failingStmt := statements.NewExpressionStmt(invalidExpr())
		blockStmt := statements.NewBlockStmt([]statements.Stmt{localStmt, failingStmt})

		err := in.Execute(blockStmt)
		utils.AssertError(t, err)

		val, err := in.Evaluate(expression.NewVariable(tokG))
		utils.AssertNoError(t, err)
		utils.AssertValue(t, val, "survives")

		_, err = in.Evaluate(expression.NewVariable(tokL))
		utils.AssertErrorMessage(t, err, "undefined variable 'temp'")
	})
}

func TestExecuteIfStatement(t *testing.T) {
	t.Run("True condition executes then branch only", func(t *testing.T) {
		in := SetUpInterpreter(t)
		stmt := statements.NewIfStmt(expression.NewLiteral(true), printStmt("then"), printStmt("else"))
		out := CaptureOutput(t, func() {
			utils.AssertNoError(t, in.Execute(stmt))
		})
		utils.AssertValue(t, out, "then\n")
	})

	t.Run("False condition executes else branch only", func(t *testing.T) {
		in := SetUpInterpreter(t)
		stmt := statements.NewIfStmt(expression.NewLiteral(false), printStmt("then"), printStmt("else"))
		out := CaptureOutput(t, func() {
			utils.AssertNoError(t, in.Execute(stmt))
		})
		utils.AssertValue(t, out, "else\n")
	})

	t.Run("False condition without else does nothing", func(t *testing.T) {
		in := SetUpInterpreter(t)
		stmt := statements.NewIfStmt(expression.NewLiteral(false), printStmt("then"), nil)
		out := CaptureOutput(t, func() {
			utils.AssertNoError(t, in.Execute(stmt))
		})
		utils.AssertValue(t, out, "")
	})

	t.Run("Error in condition is returned and no branch runs", func(t *testing.T) {
		in := SetUpInterpreter(t)
		stmt := statements.NewIfStmt(invalidExpr(), printStmt("then"), printStmt("else"))
		out := CaptureOutput(t, func() {
			utils.AssertError(t, in.Execute(stmt))
		})
		utils.AssertValue(t, out, "")
	})

	t.Run("Error in executed branch is returned", func(t *testing.T) {
		in := SetUpInterpreter(t)
		failingStmt := statements.NewExpressionStmt(invalidExpr())
		stmt := statements.NewIfStmt(expression.NewLiteral(true), failingStmt, nil)
		utils.AssertError(t, in.Execute(stmt))
	})

	t.Run("Else if executes the matching branch only", func(t *testing.T) {
		in := SetUpInterpreter(t)
		elseIf := statements.NewIfStmt(expression.NewLiteral(true), printStmt("second"), printStmt("third"))
		stmt := statements.NewIfStmt(expression.NewLiteral(false), printStmt("first"), elseIf)
		out := CaptureOutput(t, func() {
			utils.AssertNoError(t, in.Execute(stmt))
		})
		utils.AssertValue(t, out, "second\n")
	})
}

func TestExecuteWhileStatement(t *testing.T) {
	t.Run("Loop runs while condition holds and updates outer variable across iterations", func(t *testing.T) {
		in := SetUpInterpreter(t)
		tokCount := *token.NewToken(token.IDENTIFIER, "count", nil, 1)
		utils.AssertNoError(t, in.Execute(statements.NewVarStmt(tokCount, expression.NewLiteral(0.0))))

		condition := expression.NewBinary(
			expression.NewVariable(tokCount),
			*token.NewToken(token.LESS, "<", nil, 1),
			expression.NewLiteral(3.0),
		)
		body := statements.NewBlockStmt([]statements.Stmt{
			statements.NewPrintStmt(expression.NewVariable(tokCount)),
			statements.NewExpressionStmt(expression.NewAssign(
				tokCount,
				expression.NewBinary(expression.NewVariable(tokCount), *token.NewToken(token.PLUS, "+", nil, 1), expression.NewLiteral(1.0)),
			)),
		})
		stmt := statements.NewWhileStmt(condition, body)

		out := CaptureOutput(t, func() {
			utils.AssertNoError(t, in.Execute(stmt))
		})
		utils.AssertValue(t, out, "0\n1\n2\n")

		val, err := in.Evaluate(expression.NewVariable(tokCount))
		utils.AssertNoError(t, err)
		utils.AssertValue(t, val, 3.0)
	})

	t.Run("False condition from the start executes zero iterations", func(t *testing.T) {
		in := SetUpInterpreter(t)
		stmt := statements.NewWhileStmt(expression.NewLiteral(false), printStmt("never"))
		out := CaptureOutput(t, func() {
			utils.AssertNoError(t, in.Execute(stmt))
		})
		utils.AssertValue(t, out, "")
	})

	t.Run("Error in condition is returned and body never runs", func(t *testing.T) {
		in := SetUpInterpreter(t)
		stmt := statements.NewWhileStmt(invalidExpr(), printStmt("never"))
		out := CaptureOutput(t, func() {
			utils.AssertError(t, in.Execute(stmt))
		})
		utils.AssertValue(t, out, "")
	})

	t.Run("Error in body stops the loop and is returned", func(t *testing.T) {
		in := SetUpInterpreter(t)
		body := statements.NewBlockStmt([]statements.Stmt{
			printStmt("loop"),
			statements.NewExpressionStmt(invalidExpr()),
		})
		stmt := statements.NewWhileStmt(expression.NewLiteral(true), body)

		out := CaptureOutput(t, func() {
			utils.AssertError(t, in.Execute(stmt))
		})
		utils.AssertValue(t, out, "loop\n")
	})
}

func TestExecuteFunctionStatement(t *testing.T) {
	t.Run("Declare function binds callable with name and arity to environment", func(t *testing.T) {
		in := SetUpInterpreter(t)
		tokFn := *token.NewToken(token.IDENTIFIER, "sayHello", nil, 1)
		fnStmt := statements.NewFunctionStmt(
			tokFn,
			nil,
			[]statements.Stmt{printStmt("Hello")},
		)
		utils.AssertNoError(t, in.Execute(fnStmt))

		val, err := in.Evaluate(expression.NewVariable(tokFn))
		utils.AssertNoError(t, err)
		callable, ok := val.(Callable)
		if !ok {
			t.Fatalf("expected Callable in environment, got %T", val)
		}
		utils.AssertValue(t, callable.String(), "<fn sayHello>")
		utils.AssertValue(t, callable.Arity(), 0)
	})

	t.Run("Function parameters are not accessible outside function scope", func(t *testing.T) {
		in := SetUpInterpreter(t)
		tokFn := *token.NewToken(token.IDENTIFIER, "f", nil, 1)
		tokParam := *token.NewToken(token.IDENTIFIER, "x", nil, 1)
		fnStmt := statements.NewFunctionStmt(
			tokFn,
			[]token.Token{tokParam},
			[]statements.Stmt{},
		)
		utils.AssertNoError(t, in.Execute(fnStmt))

		call := expression.NewCall(
			expression.NewVariable(tokFn),
			*token.NewToken(token.RIGHT_PAREN, ")", nil, 1),
			[]expression.Expr{expression.NewLiteral(14.0)},
		)
		_, err := in.Evaluate(call)
		utils.AssertNoError(t, err)

		_, err = in.Evaluate(expression.NewVariable(tokParam))
		utils.AssertErrorMessage(t, err, "undefined variable 'x'")
	})

	t.Run("Error when invoking function with incorrect arity", func(t *testing.T) {
		in := SetUpInterpreter(t)
		tokFn := *token.NewToken(token.IDENTIFIER, "oneArg", nil, 1)
		tokParam := *token.NewToken(token.IDENTIFIER, "a", nil, 1)
		fnStmt := statements.NewFunctionStmt(
			tokFn,
			[]token.Token{tokParam},
			[]statements.Stmt{},
		)
		utils.AssertNoError(t, in.Execute(fnStmt))

		call := expression.NewCall(
			expression.NewVariable(tokFn),
			*token.NewToken(token.RIGHT_PAREN, ")", nil, 1),
			nil,
		)
		_, err := in.Evaluate(call)
		utils.AssertErrorMessage(t, err, "expected 1 arguments but got 0")
	})

	t.Run("Error inside function body propagates to caller", func(t *testing.T) {
		in := SetUpInterpreter(t)
		tokFn := *token.NewToken(token.IDENTIFIER, "failing", nil, 1)
		fnStmt := statements.NewFunctionStmt(
			tokFn,
			nil,
			[]statements.Stmt{
				statements.NewExpressionStmt(invalidExpr()),
			},
		)
		utils.AssertNoError(t, in.Execute(fnStmt))

		call := expression.NewCall(
			expression.NewVariable(tokFn),
			*token.NewToken(token.RIGHT_PAREN, ")", nil, 1),
			nil,
		)
		_, err := in.Evaluate(call)
		utils.AssertError(t, err)
	})
}

func TestExecuteReturnStatement(t *testing.T) {
	t.Run("Direct Execute of ReturnStmt emits ReturnValue", func(t *testing.T) {
		in := SetUpInterpreter(t)
		stmt := statements.NewReturnStmt(
			*token.NewToken(token.RETURN, "return", nil, 1),
			expression.NewLiteral(17.0),
		)
		err := in.Execute(stmt)
		ret, ok := err.(ReturnValue)
		if !ok {
			t.Fatalf("expected ReturnValue, got %T (%v)", err, err)
		}
		utils.AssertValue(t, ret.Value, 17.0)
	})

	t.Run("Direct Execute of ReturnStmt with failing expression returns runtime error", func(t *testing.T) {
		in := SetUpInterpreter(t)
		stmt := statements.NewReturnStmt(
			*token.NewToken(token.RETURN, "return", nil, 1),
			invalidExpr(),
		)
		err := in.Execute(stmt)
		utils.AssertErrorMessage(t, err, "operand must be a number")
	})

	t.Run("Return with explicit value returns it to caller", func(t *testing.T) {
		in := SetUpInterpreter(t)
		tokFn := *token.NewToken(token.IDENTIFIER, "getAnswer", nil, 1)
		fnStmt := statements.NewFunctionStmt(
			tokFn,
			nil,
			[]statements.Stmt{
				statements.NewReturnStmt(
					*token.NewToken(token.RETURN, "return", nil, 1),
					expression.NewLiteral(6.0),
				),
			},
		)
		utils.AssertNoError(t, in.Execute(fnStmt))

		call := expression.NewCall(
			expression.NewVariable(tokFn),
			*token.NewToken(token.RIGHT_PAREN, ")", nil, 1),
			nil,
		)
		val, err := in.Evaluate(call)
		utils.AssertNoError(t, err)
		utils.AssertValue(t, val, 6.0)
	})

	t.Run("Return without value yields nil to caller", func(t *testing.T) {
		in := SetUpInterpreter(t)
		tokFn := *token.NewToken(token.IDENTIFIER, "emptyReturn", nil, 1)
		fnStmt := statements.NewFunctionStmt(
			tokFn,
			nil,
			[]statements.Stmt{
				statements.NewReturnStmt(
					*token.NewToken(token.RETURN, "return", nil, 1),
					nil,
				),
			},
		)
		utils.AssertNoError(t, in.Execute(fnStmt))

		call := expression.NewCall(
			expression.NewVariable(tokFn),
			*token.NewToken(token.RIGHT_PAREN, ")", nil, 1),
			nil,
		)
		val, err := in.Evaluate(call)
		utils.AssertNoError(t, err)
		utils.AssertValue(t, val, nil)
	})

	t.Run("Return unwinds block and skips subsequent statements", func(t *testing.T) {
		in := SetUpInterpreter(t)
		tokFn := *token.NewToken(token.IDENTIFIER, "earlyExit", nil, 1)
		fnStmt := statements.NewFunctionStmt(
			tokFn,
			nil,
			[]statements.Stmt{
				statements.NewBlockStmt([]statements.Stmt{
					statements.NewReturnStmt(
						*token.NewToken(token.RETURN, "return", nil, 1),
						expression.NewLiteral("early"),
					),
				}),
				printStmt("unreachable"),
			},
		)
		utils.AssertNoError(t, in.Execute(fnStmt))

		call := expression.NewCall(
			expression.NewVariable(tokFn),
			*token.NewToken(token.RIGHT_PAREN, ")", nil, 1),
			nil,
		)
		out := CaptureOutput(t, func() {
			val, err := in.Evaluate(call)
			utils.AssertNoError(t, err)
			utils.AssertValue(t, val, "early")
		})
		utils.AssertValue(t, out, "")
	})

	t.Run("Return breaks while loop immediately", func(t *testing.T) {
		in := SetUpInterpreter(t)
		tokFn := *token.NewToken(token.IDENTIFIER, "loopReturn", nil, 1)
		fnStmt := statements.NewFunctionStmt(
			tokFn,
			nil,
			[]statements.Stmt{
				statements.NewWhileStmt(
					expression.NewLiteral(true),
					statements.NewReturnStmt(
						*token.NewToken(token.RETURN, "return", nil, 1),
						expression.NewLiteral(93.0),
					),
				),
			},
		)
		utils.AssertNoError(t, in.Execute(fnStmt))

		call := expression.NewCall(
			expression.NewVariable(tokFn),
			*token.NewToken(token.RIGHT_PAREN, ")", nil, 1),
			nil,
		)
		val, err := in.Evaluate(call)
		utils.AssertNoError(t, err)
		utils.AssertValue(t, val, 93.0)
	})

	t.Run("Error during return expression evaluation propagates to caller", func(t *testing.T) {
		in := SetUpInterpreter(t)
		tokFn := *token.NewToken(token.IDENTIFIER, "failingReturn", nil, 1)
		fnStmt := statements.NewFunctionStmt(
			tokFn,
			nil,
			[]statements.Stmt{
				statements.NewReturnStmt(
					*token.NewToken(token.RETURN, "return", nil, 1),
					invalidExpr(),
				),
			},
		)
		utils.AssertNoError(t, in.Execute(fnStmt))

		call := expression.NewCall(
			expression.NewVariable(tokFn),
			*token.NewToken(token.RIGHT_PAREN, ")", nil, 1),
			nil,
		)
		_, err := in.Evaluate(call)
		utils.AssertErrorMessage(t, err, "operand must be a number")
	})
}

func TestInterpret(t *testing.T) {
	in := SetUpInterpreter(t)

	t.Run("No error with empty list of statements", func(t *testing.T) {
		err := in.Interpret([]statements.Stmt{})
		utils.AssertNoError(t, err)
	})

	t.Run("No error with multiple valid statements", func(t *testing.T) {
		stmts := []statements.Stmt{
			statements.NewExpressionStmt(expression.NewLiteral(1.0)),
			statements.NewExpressionStmt(expression.NewLiteral("a")),
		}
		err := in.Interpret(stmts)
		utils.AssertNoError(t, err)
	})

	t.Run("Error when one statement is invalid", func(t *testing.T) {
		stmts := []statements.Stmt{
			statements.NewExpressionStmt(expression.NewLiteral(1.0)),
			statements.NewExpressionStmt(invalidExpr()),
		}
		err := in.Interpret(stmts)
		utils.AssertError(t, err)
	})

	t.Run("Error when statement type is unknown", func(t *testing.T) {
		err := in.Interpret([]statements.Stmt{utils.UnsupportedStmt{}})
		utils.AssertError(t, err)
	})
}
