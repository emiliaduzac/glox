package interpreter

import (
	"math"
	"testing"
	"tp-lox/internal/environment"
	"tp-lox/internal/errors"
	"tp-lox/internal/expression"
	"tp-lox/internal/statements"
	"tp-lox/internal/token"
	"tp-lox/internal/utils"
)

// Tests de Evaluate para cada construcción de expresión (Literal, Grouping,
// Unary, Binary, Variable, Assign)

func TestEvaluateUnsupportedExpression(t *testing.T) {
	in := SetUpInterpreter(t)

	t.Run("Error when expression is nil", func(t *testing.T) {
		_, err := in.Evaluate(nil)
		utils.AssertError(t, err)
	})

	t.Run("Error when expression type is unknown", func(t *testing.T) {
		_, err := in.Evaluate(utils.UnsupportedExpr{})
		utils.AssertError(t, err)
	})
}

func TestEvaluateLiterals(t *testing.T) {
	in := SetUpInterpreter(t)

	t.Run("Evaluate positive integer literal", func(t *testing.T) {
		expression := expression.NewLiteral(99.0)
		value, err := in.Evaluate(expression)
		utils.AssertNoError(t, err)
		utils.AssertValue(t, value, 99.0)
	})

	t.Run("Evaluate empty string literal", func(t *testing.T) {
		expression := expression.NewLiteral("")
		value, err := in.Evaluate(expression)
		utils.AssertNoError(t, err)
		utils.AssertValue(t, value, "")
	})

	t.Run("Evaluate string literal", func(t *testing.T) {
		expression := expression.NewLiteral("Hello World")
		value, err := in.Evaluate(expression)
		utils.AssertNoError(t, err)
		utils.AssertValue(t, value, "Hello World")
	})

	t.Run("Evaluate boolean literal", func(t *testing.T) {
		expression := expression.NewLiteral(true)
		value, err := in.Evaluate(expression)
		utils.AssertNoError(t, err)
		utils.AssertValue(t, value, true)
	})

	t.Run("Evaluate false boolean literal", func(t *testing.T) {
		expression := expression.NewLiteral(false)
		value, err := in.Evaluate(expression)
		utils.AssertNoError(t, err)
		utils.AssertValue(t, value, false)
	})

	t.Run("Evaluate nil literal", func(t *testing.T) {
		expression := expression.NewLiteral(nil)
		value, err := in.Evaluate(expression)
		utils.AssertNoError(t, err)
		utils.AssertValue(t, value, nil)
	})
}

func TestEvaluateGrouping(t *testing.T) {
	in := SetUpInterpreter(t)

	t.Run("Evaluate grouping with literal", func(t *testing.T) {
		expression := expression.NewGrouping(expression.NewLiteral(99))
		value, err := in.Evaluate(expression)
		utils.AssertNoError(t, err)
		utils.AssertValue(t, value, 99)
	})

	t.Run("Evaluate gruped string literal", func(t *testing.T) {
		expression := expression.NewGrouping(expression.NewLiteral("Hello"))
		value, err := in.Evaluate(expression)
		utils.AssertNoError(t, err)
		utils.AssertValue(t, value, "Hello")
	})

	t.Run("Evaluate nested grouping with literal", func(t *testing.T) {
		expression := expression.NewGrouping(expression.NewGrouping(expression.NewGrouping(expression.NewLiteral(88))))
		value, err := in.Evaluate(expression)
		utils.AssertNoError(t, err)
		utils.AssertValue(t, value, 88)
	})
}

func TestEvaluateUnaryMinus(t *testing.T) {
	in := SetUpInterpreter(t)

	t.Run("Negate positive number", func(t *testing.T) {
		expr := expression.NewUnary(*token.NewToken(token.MINUS, "-", nil, 1), expression.NewLiteral(77.0))
		val, err := in.Evaluate(expr)
		utils.AssertNoError(t, err)
		utils.AssertValue(t, val, -77.0)
	})

	t.Run("Negate negative number", func(t *testing.T) {
		expr := expression.NewUnary(*token.NewToken(token.MINUS, "-", nil, 1), expression.NewLiteral(-77.0))
		val, err := in.Evaluate(expr)
		utils.AssertNoError(t, err)
		utils.AssertValue(t, val, 77.0)
	})

	t.Run("Negate zero", func(t *testing.T) {
		expr := expression.NewUnary(*token.NewToken(token.MINUS, "-", nil, 1), expression.NewLiteral(0.0))
		val, err := in.Evaluate(expr)
		utils.AssertNoError(t, err)
		utils.AssertValue(t, val, 0.0)
	})

	t.Run("Error when negating string", func(t *testing.T) {
		expr := expression.NewUnary(*token.NewToken(token.MINUS, "-", nil, 1), expression.NewLiteral("hello"))
		_, err := in.Evaluate(expr)
		utils.AssertError(t, err)
	})

	t.Run("Error when negating boolean", func(t *testing.T) {
		expr := expression.NewUnary(*token.NewToken(token.MINUS, "-", nil, 1), expression.NewLiteral(true))
		_, err := in.Evaluate(expr)
		utils.AssertError(t, err)
	})

	t.Run("Error when negating nil", func(t *testing.T) {
		expr := expression.NewUnary(*token.NewToken(token.MINUS, "-", nil, 1), expression.NewLiteral(nil))
		_, err := in.Evaluate(expr)
		utils.AssertError(t, err)
	})
}

func TestCheckNumberOperand(t *testing.T) {
	op := *token.NewToken(token.MINUS, "-", nil, 1)

	t.Run("Operand is a number", func(t *testing.T) {
		num, err := checkNumberOperand(op, 10.0)
		utils.AssertNoError(t, err)
		utils.AssertValue(t, num, 10.0)
	})

	t.Run("Operand is not a number", func(t *testing.T) {
		_, err := checkNumberOperand(op, "not a number")
		utils.AssertError(t, err)
	})
}

func TestEvaluateUnaryBang(t *testing.T) {
	in := SetUpInterpreter(t)

	t.Run("Bang boolean values", func(t *testing.T) {
		exprTrue := expression.NewUnary(*token.NewToken(token.BANG, "!", nil, 1), expression.NewLiteral(true))
		val, err := in.Evaluate(exprTrue)
		utils.AssertNoError(t, err)
		utils.AssertValue(t, val, false)

		exprFalse := expression.NewUnary(*token.NewToken(token.BANG, "!", nil, 1), expression.NewLiteral(false))
		val, err = in.Evaluate(exprFalse)
		utils.AssertNoError(t, err)
		utils.AssertValue(t, val, true)
	})

	t.Run("Bang nil", func(t *testing.T) {
		expr := expression.NewUnary(*token.NewToken(token.BANG, "!", nil, 1), expression.NewLiteral(nil))
		val, err := in.Evaluate(expr)
		utils.AssertNoError(t, err)
		utils.AssertValue(t, val, true)
	})

	t.Run("Bang truthy non-booleans", func(t *testing.T) {
		exprZero := expression.NewUnary(*token.NewToken(token.BANG, "!", nil, 1), expression.NewLiteral(0.0))
		val, err := in.Evaluate(exprZero)
		utils.AssertNoError(t, err)
		utils.AssertValue(t, val, false)

		exprStr := expression.NewUnary(*token.NewToken(token.BANG, "!", nil, 1), expression.NewLiteral(""))
		val, err = in.Evaluate(exprStr)
		utils.AssertNoError(t, err)
		utils.AssertValue(t, val, false)
	})

	t.Run("Propagates error from inner expression", func(t *testing.T) {
		inner := expression.NewUnary(*token.NewToken(token.MINUS, "-", nil, 1), expression.NewLiteral("hello"))
		expr := expression.NewUnary(*token.NewToken(token.BANG, "!", nil, 1), inner)
		_, err := in.Evaluate(expr)
		utils.AssertError(t, err)
	})

	t.Run("Error when unary operator is unknown", func(t *testing.T) {
		expr := expression.NewUnary(*token.NewToken(token.EOF, "EOF", nil, 1), expression.NewLiteral(10.0))
		_, err := in.Evaluate(expr)
		utils.AssertError(t, err)
	})
}

func TestEvaluateBinary(t *testing.T) {
	in := SetUpInterpreter(t)

	t.Run("Propagates error from left operand", func(t *testing.T) {
		leftErr := expression.NewUnary(
			*token.NewToken(token.MINUS, "-", nil, 1),
			expression.NewLiteral("invalid"),
		)
		expr := expression.NewBinary(
			leftErr,
			*token.NewToken(token.PLUS, "+", nil, 1),
			expression.NewLiteral(10.0),
		)
		_, err := in.Evaluate(expr)
		utils.AssertError(t, err)
	})

	t.Run("Propagates error from right operand", func(t *testing.T) {
		rightErr := expression.NewUnary(
			*token.NewToken(token.MINUS, "-", nil, 1),
			expression.NewLiteral("invalid"),
		)
		expr := expression.NewBinary(
			expression.NewLiteral(10.0),
			*token.NewToken(token.PLUS, "+", nil, 1),
			rightErr,
		)
		_, err := in.Evaluate(expr)
		utils.AssertError(t, err)
	})

	t.Run("Error when binary operator is unknown", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(10.0),
			*token.NewToken(token.EOF, "EOF", nil, 1),
			expression.NewLiteral(20.0),
		)
		_, err := in.Evaluate(expr)
		utils.AssertError(t, err)
	})
}

func TestCheckNumberOperands(t *testing.T) {
	op := *token.NewToken(token.PLUS, "+", nil, 1)

	t.Run("Both operands are numbers", func(t *testing.T) {
		left, right, err := checkNumberOperands(op, 10.0, 20.0)
		utils.AssertNoError(t, err)
		utils.AssertValue(t, left, 10.0)
		utils.AssertValue(t, right, 20.0)
	})

	t.Run("Left operand is not a number", func(t *testing.T) {
		_, _, err := checkNumberOperands(op, "not a number", 20.0)
		utils.AssertError(t, err)
	})

	t.Run("Right operand is not a number", func(t *testing.T) {
		_, _, err := checkNumberOperands(op, 10.0, "not a number")
		utils.AssertError(t, err)
	})

	t.Run("Neither operand is a number", func(t *testing.T) {
		_, _, err := checkNumberOperands(op, "a", "b")
		utils.AssertError(t, err)
	})
}

func TestEvaluateBinaryPlus(t *testing.T) {
	in := SetUpInterpreter(t)

	t.Run("Sum of two positive numbers", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(11.0),
			*token.NewToken(token.PLUS, "+", nil, 1),
			expression.NewLiteral(22.0),
		)
		val, err := in.Evaluate(expr)
		utils.AssertNoError(t, err)
		utils.AssertValue(t, val, 33.0)
	})

	t.Run("Sum of negative numbers", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(-10.0),
			*token.NewToken(token.PLUS, "+", nil, 1),
			expression.NewLiteral(-5.0),
		)
		val, err := in.Evaluate(expr)
		utils.AssertNoError(t, err)
		utils.AssertValue(t, val, -15.0)
	})

	t.Run("Sum of positive and negative number", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(10.0),
			*token.NewToken(token.PLUS, "+", nil, 1),
			expression.NewLiteral(-3.0),
		)
		val, err := in.Evaluate(expr)
		utils.AssertNoError(t, err)
		utils.AssertValue(t, val, 7.0)
	})

	t.Run("Sum of a negative number and a positive number", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(-7.0),
			*token.NewToken(token.PLUS, "+", nil, 1),
			expression.NewLiteral(2.0),
		)
		val, err := in.Evaluate(expr)
		utils.AssertNoError(t, err)
		utils.AssertValue(t, val, -5.0)
	})

	t.Run("Concatenation of two strings", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral("Hello "),
			*token.NewToken(token.PLUS, "+", nil, 1),
			expression.NewLiteral("world"),
		)
		val, err := in.Evaluate(expr)
		utils.AssertNoError(t, err)
		utils.AssertValue(t, val, "Hello world")
	})

	t.Run("Concatenation with empty string", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(""),
			*token.NewToken(token.PLUS, "+", nil, 1),
			expression.NewLiteral("abc"),
		)
		val, err := in.Evaluate(expr)
		utils.AssertNoError(t, err)
		utils.AssertValue(t, val, "abc")
	})

	t.Run("Error when adding string and number", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral("a"),
			*token.NewToken(token.PLUS, "+", nil, 1),
			expression.NewLiteral(1.0),
		)
		_, err := in.Evaluate(expr)
		utils.AssertError(t, err)
	})

	t.Run("Error when adding number and string", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(1.0),
			*token.NewToken(token.PLUS, "+", nil, 1),
			expression.NewLiteral("a"),
		)
		_, err := in.Evaluate(expr)
		utils.AssertError(t, err)
	})

	t.Run("Error when adding incompatible types", func(t *testing.T) {
		exprBool := expression.NewBinary(
			expression.NewLiteral(true),
			*token.NewToken(token.PLUS, "+", nil, 1),
			expression.NewLiteral(false),
		)
		_, err := in.Evaluate(exprBool)
		utils.AssertError(t, err)

		exprNil := expression.NewBinary(
			expression.NewLiteral(nil),
			*token.NewToken(token.PLUS, "+", nil, 1),
			expression.NewLiteral(nil),
		)
		_, err = in.Evaluate(exprNil)
		utils.AssertError(t, err)
	})
}

func TestEvaluateBinaryMinus(t *testing.T) {
	in := SetUpInterpreter(t)

	t.Run("Subtraction of two positive numbers", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(50.0),
			*token.NewToken(token.MINUS, "-", nil, 1),
			expression.NewLiteral(40.0),
		)
		val, err := in.Evaluate(expr)
		utils.AssertNoError(t, err)
		utils.AssertValue(t, val, 10.0)
	})

	t.Run("Subtraction of negative numbers", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(-1.0),
			*token.NewToken(token.MINUS, "-", nil, 1),
			expression.NewLiteral(-5.0),
		)
		val, err := in.Evaluate(expr)
		utils.AssertNoError(t, err)
		utils.AssertValue(t, val, 4.0)
	})

	t.Run("Subtraction with negative result", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(5.0),
			*token.NewToken(token.MINUS, "-", nil, 1),
			expression.NewLiteral(10.0),
		)
		val, err := in.Evaluate(expr)
		utils.AssertNoError(t, err)
		utils.AssertValue(t, val, -5.0)
	})

	t.Run("Error when left operand is not a number", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral("a"),
			*token.NewToken(token.MINUS, "-", nil, 1),
			expression.NewLiteral(10.0),
		)
		_, err := in.Evaluate(expr)
		utils.AssertError(t, err)
	})

	t.Run("Error when right operand is not a number", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(10.0),
			*token.NewToken(token.MINUS, "-", nil, 1),
			expression.NewLiteral("a"),
		)
		_, err := in.Evaluate(expr)
		utils.AssertError(t, err)
	})

	t.Run("Error when operands are booleans", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(true),
			*token.NewToken(token.MINUS, "-", nil, 1),
			expression.NewLiteral(false),
		)
		_, err := in.Evaluate(expr)
		utils.AssertError(t, err)
	})

	t.Run("Error when operands are nil", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(nil),
			*token.NewToken(token.MINUS, "-", nil, 1),
			expression.NewLiteral(nil),
		)
		_, err := in.Evaluate(expr)
		utils.AssertError(t, err)
	})
}

func TestEvaluateBinaryStar(t *testing.T) {
	in := SetUpInterpreter(t)

	t.Run("Multiplication of two positive numbers", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(5.0),
			*token.NewToken(token.STAR, "*", nil, 1),
			expression.NewLiteral(4.0),
		)
		val, err := in.Evaluate(expr)
		utils.AssertNoError(t, err)
		utils.AssertValue(t, val, 20.0)
	})

	t.Run("Multiplication of negative numbers", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(-2.0),
			*token.NewToken(token.STAR, "*", nil, 1),
			expression.NewLiteral(-5.0),
		)
		val, err := in.Evaluate(expr)
		utils.AssertNoError(t, err)
		utils.AssertValue(t, val, 10.0)
	})

	t.Run("Multiplication with negative result", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(5.0),
			*token.NewToken(token.STAR, "*", nil, 1),
			expression.NewLiteral(-10.0),
		)
		val, err := in.Evaluate(expr)
		utils.AssertNoError(t, err)
		utils.AssertValue(t, val, -50.0)
	})

	t.Run("Error when left operand is not a number", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral("a"),
			*token.NewToken(token.STAR, "*", nil, 1),
			expression.NewLiteral(10.0),
		)
		_, err := in.Evaluate(expr)
		utils.AssertError(t, err)
	})

	t.Run("Error when right operand is not a number", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(10.0),
			*token.NewToken(token.STAR, "*", nil, 1),
			expression.NewLiteral("a"),
		)
		_, err := in.Evaluate(expr)
		utils.AssertError(t, err)
	})

	t.Run("Error when operands are booleans", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(true),
			*token.NewToken(token.STAR, "*", nil, 1),
			expression.NewLiteral(false),
		)
		_, err := in.Evaluate(expr)
		utils.AssertError(t, err)
	})

	t.Run("Error when operands are nil", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(nil),
			*token.NewToken(token.STAR, "*", nil, 1),
			expression.NewLiteral(nil),
		)
		_, err := in.Evaluate(expr)
		utils.AssertError(t, err)
	})
}

func TestEvaluateBinarySlash(t *testing.T) {
	in := SetUpInterpreter(t)

	t.Run("Division of two positive numbers", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(20.0),
			*token.NewToken(token.SLASH, "/", nil, 1),
			expression.NewLiteral(4.0),
		)
		val, err := in.Evaluate(expr)
		utils.AssertNoError(t, err)
		utils.AssertValue(t, val, 5.0)
	})

	t.Run("Division of negative numbers", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(-10.0),
			*token.NewToken(token.SLASH, "/", nil, 1),
			expression.NewLiteral(-2.0),
		)
		val, err := in.Evaluate(expr)
		utils.AssertNoError(t, err)
		utils.AssertValue(t, val, 5.0)
	})

	t.Run("Division with negative result", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(10.0),
			*token.NewToken(token.SLASH, "/", nil, 1),
			expression.NewLiteral(-2.0),
		)
		val, err := in.Evaluate(expr)
		utils.AssertNoError(t, err)
		utils.AssertValue(t, val, -5.0)
	})

	t.Run("Division with fractional result", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(7.0),
			*token.NewToken(token.SLASH, "/", nil, 1),
			expression.NewLiteral(2.0),
		)
		val, err := in.Evaluate(expr)
		utils.AssertNoError(t, err)
		utils.AssertValue(t, val, 3.5)
	})

	t.Run("Error when left operand is not a number", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral("a"),
			*token.NewToken(token.SLASH, "/", nil, 1),
			expression.NewLiteral(10.0),
		)
		_, err := in.Evaluate(expr)
		utils.AssertError(t, err)
	})

	t.Run("Error when right operand is not a number", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(10.0),
			*token.NewToken(token.SLASH, "/", nil, 1),
			expression.NewLiteral("a"),
		)
		_, err := in.Evaluate(expr)
		utils.AssertError(t, err)
	})

	t.Run("A positive number divided by zero produces positive infinity", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(10.0),
			*token.NewToken(token.SLASH, "/", nil, 1),
			expression.NewLiteral(0.0),
		)
		val, err := in.Evaluate(expr)
		utils.AssertNoError(t, err)
		utils.AssertValue(t, val, math.Inf(1))
	})

	t.Run("A negative number divided by zero produces negative infinity", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(-10.0),
			*token.NewToken(token.SLASH, "/", nil, 1),
			expression.NewLiteral(0.0),
		)
		val, err := in.Evaluate(expr)
		utils.AssertNoError(t, err)
		utils.AssertValue(t, val, math.Inf(-1))
	})

	t.Run("Zero divided by zero produces NaN", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(0.0),
			*token.NewToken(token.SLASH, "/", nil, 1),
			expression.NewLiteral(0.0),
		)
		val, err := in.Evaluate(expr)
		utils.AssertNoError(t, err)

		num, ok := val.(float64)
		if !ok || !math.IsNaN(num) {
			t.Errorf("expected NaN, got %v", val)
		}
	})

	t.Run("Error when operands are booleans", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(true),
			*token.NewToken(token.SLASH, "/", nil, 1),
			expression.NewLiteral(false),
		)
		_, err := in.Evaluate(expr)
		utils.AssertError(t, err)
	})

	t.Run("Error when operands are nil", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(nil),
			*token.NewToken(token.SLASH, "/", nil, 1),
			expression.NewLiteral(nil),
		)
		_, err := in.Evaluate(expr)
		utils.AssertError(t, err)
	})
}

func TestEvaluateBinaryPercent(t *testing.T) {
	in := SetUpInterpreter(t)

	t.Run("Modulo of two positive numbers", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(10.0),
			*token.NewToken(token.PERCENT, "%", nil, 1),
			expression.NewLiteral(3.0),
		)
		val, err := in.Evaluate(expr)
		utils.AssertNoError(t, err)
		utils.AssertValue(t, val, 1.0)
	})

	t.Run("Modulo with negative dividend", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(-10.0),
			*token.NewToken(token.PERCENT, "%", nil, 1),
			expression.NewLiteral(3.0),
		)
		val, err := in.Evaluate(expr)
		utils.AssertNoError(t, err)
		utils.AssertValue(t, val, -1.0)
	})

	t.Run("Exact modulo", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(10.0),
			*token.NewToken(token.PERCENT, "%", nil, 1),
			expression.NewLiteral(5.0),
		)
		val, err := in.Evaluate(expr)
		utils.AssertNoError(t, err)
		utils.AssertValue(t, val, 0.0)
	})

	t.Run("Modulo with negative divisor", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(10.0),
			*token.NewToken(token.PERCENT, "%", nil, 1),
			expression.NewLiteral(-3.0),
		)
		val, err := in.Evaluate(expr)
		utils.AssertNoError(t, err)
		utils.AssertValue(t, val, 1.0)
	})

	t.Run("Modulo of two negative numbers", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(-10.0),
			*token.NewToken(token.PERCENT, "%", nil, 1),
			expression.NewLiteral(-3.0),
		)
		val, err := in.Evaluate(expr)
		utils.AssertNoError(t, err)
		utils.AssertValue(t, val, -1.0)
	})

	t.Run("Zero modulo positive number produces zero", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(0.0),
			*token.NewToken(token.PERCENT, "%", nil, 1),
			expression.NewLiteral(5.0),
		)
		val, err := in.Evaluate(expr)
		utils.AssertNoError(t, err)
		utils.AssertValue(t, val, 0.0)
	})

	t.Run("Modulo of floating point numbers", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(5.5),
			*token.NewToken(token.PERCENT, "%", nil, 1),
			expression.NewLiteral(2.0),
		)
		val, err := in.Evaluate(expr)
		utils.AssertNoError(t, err)
		utils.AssertValue(t, val, 1.5)
	})

	t.Run("Modulo by zero produces NaN", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(10.0),
			*token.NewToken(token.PERCENT, "%", nil, 1),
			expression.NewLiteral(0.0),
		)
		val, err := in.Evaluate(expr)
		utils.AssertNoError(t, err)

		num, ok := val.(float64)
		if !ok || !math.IsNaN(num) {
			t.Errorf("expected NaN, got %v", val)
		}
	})

	t.Run("Zero modulo zero produces NaN", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(0.0),
			*token.NewToken(token.PERCENT, "%", nil, 1),
			expression.NewLiteral(0.0),
		)
		val, err := in.Evaluate(expr)
		utils.AssertNoError(t, err)

		num, ok := val.(float64)
		if !ok || !math.IsNaN(num) {
			t.Errorf("expected NaN, got %v", val)
		}
	})

	t.Run("Error when left operand is not a number", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral("a"),
			*token.NewToken(token.PERCENT, "%", nil, 1),
			expression.NewLiteral(10.0),
		)
		_, err := in.Evaluate(expr)
		utils.AssertError(t, err)
	})

	t.Run("Error when right operand is not a number", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(10.0),
			*token.NewToken(token.PERCENT, "%", nil, 1),
			expression.NewLiteral("a"),
		)
		_, err := in.Evaluate(expr)
		utils.AssertError(t, err)
	})

	t.Run("Error when operands are booleans", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(true),
			*token.NewToken(token.PERCENT, "%", nil, 1),
			expression.NewLiteral(false),
		)
		_, err := in.Evaluate(expr)
		utils.AssertError(t, err)
	})

	t.Run("Error when operands are nil", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(nil),
			*token.NewToken(token.PERCENT, "%", nil, 1),
			expression.NewLiteral(nil),
		)
		_, err := in.Evaluate(expr)
		utils.AssertError(t, err)
	})
}

func TestEvaluateBinaryGreater(t *testing.T) {
	in := SetUpInterpreter(t)

	t.Run("Greater than with two positive numbers when true", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(10.0),
			*token.NewToken(token.GREATER, ">", nil, 1),
			expression.NewLiteral(5.0),
		)
		val, err := in.Evaluate(expr)
		utils.AssertNoError(t, err)
		utils.AssertValue(t, val, true)
	})

	t.Run("Greater than with two positive numbers when false", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(5.0),
			*token.NewToken(token.GREATER, ">", nil, 1),
			expression.NewLiteral(10.0),
		)
		val, err := in.Evaluate(expr)
		utils.AssertNoError(t, err)
		utils.AssertValue(t, val, false)
	})

	t.Run("Greater than with negative numbers when true", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(-5.0),
			*token.NewToken(token.GREATER, ">", nil, 1),
			expression.NewLiteral(-10.0),
		)
		val, err := in.Evaluate(expr)
		utils.AssertNoError(t, err)
		utils.AssertValue(t, val, true)
	})

	t.Run("Greater than with negative numbers when false", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(-10.0),
			*token.NewToken(token.GREATER, ">", nil, 1),
			expression.NewLiteral(-5.0),
		)
		val, err := in.Evaluate(expr)
		utils.AssertNoError(t, err)
		utils.AssertValue(t, val, false)
	})

	t.Run("Greater than with two equal numbers", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(10.0),
			*token.NewToken(token.GREATER, ">", nil, 1),
			expression.NewLiteral(10.0),
		)
		val, err := in.Evaluate(expr)
		utils.AssertNoError(t, err)
		utils.AssertValue(t, val, false)
	})

	t.Run("Error when left operand is not a number", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral("a"),
			*token.NewToken(token.GREATER, ">", nil, 1),
			expression.NewLiteral(10.0),
		)
		_, err := in.Evaluate(expr)
		utils.AssertError(t, err)
	})

	t.Run("Error when right operand is not a number", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(10.0),
			*token.NewToken(token.GREATER, ">", nil, 1),
			expression.NewLiteral("a"),
		)
		_, err := in.Evaluate(expr)
		utils.AssertError(t, err)
	})

	t.Run("Error when comparing strings", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral("a"),
			*token.NewToken(token.GREATER, ">", nil, 1),
			expression.NewLiteral("b"),
		)
		_, err := in.Evaluate(expr)
		utils.AssertError(t, err)
	})

	t.Run("Error when comparing booleans", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(true),
			*token.NewToken(token.GREATER, ">", nil, 1),
			expression.NewLiteral(false),
		)
		_, err := in.Evaluate(expr)
		utils.AssertError(t, err)
	})

	t.Run("Error when comparing nil values", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(nil),
			*token.NewToken(token.GREATER, ">", nil, 1),
			expression.NewLiteral(nil),
		)
		_, err := in.Evaluate(expr)
		utils.AssertError(t, err)
	})
}

func TestEvaluateBinaryGreaterEqual(t *testing.T) {
	in := SetUpInterpreter(t)

	t.Run("Greater equal than with two positive numbers when true", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(10.0),
			*token.NewToken(token.GREATER_EQUAL, ">=", nil, 1),
			expression.NewLiteral(5.0),
		)
		val, err := in.Evaluate(expr)
		utils.AssertNoError(t, err)
		utils.AssertValue(t, val, true)
	})

	t.Run("Greater equal than with two positive numbers when false", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(5.0),
			*token.NewToken(token.GREATER_EQUAL, ">=", nil, 1),
			expression.NewLiteral(10.0),
		)
		val, err := in.Evaluate(expr)
		utils.AssertNoError(t, err)
		utils.AssertValue(t, val, false)
	})

	t.Run("Greater equal than with two equal positive numbers", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(10.0),
			*token.NewToken(token.GREATER_EQUAL, ">=", nil, 1),
			expression.NewLiteral(10.0),
		)
		val, err := in.Evaluate(expr)
		utils.AssertNoError(t, err)
		utils.AssertValue(t, val, true)
	})

	t.Run("Greater equal than with negative numbers when true", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(-5.0),
			*token.NewToken(token.GREATER_EQUAL, ">=", nil, 1),
			expression.NewLiteral(-10.0),
		)
		val, err := in.Evaluate(expr)
		utils.AssertNoError(t, err)
		utils.AssertValue(t, val, true)
	})

	t.Run("Greater equal than with two equal negative numbers", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(-10.0),
			*token.NewToken(token.GREATER_EQUAL, ">=", nil, 1),
			expression.NewLiteral(-10.0),
		)
		val, err := in.Evaluate(expr)
		utils.AssertNoError(t, err)
		utils.AssertValue(t, val, true)
	})

	t.Run("Greater equal than with negative numbers when false", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(-10.0),
			*token.NewToken(token.GREATER_EQUAL, ">=", nil, 1),
			expression.NewLiteral(-5.0),
		)
		val, err := in.Evaluate(expr)
		utils.AssertNoError(t, err)
		utils.AssertValue(t, val, false)
	})

	t.Run("Error when left operand is not a number", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral("a"),
			*token.NewToken(token.GREATER_EQUAL, ">=", nil, 1),
			expression.NewLiteral(10.0),
		)
		_, err := in.Evaluate(expr)
		utils.AssertError(t, err)
	})

	t.Run("Error when right operand is not a number", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(10.0),
			*token.NewToken(token.GREATER_EQUAL, ">=", nil, 1),
			expression.NewLiteral("a"),
		)
		_, err := in.Evaluate(expr)
		utils.AssertError(t, err)
	})

	t.Run("Error when comparing strings", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral("a"),
			*token.NewToken(token.GREATER_EQUAL, ">=", nil, 1),
			expression.NewLiteral("b"),
		)
		_, err := in.Evaluate(expr)
		utils.AssertError(t, err)
	})

	t.Run("Error when comparing booleans", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(true),
			*token.NewToken(token.GREATER_EQUAL, ">=", nil, 1),
			expression.NewLiteral(false),
		)
		_, err := in.Evaluate(expr)
		utils.AssertError(t, err)
	})

	t.Run("Error when comparing nil values", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(nil),
			*token.NewToken(token.GREATER_EQUAL, ">=", nil, 1),
			expression.NewLiteral(nil),
		)
		_, err := in.Evaluate(expr)
		utils.AssertError(t, err)
	})
}

func TestEvaluateBinaryLess(t *testing.T) {
	in := SetUpInterpreter(t)

	t.Run("Less than with two positive numbers when true", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(10.0),
			*token.NewToken(token.LESS, "<", nil, 1),
			expression.NewLiteral(50.0),
		)
		val, err := in.Evaluate(expr)
		utils.AssertNoError(t, err)
		utils.AssertValue(t, val, true)
	})

	t.Run("Less than with two positive numbers when false", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(50.0),
			*token.NewToken(token.LESS, "<", nil, 1),
			expression.NewLiteral(10.0),
		)
		val, err := in.Evaluate(expr)
		utils.AssertNoError(t, err)
		utils.AssertValue(t, val, false)
	})

	t.Run("Less than with negative numbers when true", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(-50.0),
			*token.NewToken(token.LESS, "<", nil, 1),
			expression.NewLiteral(-10.0),
		)
		val, err := in.Evaluate(expr)
		utils.AssertNoError(t, err)
		utils.AssertValue(t, val, true)
	})

	t.Run("Less than with negative numbers when false", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(-10.0),
			*token.NewToken(token.LESS, "<", nil, 1),
			expression.NewLiteral(-50.0),
		)
		val, err := in.Evaluate(expr)
		utils.AssertNoError(t, err)
		utils.AssertValue(t, val, false)
	})

	t.Run("Less than with two equal numbers", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(10.0),
			*token.NewToken(token.LESS, "<", nil, 1),
			expression.NewLiteral(10.0),
		)
		val, err := in.Evaluate(expr)
		utils.AssertNoError(t, err)
		utils.AssertValue(t, val, false)
	})

	t.Run("Error when left operand is not a number", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral("a"),
			*token.NewToken(token.LESS, "<", nil, 1),
			expression.NewLiteral(10.0),
		)
		_, err := in.Evaluate(expr)
		utils.AssertError(t, err)
	})

	t.Run("Error when right operand is not a number", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(10.0),
			*token.NewToken(token.LESS, "<", nil, 1),
			expression.NewLiteral("a"),
		)
		_, err := in.Evaluate(expr)
		utils.AssertError(t, err)
	})

	t.Run("Error when comparing strings", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral("a"),
			*token.NewToken(token.LESS, "<", nil, 1),
			expression.NewLiteral("b"),
		)
		_, err := in.Evaluate(expr)
		utils.AssertError(t, err)
	})

	t.Run("Error when comparing booleans", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(true),
			*token.NewToken(token.LESS, "<", nil, 1),
			expression.NewLiteral(false),
		)
		_, err := in.Evaluate(expr)
		utils.AssertError(t, err)
	})

	t.Run("Error when comparing nil values", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(nil),
			*token.NewToken(token.LESS, "<", nil, 1),
			expression.NewLiteral(nil),
		)
		_, err := in.Evaluate(expr)
		utils.AssertError(t, err)
	})
}

func TestEvaluateBinaryLessEqual(t *testing.T) {
	in := SetUpInterpreter(t)

	t.Run("Less equal than with two positive numbers when true", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(10.0),
			*token.NewToken(token.LESS_EQUAL, "<=", nil, 1),
			expression.NewLiteral(50.0),
		)
		val, err := in.Evaluate(expr)
		utils.AssertNoError(t, err)
		utils.AssertValue(t, val, true)
	})

	t.Run("Less equal than with two positive numbers when false", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(50.0),
			*token.NewToken(token.LESS_EQUAL, "<=", nil, 1),
			expression.NewLiteral(10.0),
		)
		val, err := in.Evaluate(expr)
		utils.AssertNoError(t, err)
		utils.AssertValue(t, val, false)
	})

	t.Run("Less equal than with two equal positive numbers", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(10.0),
			*token.NewToken(token.LESS_EQUAL, "<=", nil, 1),
			expression.NewLiteral(10.0),
		)
		val, err := in.Evaluate(expr)
		utils.AssertNoError(t, err)
		utils.AssertValue(t, val, true)
	})

	t.Run("Less equal than with negative numbers when true", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(-50.0),
			*token.NewToken(token.LESS_EQUAL, "<=", nil, 1),
			expression.NewLiteral(-10.0),
		)
		val, err := in.Evaluate(expr)
		utils.AssertNoError(t, err)
		utils.AssertValue(t, val, true)
	})

	t.Run("Less equal than with two equal negative numbers", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(-10.0),
			*token.NewToken(token.LESS_EQUAL, "<=", nil, 1),
			expression.NewLiteral(-10.0),
		)
		val, err := in.Evaluate(expr)
		utils.AssertNoError(t, err)
		utils.AssertValue(t, val, true)
	})

	t.Run("Less equal than with negative numbers when false", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(-10.0),
			*token.NewToken(token.LESS_EQUAL, "<=", nil, 1),
			expression.NewLiteral(-50.0),
		)
		val, err := in.Evaluate(expr)
		utils.AssertNoError(t, err)
		utils.AssertValue(t, val, false)
	})

	t.Run("Error when left operand is not a number", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral("a"),
			*token.NewToken(token.LESS_EQUAL, "<=", nil, 1),
			expression.NewLiteral(10.0),
		)
		_, err := in.Evaluate(expr)
		utils.AssertError(t, err)
	})

	t.Run("Error when right operand is not a number", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(10.0),
			*token.NewToken(token.LESS_EQUAL, "<=", nil, 1),
			expression.NewLiteral("a"),
		)
		_, err := in.Evaluate(expr)
		utils.AssertError(t, err)
	})

	t.Run("Error when comparing strings", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral("a"),
			*token.NewToken(token.LESS_EQUAL, "<=", nil, 1),
			expression.NewLiteral("b"),
		)
		_, err := in.Evaluate(expr)
		utils.AssertError(t, err)
	})

	t.Run("Error when comparing booleans", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(true),
			*token.NewToken(token.LESS_EQUAL, "<=", nil, 1),
			expression.NewLiteral(false),
		)
		_, err := in.Evaluate(expr)
		utils.AssertError(t, err)
	})

	t.Run("Error when comparing nil values", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(nil),
			*token.NewToken(token.LESS_EQUAL, "<=", nil, 1),
			expression.NewLiteral(nil),
		)
		_, err := in.Evaluate(expr)
		utils.AssertError(t, err)
	})
}

func TestEvaluateBinaryEqualEqual(t *testing.T) {
	in := SetUpInterpreter(t)

	t.Run("Equal numbers when true", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(42.0),
			*token.NewToken(token.EQUAL_EQUAL, "==", nil, 1),
			expression.NewLiteral(42.0),
		)
		val, err := in.Evaluate(expr)
		utils.AssertNoError(t, err)
		utils.AssertValue(t, val, true)
	})

	t.Run("Equal numbers when false", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(42.0),
			*token.NewToken(token.EQUAL_EQUAL, "==", nil, 1),
			expression.NewLiteral(10.0),
		)
		val, err := in.Evaluate(expr)
		utils.AssertNoError(t, err)
		utils.AssertValue(t, val, false)
	})

	t.Run("Equal negative numbers", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(-10.0),
			*token.NewToken(token.EQUAL_EQUAL, "==", nil, 1),
			expression.NewLiteral(-10.0),
		)
		val, err := in.Evaluate(expr)
		utils.AssertNoError(t, err)
		utils.AssertValue(t, val, true)
	})

	t.Run("Equal strings when true", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral("hello"),
			*token.NewToken(token.EQUAL_EQUAL, "==", nil, 1),
			expression.NewLiteral("hello"),
		)
		val, err := in.Evaluate(expr)
		utils.AssertNoError(t, err)
		utils.AssertValue(t, val, true)
	})

	t.Run("Equal strings when false", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral("hello"),
			*token.NewToken(token.EQUAL_EQUAL, "==", nil, 1),
			expression.NewLiteral("world"),
		)
		val, err := in.Evaluate(expr)
		utils.AssertNoError(t, err)
		utils.AssertValue(t, val, false)
	})

	t.Run("Equal booleans when true", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(true),
			*token.NewToken(token.EQUAL_EQUAL, "==", nil, 1),
			expression.NewLiteral(true),
		)
		val, err := in.Evaluate(expr)
		utils.AssertNoError(t, err)
		utils.AssertValue(t, val, true)
	})

	t.Run("Equal booleans when false", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(true),
			*token.NewToken(token.EQUAL_EQUAL, "==", nil, 1),
			expression.NewLiteral(false),
		)
		val, err := in.Evaluate(expr)
		utils.AssertNoError(t, err)
		utils.AssertValue(t, val, false)
	})

	t.Run("Equal nil with nil", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(nil),
			*token.NewToken(token.EQUAL_EQUAL, "==", nil, 1),
			expression.NewLiteral(nil),
		)
		val, err := in.Evaluate(expr)
		utils.AssertNoError(t, err)
		utils.AssertValue(t, val, true)
	})

	t.Run("Equal nil with non-nil produces false", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(nil),
			*token.NewToken(token.EQUAL_EQUAL, "==", nil, 1),
			expression.NewLiteral(0.0),
		)
		val, err := in.Evaluate(expr)
		utils.AssertNoError(t, err)
		utils.AssertValue(t, val, false)
	})

	t.Run("Equal with different types produces false without error", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(42.0),
			*token.NewToken(token.EQUAL_EQUAL, "==", nil, 1),
			expression.NewLiteral("42.0"),
		)
		val, err := in.Evaluate(expr)
		utils.AssertNoError(t, err)
		utils.AssertValue(t, val, false)
	})

	t.Run("Equal boolean and number produces false without error", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(true),
			*token.NewToken(token.EQUAL_EQUAL, "==", nil, 1),
			expression.NewLiteral(1.0),
		)
		val, err := in.Evaluate(expr)
		utils.AssertNoError(t, err)
		utils.AssertValue(t, val, false)
	})
}

func TestEvaluateBinaryBangEqual(t *testing.T) {
	in := SetUpInterpreter(t)

	t.Run("Bang equal numbers when false", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(42.0),
			*token.NewToken(token.BANG_EQUAL, "!=", nil, 1),
			expression.NewLiteral(42.0),
		)
		val, err := in.Evaluate(expr)
		utils.AssertNoError(t, err)
		utils.AssertValue(t, val, false)
	})

	t.Run("Bang equal numbers when true", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(42.0),
			*token.NewToken(token.BANG_EQUAL, "!=", nil, 1),
			expression.NewLiteral(10.0),
		)
		val, err := in.Evaluate(expr)
		utils.AssertNoError(t, err)
		utils.AssertValue(t, val, true)
	})

	t.Run("Bang equal negative numbers", func(t *testing.T) {
		exprEqual := expression.NewBinary(
			expression.NewLiteral(-10.0),
			*token.NewToken(token.BANG_EQUAL, "!=", nil, 1),
			expression.NewLiteral(-10.0),
		)
		val, err := in.Evaluate(exprEqual)
		utils.AssertNoError(t, err)
		utils.AssertValue(t, val, false)

		exprDiff := expression.NewBinary(
			expression.NewLiteral(-10.0),
			*token.NewToken(token.BANG_EQUAL, "!=", nil, 1),
			expression.NewLiteral(-5.0),
		)
		val, err = in.Evaluate(exprDiff)
		utils.AssertNoError(t, err)
		utils.AssertValue(t, val, true)
	})

	t.Run("Bang equal strings when false", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral("hello"),
			*token.NewToken(token.BANG_EQUAL, "!=", nil, 1),
			expression.NewLiteral("hello"),
		)
		val, err := in.Evaluate(expr)
		utils.AssertNoError(t, err)
		utils.AssertValue(t, val, false)
	})

	t.Run("Bang equal strings when true", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral("hello"),
			*token.NewToken(token.BANG_EQUAL, "!=", nil, 1),
			expression.NewLiteral("world"),
		)
		val, err := in.Evaluate(expr)
		utils.AssertNoError(t, err)
		utils.AssertValue(t, val, true)
	})

	t.Run("Bang equal booleans when false", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(true),
			*token.NewToken(token.BANG_EQUAL, "!=", nil, 1),
			expression.NewLiteral(true),
		)
		val, err := in.Evaluate(expr)
		utils.AssertNoError(t, err)
		utils.AssertValue(t, val, false)
	})

	t.Run("Bang equal booleans when true", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(true),
			*token.NewToken(token.BANG_EQUAL, "!=", nil, 1),
			expression.NewLiteral(false),
		)
		val, err := in.Evaluate(expr)
		utils.AssertNoError(t, err)
		utils.AssertValue(t, val, true)
	})

	t.Run("Bang equal nil with nil produces false", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(nil),
			*token.NewToken(token.BANG_EQUAL, "!=", nil, 1),
			expression.NewLiteral(nil),
		)
		val, err := in.Evaluate(expr)
		utils.AssertNoError(t, err)
		utils.AssertValue(t, val, false)
	})

	t.Run("Bang equal nil with non-nil produces true", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(nil),
			*token.NewToken(token.BANG_EQUAL, "!=", nil, 1),
			expression.NewLiteral(0.0),
		)
		val, err := in.Evaluate(expr)
		utils.AssertNoError(t, err)
		utils.AssertValue(t, val, true)
	})

	t.Run("Bang equal with different types produces true without error", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(42.0),
			*token.NewToken(token.BANG_EQUAL, "!=", nil, 1),
			expression.NewLiteral("42.0"),
		)
		val, err := in.Evaluate(expr)
		utils.AssertNoError(t, err)
		utils.AssertValue(t, val, true)
	})

	t.Run("Bang equal boolean and number produces true without error", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(true),
			*token.NewToken(token.BANG_EQUAL, "!=", nil, 1),
			expression.NewLiteral(1.0),
		)
		val, err := in.Evaluate(expr)
		utils.AssertNoError(t, err)
		utils.AssertValue(t, val, true)
	})
}

func TestEvaluateVariable(t *testing.T) {
	in := SetUpInterpreter(t)

	t.Run("Evaluate a declared variable", func(t *testing.T) {
		tok := *token.NewToken(token.IDENTIFIER, "x", nil, 1)
		utils.AssertNoError(t, in.Execute(statements.NewVarStmt(tok, expression.NewLiteral(42.0))))

		val, err := in.Evaluate(expression.NewVariable(tok))
		utils.AssertNoError(t, err)
		utils.AssertValue(t, val, 42.0)
	})

	t.Run("Error when evaluating undeclared variable", func(t *testing.T) {
		tok := *token.NewToken(token.IDENTIFIER, "undeclared", nil, 1)
		_, err := in.Evaluate(expression.NewVariable(tok))
		utils.AssertErrorMessage(t, err, "undefined variable 'undeclared'")
	})

	t.Run("Resolved variable reads from the environment at its distance", func(t *testing.T) {
		in := SetUpInterpreter(t)
		tok := *token.NewToken(token.IDENTIFIER, "a", nil, 1)
		in.globals.Define("a", "outer")
		in.environment = environment.NewEnclosedEnvironment(in.globals)
		in.environment.Define("a", "inner")

		local := expression.NewVariable(tok)
		enclosing := expression.NewVariable(tok)
		in.Resolve(local, 0)
		in.Resolve(enclosing, 1)

		val, err := in.Evaluate(local)
		utils.AssertNoError(t, err)
		utils.AssertValue(t, val, "inner")

		val, err = in.Evaluate(enclosing)
		utils.AssertNoError(t, err)
		utils.AssertValue(t, val, "outer")
	})

	t.Run("Unresolved variable reads from globals ignoring a local with the same name", func(t *testing.T) {
		in := SetUpInterpreter(t)
		tok := *token.NewToken(token.IDENTIFIER, "a", nil, 1)
		in.globals.Define("a", "global")
		in.environment = environment.NewEnclosedEnvironment(in.globals)
		in.environment.Define("a", "block")

		val, err := in.Evaluate(expression.NewVariable(tok))
		utils.AssertNoError(t, err)
		utils.AssertValue(t, val, "global")
	})
}

func TestEvaluateAssign(t *testing.T) {
	in := SetUpInterpreter(t)

	t.Run("Assign a number to existing variable", func(t *testing.T) {
		tok := *token.NewToken(token.IDENTIFIER, "x", nil, 1)
		utils.AssertNoError(t, in.Execute(statements.NewVarStmt(tok, expression.NewLiteral(10.0))))

		assignExpr := expression.NewAssign(tok, expression.NewLiteral(20.0))
		val, err := in.Evaluate(assignExpr)
		utils.AssertNoError(t, err)
		utils.AssertValue(t, val, 20.0)

		varVal, err := in.Evaluate(expression.NewVariable(tok))
		utils.AssertNoError(t, err)
		utils.AssertValue(t, varVal, 20.0)
	})

	t.Run("Nested assignment propagates inner value and updates both variables", func(t *testing.T) {
		tokA := *token.NewToken(token.IDENTIFIER, "a", nil, 1)
		tokB := *token.NewToken(token.IDENTIFIER, "b", nil, 1)
		utils.AssertNoError(t, in.Execute(statements.NewVarStmt(tokA, expression.NewLiteral(1.0))))
		utils.AssertNoError(t, in.Execute(statements.NewVarStmt(tokB, expression.NewLiteral(2.0))))

		nested := expression.NewAssign(tokA, expression.NewAssign(tokB, expression.NewLiteral(5.0)))
		val, err := in.Evaluate(nested)
		utils.AssertNoError(t, err)
		utils.AssertValue(t, val, 5.0)

		varValA, err := in.Evaluate(expression.NewVariable(tokA))
		utils.AssertNoError(t, err)
		utils.AssertValue(t, varValA, 5.0)

		varValB, err := in.Evaluate(expression.NewVariable(tokB))
		utils.AssertNoError(t, err)
		utils.AssertValue(t, varValB, 5.0)
	})

	t.Run("Assign to undeclared produces runtime error", func(t *testing.T) {
		tok := *token.NewToken(token.IDENTIFIER, "undeclared", nil, 1)
		assignExpr := expression.NewAssign(tok, expression.NewLiteral(30.0))
		_, err := in.Evaluate(assignExpr)
		utils.AssertErrorMessage(t, err, "undefined variable 'undeclared'")
	})

	t.Run("Error evaluating right-hand expression preserves previous variable value", func(t *testing.T) {
		tok := *token.NewToken(token.IDENTIFIER, "safe", nil, 1)
		utils.AssertNoError(t, in.Execute(statements.NewVarStmt(tok, expression.NewLiteral(10.0))))

		assignExpr := expression.NewAssign(tok, invalidExpr())
		_, err := in.Evaluate(assignExpr)
		utils.AssertError(t, err)

		varVal, err := in.Evaluate(expression.NewVariable(tok))
		utils.AssertNoError(t, err)
		utils.AssertValue(t, varVal, 10.0)
	})

	t.Run("Resolved assignment modifies the environment at its distance", func(t *testing.T) {
		in := SetUpInterpreter(t)
		tok := *token.NewToken(token.IDENTIFIER, "a", nil, 1)
		outer := in.globals
		outer.Define("a", "outer")
		inner := environment.NewEnclosedEnvironment(outer)
		inner.Define("a", "inner")
		in.environment = inner

		assign := expression.NewAssign(tok, expression.NewLiteral("changed"))
		in.Resolve(assign, 1)

		_, err := in.Evaluate(assign)
		utils.AssertNoError(t, err)

		outerVal, err := outer.Get(tok)
		utils.AssertNoError(t, err)
		utils.AssertValue(t, outerVal, "changed")

		innerVal, err := inner.Get(tok)
		utils.AssertNoError(t, err)
		utils.AssertValue(t, innerVal, "inner")
	})

	t.Run("Unresolved assignment modifies globals ignoring a local with the same name", func(t *testing.T) {
		in := SetUpInterpreter(t)
		tok := *token.NewToken(token.IDENTIFIER, "a", nil, 1)
		in.globals.Define("a", "global")
		block := environment.NewEnclosedEnvironment(in.globals)
		block.Define("a", "block")
		in.environment = block

		_, err := in.Evaluate(expression.NewAssign(tok, expression.NewLiteral("changed")))
		utils.AssertNoError(t, err)

		globalVal, err := in.globals.Get(tok)
		utils.AssertNoError(t, err)
		utils.AssertValue(t, globalVal, "changed")

		blockVal, err := block.Get(tok)
		utils.AssertNoError(t, err)
		utils.AssertValue(t, blockVal, "block")
	})
}

func TestEvaluateLogical(t *testing.T) {
	in := SetUpInterpreter(t)

	t.Run("Or with truthy left short-circuits and does not evaluate right", func(t *testing.T) {
		expr := expression.NewLogical(
			expression.NewLiteral(true),
			*token.NewToken(token.OR, "or", nil, 1),
			invalidExpr(),
		)
		val, err := in.Evaluate(expr)
		utils.AssertNoError(t, err)
		utils.AssertValue(t, val, true)
	})

	t.Run("Or with falsy left evaluates and returns right", func(t *testing.T) {
		expr := expression.NewLogical(
			expression.NewLiteral(false),
			*token.NewToken(token.OR, "or", nil, 1),
			expression.NewLiteral("hello"),
		)
		val, err := in.Evaluate(expr)
		utils.AssertNoError(t, err)
		utils.AssertValue(t, val, "hello")
	})

	t.Run("Or returns the truthy operand as-is, not coerced to boolean", func(t *testing.T) {
		expr := expression.NewLogical(
			expression.NewLiteral(0.0),
			*token.NewToken(token.OR, "or", nil, 1),
			expression.NewLiteral("ignored"),
		)
		val, err := in.Evaluate(expr)
		utils.AssertNoError(t, err)
		utils.AssertValue(t, val, 0.0)
	})

	t.Run("Or propagates error from left operand", func(t *testing.T) {
		expr := expression.NewLogical(
			invalidExpr(),
			*token.NewToken(token.OR, "or", nil, 1),
			expression.NewLiteral(true),
		)
		_, err := in.Evaluate(expr)
		utils.AssertError(t, err)
	})

	t.Run("Or propagates error from right operand when left is falsy", func(t *testing.T) {
		expr := expression.NewLogical(
			expression.NewLiteral(false),
			*token.NewToken(token.OR, "or", nil, 1),
			invalidExpr(),
		)
		_, err := in.Evaluate(expr)
		utils.AssertError(t, err)
	})

	t.Run("And with falsy left short-circuits and does not evaluate right", func(t *testing.T) {
		expr := expression.NewLogical(
			expression.NewLiteral(false),
			*token.NewToken(token.AND, "and", nil, 1),
			invalidExpr(),
		)
		val, err := in.Evaluate(expr)
		utils.AssertNoError(t, err)
		utils.AssertValue(t, val, false)
	})

	t.Run("And with truthy left evaluates and returns right", func(t *testing.T) {
		expr := expression.NewLogical(
			expression.NewLiteral(true),
			*token.NewToken(token.AND, "and", nil, 1),
			expression.NewLiteral("hello"),
		)
		val, err := in.Evaluate(expr)
		utils.AssertNoError(t, err)
		utils.AssertValue(t, val, "hello")
	})

	t.Run("And propagates error from left operand", func(t *testing.T) {
		expr := expression.NewLogical(
			invalidExpr(),
			*token.NewToken(token.AND, "and", nil, 1),
			expression.NewLiteral(true),
		)
		_, err := in.Evaluate(expr)
		utils.AssertError(t, err)
	})

	t.Run("And propagates error from right operand when left is truthy", func(t *testing.T) {
		expr := expression.NewLogical(
			expression.NewLiteral(true),
			*token.NewToken(token.AND, "and", nil, 1),
			invalidExpr(),
		)
		_, err := in.Evaluate(expr)
		utils.AssertError(t, err)
	})
}

type testCallable struct {
	arity int
	call  func(in *Interpreter, arguments []any) (any, error)
}

func (c *testCallable) Arity() int {
	return c.arity
}

func (c *testCallable) Call(in *Interpreter, arguments []any) (any, error) {
	return c.call(in, arguments)
}

func (c *testCallable) String() string {
	return "<fn test>"
}

func TestEvaluateCall(t *testing.T) {
	in := SetUpInterpreter(t)
	paren := *token.NewToken(token.RIGHT_PAREN, ")", nil, 1)

	t.Run("Evaluate call expression returns result of invocation", func(t *testing.T) {
		in.environment.Define("add", &testCallable{
			arity: 2,
			call: func(in *Interpreter, arguments []any) (any, error) {
				a := arguments[0].(float64)
				b := arguments[1].(float64)
				return a + b, nil
			},
		})

		call := expression.NewCall(
			expression.NewVariable(*token.NewToken(token.IDENTIFIER, "add", nil, 1)),
			paren,
			[]expression.Expr{
				expression.NewLiteral(2.0),
				expression.NewLiteral(3.0),
			},
		)
		val, err := in.Evaluate(call)
		utils.AssertNoError(t, err)
		utils.AssertValue(t, val, 5.0)
	})

	t.Run("Evaluate call evaluates arguments in left to right order", func(t *testing.T) {
		in.environment.Define("a", 1.0)
		in.environment.Define("b", 2.0)
		in.environment.Define("pair", &testCallable{
			arity: 2,
			call: func(in *Interpreter, arguments []any) (any, error) {
				return arguments, nil
			},
		})

		call := expression.NewCall(
			expression.NewVariable(*token.NewToken(token.IDENTIFIER, "pair", nil, 1)),
			paren,
			[]expression.Expr{
				expression.NewAssign(*token.NewToken(token.IDENTIFIER, "a", nil, 1), expression.NewLiteral(10.0)),
				expression.NewAssign(*token.NewToken(token.IDENTIFIER, "b", nil, 1), expression.NewBinary(
					expression.NewVariable(*token.NewToken(token.IDENTIFIER, "a", nil, 1)),
					*token.NewToken(token.PLUS, "+", nil, 1),
					expression.NewLiteral(5.0),
				)),
			},
		)

		val, err := in.Evaluate(call)
		utils.AssertNoError(t, err)

		args := val.([]any)
		utils.AssertValue(t, args[0], 10.0)
		utils.AssertValue(t, args[1], 15.0)

		valB, _ := in.environment.Get(*token.NewToken(token.IDENTIFIER, "b", nil, 1))
		utils.AssertValue(t, valB, 15.0)
	})

	t.Run("Error when callee is not callable", func(t *testing.T) {
		nonCallables := []any{123.0, "not a func", true, false, nil}
		for _, target := range nonCallables {
			call := expression.NewCall(expression.NewLiteral(target), paren, nil)
			_, err := in.Evaluate(call)
			utils.AssertErrorMessage(t, err, "can only call functions and classes")
		}
	})

	t.Run("Error when argument count does not match arity", func(t *testing.T) {
		in.environment.Define("binaryFn", &testCallable{
			arity: 2,
			call: func(in *Interpreter, arguments []any) (any, error) {
				return nil, nil
			},
		})

		callee := expression.NewVariable(*token.NewToken(token.IDENTIFIER, "binaryFn", nil, 1))

		callTooFew := expression.NewCall(callee, paren, []expression.Expr{expression.NewLiteral(1.0)})
		_, errTooFew := in.Evaluate(callTooFew)
		utils.AssertErrorMessage(t, errTooFew, "expected 2 arguments but got 1")

		callTooMany := expression.NewCall(callee, paren, []expression.Expr{
			expression.NewLiteral(1.0),
			expression.NewLiteral(2.0),
			expression.NewLiteral(3.0),
		})
		_, errTooMany := in.Evaluate(callTooMany)
		utils.AssertErrorMessage(t, errTooMany, "expected 2 arguments but got 3")
	})

	t.Run("Error when callee evaluation fails", func(t *testing.T) {
		call := expression.NewCall(invalidExpr(), paren, nil)
		_, err := in.Evaluate(call)
		utils.AssertError(t, err)
	})

	t.Run("Error when argument evaluation fails", func(t *testing.T) {
		in.environment.Define("unaryFn", &testCallable{
			arity: 1,
			call: func(in *Interpreter, arguments []any) (any, error) {
				return nil, nil
			},
		})

		call := expression.NewCall(
			expression.NewVariable(*token.NewToken(token.IDENTIFIER, "unaryFn", nil, 1)),
			paren,
			[]expression.Expr{invalidExpr()},
		)
		_, err := in.Evaluate(call)
		utils.AssertError(t, err)
	})

	t.Run("Error when invocation returns error", func(t *testing.T) {
		in.environment.Define("failingFn", &testCallable{
			arity: 0,
			call: func(in *Interpreter, arguments []any) (any, error) {
				return nil, errors.NewLoxError(1, "invocation failure")
			},
		})

		call := expression.NewCall(
			expression.NewVariable(*token.NewToken(token.IDENTIFIER, "failingFn", nil, 1)),
			paren,
			nil,
		)
		_, err := in.Evaluate(call)
		utils.AssertErrorMessage(t, err, "invocation failure")
	})
}

func TestNativeClock(t *testing.T) {
	in := SetUpInterpreter(t)
	paren := *token.NewToken(token.RIGHT_PAREN, ")", nil, 1)

	t.Run("Evaluate native clock returns positive timestamp", func(t *testing.T) {
		call := expression.NewCall(
			expression.NewVariable(*token.NewToken(token.IDENTIFIER, "clock", nil, 1)),
			paren,
			nil,
		)
		val, err := in.Evaluate(call)
		utils.AssertNoError(t, err)

		num, ok := val.(float64)
		if !ok || num <= 0 {
			t.Fatalf("expected positive float64, got %v", val)
		}
	})

	t.Run("Error when calling native clock with arguments", func(t *testing.T) {
		call := expression.NewCall(
			expression.NewVariable(*token.NewToken(token.IDENTIFIER, "clock", nil, 1)),
			paren,
			[]expression.Expr{expression.NewLiteral(1.0)},
		)
		_, err := in.Evaluate(call)
		utils.AssertErrorMessage(t, err, "expected 0 arguments but got 1")
	})
}
