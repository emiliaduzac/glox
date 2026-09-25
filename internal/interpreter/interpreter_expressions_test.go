package interpreter

import (
	"math"
	"testing"
	"tp-lox/internal/expression"
	"tp-lox/internal/token"
)

func TestEvaluateUnsupportedExpression(t *testing.T) {
	in := SetUpInterpreter(t)

	t.Run("Error when expression is nil", func(t *testing.T) {
		_, err := in.Evaluate(nil)
		AssertError(t, err)
	})

	t.Run("Error when expression type is unknown", func(t *testing.T) {
		_, err := in.Evaluate(UnsupportedExpr{})
		AssertError(t, err)
	})
}

func TestEvaluateLiterals(t *testing.T) {
	in := SetUpInterpreter(t)

	t.Run("Evaluate positive integer literal", func(t *testing.T) {
		expression := expression.NewLiteral(99.0)
		value, err := in.Evaluate(expression)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		AssertValue(t, value, 99.0)
	})

	t.Run("Evaluate empty string literal", func(t *testing.T) {
		expression := expression.NewLiteral("")
		value, err := in.Evaluate(expression)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		AssertValue(t, value, "")
	})

	t.Run("Evaluate string literal", func(t *testing.T) {
		expression := expression.NewLiteral("Hello World")
		value, err := in.Evaluate(expression)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		AssertValue(t, value, "Hello World")
	})

	t.Run("Evaluate boolean literal", func(t *testing.T) {
		expression := expression.NewLiteral(true)
		value, err := in.Evaluate(expression)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		AssertValue(t, value, true)
	})

	t.Run("Evaluate false boolean literal", func(t *testing.T) {
		expression := expression.NewLiteral(false)
		value, err := in.Evaluate(expression)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		AssertValue(t, value, false)
	})

	t.Run("Evaluate nil literal", func(t *testing.T) {
		expression := expression.NewLiteral(nil)
		value, err := in.Evaluate(expression)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		AssertValue(t, value, nil)
	})
}

func TestStringify(t *testing.T) {
	t.Run("Stringify nil value", func(t *testing.T) {
		AssertValue(t, Stringify(nil), "nil")
	})

	t.Run("Stringify integer value", func(t *testing.T) {
		AssertValue(t, Stringify(99), "99")
	})

	t.Run("Stringify string value", func(t *testing.T) {
		AssertValue(t, Stringify("Hello"), "Hello")
	})

	t.Run("Stringify boolean true value", func(t *testing.T) {
		AssertValue(t, Stringify(true), "true")
	})

	t.Run("Stringify boolean false value", func(t *testing.T) {
		AssertValue(t, Stringify(false), "false")
	})
}

func TestEvaluateGrouping(t *testing.T) {
	in := SetUpInterpreter(t)

	t.Run("Evaluate grouping with literal", func(t *testing.T) {
		expression := expression.NewGrouping(expression.NewLiteral(99))
		value, err := in.Evaluate(expression)
		AssertNoError(t, err)
		AssertValue(t, value, 99)
	})

	t.Run("Evaluate gruped string literal", func(t *testing.T) {
		expression := expression.NewGrouping(expression.NewLiteral("Hello"))
		value, err := in.Evaluate(expression)
		AssertNoError(t, err)
		AssertValue(t, value, "Hello")
	})

	t.Run("Evaluate nested grouping with literal", func(t *testing.T) {
		expression := expression.NewGrouping(expression.NewGrouping(expression.NewGrouping(expression.NewLiteral(88))))
		value, err := in.Evaluate(expression)
		AssertNoError(t, err)
		AssertValue(t, value, 88)
	})
}

func TestIsTruthy(t *testing.T) {
	t.Run("Nil is falsy", func(t *testing.T) {
		AssertValue(t, isTruthy(nil), false)
	})

	t.Run("False is falsy", func(t *testing.T) {
		AssertValue(t, isTruthy(false), false)
	})

	t.Run("True is truthy", func(t *testing.T) {
		AssertValue(t, isTruthy(true), true)
	})

	t.Run("Zero number is truthy", func(t *testing.T) {
		AssertValue(t, isTruthy(0.0), true)
	})

	t.Run("Positive number is truthy", func(t *testing.T) {
		AssertValue(t, isTruthy(88.0), true)
	})

	t.Run("Empty string is truthy", func(t *testing.T) {
		AssertValue(t, isTruthy(""), true)
	})

	t.Run("Non-empty string is truthy", func(t *testing.T) {
		AssertValue(t, isTruthy("hello"), true)
	})
}

func TestIsEqual(t *testing.T) {
	t.Run("Nil equals nil", func(t *testing.T) {
		AssertValue(t, isEqual(nil, nil), true)
	})

	t.Run("Nil does not equal non-nil", func(t *testing.T) {
		AssertValue(t, isEqual(nil, false), false)
		AssertValue(t, isEqual(nil, 0.0), false)
		AssertValue(t, isEqual(nil, ""), false)
		AssertValue(t, isEqual(false, nil), false)
		AssertValue(t, isEqual(0.0, nil), false)
		AssertValue(t, isEqual("", nil), false)
	})

	t.Run("Same numbers are equal", func(t *testing.T) {
		AssertValue(t, isEqual(11.0, 11.0), true)
		AssertValue(t, isEqual(-10.0, -10.0), true)
		AssertValue(t, isEqual(0.0, 0.0), true)
	})

	t.Run("Different numbers are not equal", func(t *testing.T) {
		AssertValue(t, isEqual(11.0, 10.0), false)
		AssertValue(t, isEqual(-5.0, 5.0), false)
	})

	t.Run("Same strings are equal", func(t *testing.T) {
		AssertValue(t, isEqual("hello", "hello"), true)
		AssertValue(t, isEqual("", ""), true)
	})

	t.Run("Different strings are not equal", func(t *testing.T) {
		AssertValue(t, isEqual("hello", "world"), false)
	})

	t.Run("Same booleans are equal", func(t *testing.T) {
		AssertValue(t, isEqual(true, true), true)
		AssertValue(t, isEqual(false, false), true)
	})

	t.Run("Different booleans are not equal", func(t *testing.T) {
		AssertValue(t, isEqual(true, false), false)
	})

	t.Run("Different types are not equal without coercion", func(t *testing.T) {
		AssertValue(t, isEqual(42.0, "42.0"), false)
		AssertValue(t, isEqual(0.0, false), false)
		AssertValue(t, isEqual("", false), false)
		AssertValue(t, isEqual(true, 1.0), false)
	})
}

func TestCheckNumberOperand(t *testing.T) {
	op := *token.NewToken(token.MINUS, "-", nil, 1)

	t.Run("Operand is a number", func(t *testing.T) {
		num, err := checkNumberOperand(op, 10.0)
		AssertNoError(t, err)
		AssertValue(t, num, 10.0)
	})

	t.Run("Operand is not a number", func(t *testing.T) {
		_, err := checkNumberOperand(op, "not a number")
		AssertError(t, err)
	})
}

func TestCheckNumberOperands(t *testing.T) {
	op := *token.NewToken(token.PLUS, "+", nil, 1)

	t.Run("Both operands are numbers", func(t *testing.T) {
		left, right, err := checkNumberOperands(op, 10.0, 20.0)
		AssertNoError(t, err)
		AssertValue(t, left, 10.0)
		AssertValue(t, right, 20.0)
	})

	t.Run("Left operand is not a number", func(t *testing.T) {
		_, _, err := checkNumberOperands(op, "not a number", 20.0)
		AssertError(t, err)
	})

	t.Run("Right operand is not a number", func(t *testing.T) {
		_, _, err := checkNumberOperands(op, 10.0, "not a number")
		AssertError(t, err)
	})

	t.Run("Neither operand is a number", func(t *testing.T) {
		_, _, err := checkNumberOperands(op, "a", "b")
		AssertError(t, err)
	})
}

func TestEvaluateUnaryMinus(t *testing.T) {
	in := SetUpInterpreter(t)

	t.Run("Negate positive number", func(t *testing.T) {
		expr := expression.NewUnary(*token.NewToken(token.MINUS, "-", nil, 1), expression.NewLiteral(77.0))
		val, err := in.Evaluate(expr)
		AssertNoError(t, err)
		AssertValue(t, val, -77.0)
	})

	t.Run("Negate negative number", func(t *testing.T) {
		expr := expression.NewUnary(*token.NewToken(token.MINUS, "-", nil, 1), expression.NewLiteral(-77.0))
		val, err := in.Evaluate(expr)
		AssertNoError(t, err)
		AssertValue(t, val, 77.0)
	})

	t.Run("Negate zero", func(t *testing.T) {
		expr := expression.NewUnary(*token.NewToken(token.MINUS, "-", nil, 1), expression.NewLiteral(0.0))
		val, err := in.Evaluate(expr)
		AssertNoError(t, err)
		AssertValue(t, val, 0.0)
	})

	t.Run("Error when negating string", func(t *testing.T) {
		expr := expression.NewUnary(*token.NewToken(token.MINUS, "-", nil, 1), expression.NewLiteral("hello"))
		_, err := in.Evaluate(expr)
		AssertError(t, err)
	})

	t.Run("Error when negating boolean", func(t *testing.T) {
		expr := expression.NewUnary(*token.NewToken(token.MINUS, "-", nil, 1), expression.NewLiteral(true))
		_, err := in.Evaluate(expr)
		AssertError(t, err)
	})

	t.Run("Error when negating nil", func(t *testing.T) {
		expr := expression.NewUnary(*token.NewToken(token.MINUS, "-", nil, 1), expression.NewLiteral(nil))
		_, err := in.Evaluate(expr)
		AssertError(t, err)
	})
}

func TestEvaluateUnaryBang(t *testing.T) {
	in := SetUpInterpreter(t)

	t.Run("Bang boolean values", func(t *testing.T) {
		exprTrue := expression.NewUnary(*token.NewToken(token.BANG, "!", nil, 1), expression.NewLiteral(true))
		val, err := in.Evaluate(exprTrue)
		AssertNoError(t, err)
		AssertValue(t, val, false)

		exprFalse := expression.NewUnary(*token.NewToken(token.BANG, "!", nil, 1), expression.NewLiteral(false))
		val, err = in.Evaluate(exprFalse)
		AssertNoError(t, err)
		AssertValue(t, val, true)
	})

	t.Run("Bang nil", func(t *testing.T) {
		expr := expression.NewUnary(*token.NewToken(token.BANG, "!", nil, 1), expression.NewLiteral(nil))
		val, err := in.Evaluate(expr)
		AssertNoError(t, err)
		AssertValue(t, val, true)
	})

	t.Run("Bang truthy non-booleans", func(t *testing.T) {
		exprZero := expression.NewUnary(*token.NewToken(token.BANG, "!", nil, 1), expression.NewLiteral(0.0))
		val, err := in.Evaluate(exprZero)
		AssertNoError(t, err)
		AssertValue(t, val, false)

		exprStr := expression.NewUnary(*token.NewToken(token.BANG, "!", nil, 1), expression.NewLiteral(""))
		val, err = in.Evaluate(exprStr)
		AssertNoError(t, err)
		AssertValue(t, val, false)
	})

	t.Run("Propagates error from inner expression", func(t *testing.T) {
		inner := expression.NewUnary(*token.NewToken(token.MINUS, "-", nil, 1), expression.NewLiteral("hello"))
		expr := expression.NewUnary(*token.NewToken(token.BANG, "!", nil, 1), inner)
		_, err := in.Evaluate(expr)
		AssertError(t, err)
	})

	t.Run("Error when unary operator is unknown", func(t *testing.T) {
		expr := expression.NewUnary(*token.NewToken(token.EOF, "EOF", nil, 1), expression.NewLiteral(10.0))
		_, err := in.Evaluate(expr)
		AssertError(t, err)
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
		AssertError(t, err)
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
		AssertError(t, err)
	})

	t.Run("Error when binary operator is unknown", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(10.0),
			*token.NewToken(token.EOF, "EOF", nil, 1),
			expression.NewLiteral(20.0),
		)
		_, err := in.Evaluate(expr)
		AssertError(t, err)
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
		AssertNoError(t, err)
		AssertValue(t, val, 33.0)
	})

	t.Run("Sum of negative numbers", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(-10.0),
			*token.NewToken(token.PLUS, "+", nil, 1),
			expression.NewLiteral(-5.0),
		)
		val, err := in.Evaluate(expr)
		AssertNoError(t, err)
		AssertValue(t, val, -15.0)
	})

	t.Run("Sum of positive and negative number", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(10.0),
			*token.NewToken(token.PLUS, "+", nil, 1),
			expression.NewLiteral(-3.0),
		)
		val, err := in.Evaluate(expr)
		AssertNoError(t, err)
		AssertValue(t, val, 7.0)
	})

	t.Run("Sum of a negative number and a positive number", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(-7.0),
			*token.NewToken(token.PLUS, "+", nil, 1),
			expression.NewLiteral(2.0),
		)
		val, err := in.Evaluate(expr)
		AssertNoError(t, err)
		AssertValue(t, val, -5.0)
	})

	t.Run("Concatenation of two strings", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral("Hello "),
			*token.NewToken(token.PLUS, "+", nil, 1),
			expression.NewLiteral("world"),
		)
		val, err := in.Evaluate(expr)
		AssertNoError(t, err)
		AssertValue(t, val, "Hello world")
	})

	t.Run("Concatenation with empty string", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(""),
			*token.NewToken(token.PLUS, "+", nil, 1),
			expression.NewLiteral("abc"),
		)
		val, err := in.Evaluate(expr)
		AssertNoError(t, err)
		AssertValue(t, val, "abc")
	})

	t.Run("Error when adding string and number", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral("a"),
			*token.NewToken(token.PLUS, "+", nil, 1),
			expression.NewLiteral(1.0),
		)
		_, err := in.Evaluate(expr)
		AssertError(t, err)
	})

	t.Run("Error when adding number and string", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(1.0),
			*token.NewToken(token.PLUS, "+", nil, 1),
			expression.NewLiteral("a"),
		)
		_, err := in.Evaluate(expr)
		AssertError(t, err)
	})

	t.Run("Error when adding incompatible types", func(t *testing.T) {
		exprBool := expression.NewBinary(
			expression.NewLiteral(true),
			*token.NewToken(token.PLUS, "+", nil, 1),
			expression.NewLiteral(false),
		)
		_, err := in.Evaluate(exprBool)
		AssertError(t, err)

		exprNil := expression.NewBinary(
			expression.NewLiteral(nil),
			*token.NewToken(token.PLUS, "+", nil, 1),
			expression.NewLiteral(nil),
		)
		_, err = in.Evaluate(exprNil)
		AssertError(t, err)
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
		AssertNoError(t, err)
		AssertValue(t, val, 10.0)
	})

	t.Run("Subtraction of negative numbers", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(-1.0),
			*token.NewToken(token.MINUS, "-", nil, 1),
			expression.NewLiteral(-5.0),
		)
		val, err := in.Evaluate(expr)
		AssertNoError(t, err)
		AssertValue(t, val, 4.0)
	})

	t.Run("Subtraction with negative result", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(5.0),
			*token.NewToken(token.MINUS, "-", nil, 1),
			expression.NewLiteral(10.0),
		)
		val, err := in.Evaluate(expr)
		AssertNoError(t, err)
		AssertValue(t, val, -5.0)
	})

	t.Run("Error when left operand is not a number", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral("a"),
			*token.NewToken(token.MINUS, "-", nil, 1),
			expression.NewLiteral(10.0),
		)
		_, err := in.Evaluate(expr)
		AssertError(t, err)
	})

	t.Run("Error when right operand is not a number", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(10.0),
			*token.NewToken(token.MINUS, "-", nil, 1),
			expression.NewLiteral("a"),
		)
		_, err := in.Evaluate(expr)
		AssertError(t, err)
	})

	t.Run("Error when operands are booleans", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(true),
			*token.NewToken(token.MINUS, "-", nil, 1),
			expression.NewLiteral(false),
		)
		_, err := in.Evaluate(expr)
		AssertError(t, err)
	})

	t.Run("Error when operands are nil", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(nil),
			*token.NewToken(token.MINUS, "-", nil, 1),
			expression.NewLiteral(nil),
		)
		_, err := in.Evaluate(expr)
		AssertError(t, err)
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
		AssertNoError(t, err)
		AssertValue(t, val, 20.0)
	})

	t.Run("Multiplication of negative numbers", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(-2.0),
			*token.NewToken(token.STAR, "*", nil, 1),
			expression.NewLiteral(-5.0),
		)
		val, err := in.Evaluate(expr)
		AssertNoError(t, err)
		AssertValue(t, val, 10.0)
	})

	t.Run("Multiplication with negative result", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(5.0),
			*token.NewToken(token.STAR, "*", nil, 1),
			expression.NewLiteral(-10.0),
		)
		val, err := in.Evaluate(expr)
		AssertNoError(t, err)
		AssertValue(t, val, -50.0)
	})

	t.Run("Error when left operand is not a number", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral("a"),
			*token.NewToken(token.STAR, "*", nil, 1),
			expression.NewLiteral(10.0),
		)
		_, err := in.Evaluate(expr)
		AssertError(t, err)
	})

	t.Run("Error when right operand is not a number", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(10.0),
			*token.NewToken(token.STAR, "*", nil, 1),
			expression.NewLiteral("a"),
		)
		_, err := in.Evaluate(expr)
		AssertError(t, err)
	})

	t.Run("Error when operands are booleans", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(true),
			*token.NewToken(token.STAR, "*", nil, 1),
			expression.NewLiteral(false),
		)
		_, err := in.Evaluate(expr)
		AssertError(t, err)
	})

	t.Run("Error when operands are nil", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(nil),
			*token.NewToken(token.STAR, "*", nil, 1),
			expression.NewLiteral(nil),
		)
		_, err := in.Evaluate(expr)
		AssertError(t, err)
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
		AssertNoError(t, err)
		AssertValue(t, val, 5.0)
	})

	t.Run("Division of negative numbers", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(-10.0),
			*token.NewToken(token.SLASH, "/", nil, 1),
			expression.NewLiteral(-2.0),
		)
		val, err := in.Evaluate(expr)
		AssertNoError(t, err)
		AssertValue(t, val, 5.0)
	})

	t.Run("Division with negative result", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(10.0),
			*token.NewToken(token.SLASH, "/", nil, 1),
			expression.NewLiteral(-2.0),
		)
		val, err := in.Evaluate(expr)
		AssertNoError(t, err)
		AssertValue(t, val, -5.0)
	})

	t.Run("Division with fractional result", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(7.0),
			*token.NewToken(token.SLASH, "/", nil, 1),
			expression.NewLiteral(2.0),
		)
		val, err := in.Evaluate(expr)
		AssertNoError(t, err)
		AssertValue(t, val, 3.5)
	})

	t.Run("Error when left operand is not a number", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral("a"),
			*token.NewToken(token.SLASH, "/", nil, 1),
			expression.NewLiteral(10.0),
		)
		_, err := in.Evaluate(expr)
		AssertError(t, err)
	})

	t.Run("Error when right operand is not a number", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(10.0),
			*token.NewToken(token.SLASH, "/", nil, 1),
			expression.NewLiteral("a"),
		)
		_, err := in.Evaluate(expr)
		AssertError(t, err)
	})

	t.Run("A positive number divided by zero produces positive infinity", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(10.0),
			*token.NewToken(token.SLASH, "/", nil, 1),
			expression.NewLiteral(0.0),
		)
		val, err := in.Evaluate(expr)
		AssertNoError(t, err)
		AssertValue(t, val, math.Inf(1))
	})

	t.Run("A negative number divided by zero produces negative infinity", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(-10.0),
			*token.NewToken(token.SLASH, "/", nil, 1),
			expression.NewLiteral(0.0),
		)
		val, err := in.Evaluate(expr)
		AssertNoError(t, err)
		AssertValue(t, val, math.Inf(-1))
	})

	t.Run("Zero divided by zero produces NaN", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(0.0),
			*token.NewToken(token.SLASH, "/", nil, 1),
			expression.NewLiteral(0.0),
		)
		val, err := in.Evaluate(expr)
		AssertNoError(t, err)

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
		AssertError(t, err)
	})

	t.Run("Error when operands are nil", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(nil),
			*token.NewToken(token.SLASH, "/", nil, 1),
			expression.NewLiteral(nil),
		)
		_, err := in.Evaluate(expr)
		AssertError(t, err)
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
		AssertNoError(t, err)
		AssertValue(t, val, 1.0)
	})

	t.Run("Modulo with negative dividend", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(-10.0),
			*token.NewToken(token.PERCENT, "%", nil, 1),
			expression.NewLiteral(3.0),
		)
		val, err := in.Evaluate(expr)
		AssertNoError(t, err)
		AssertValue(t, val, -1.0)
	})

	t.Run("Exact modulo", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(10.0),
			*token.NewToken(token.PERCENT, "%", nil, 1),
			expression.NewLiteral(5.0),
		)
		val, err := in.Evaluate(expr)
		AssertNoError(t, err)
		AssertValue(t, val, 0.0)
	})

	t.Run("Modulo with negative divisor", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(10.0),
			*token.NewToken(token.PERCENT, "%", nil, 1),
			expression.NewLiteral(-3.0),
		)
		val, err := in.Evaluate(expr)
		AssertNoError(t, err)
		AssertValue(t, val, 1.0)
	})

	t.Run("Modulo of two negative numbers", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(-10.0),
			*token.NewToken(token.PERCENT, "%", nil, 1),
			expression.NewLiteral(-3.0),
		)
		val, err := in.Evaluate(expr)
		AssertNoError(t, err)
		AssertValue(t, val, -1.0)
	})

	t.Run("Zero modulo positive number produces zero", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(0.0),
			*token.NewToken(token.PERCENT, "%", nil, 1),
			expression.NewLiteral(5.0),
		)
		val, err := in.Evaluate(expr)
		AssertNoError(t, err)
		AssertValue(t, val, 0.0)
	})

	t.Run("Modulo of floating point numbers", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(5.5),
			*token.NewToken(token.PERCENT, "%", nil, 1),
			expression.NewLiteral(2.0),
		)
		val, err := in.Evaluate(expr)
		AssertNoError(t, err)
		AssertValue(t, val, 1.5)
	})

	t.Run("Modulo by zero produces NaN", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(10.0),
			*token.NewToken(token.PERCENT, "%", nil, 1),
			expression.NewLiteral(0.0),
		)
		val, err := in.Evaluate(expr)
		AssertNoError(t, err)

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
		AssertNoError(t, err)

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
		AssertError(t, err)
	})

	t.Run("Error when right operand is not a number", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(10.0),
			*token.NewToken(token.PERCENT, "%", nil, 1),
			expression.NewLiteral("a"),
		)
		_, err := in.Evaluate(expr)
		AssertError(t, err)
	})

	t.Run("Error when operands are booleans", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(true),
			*token.NewToken(token.PERCENT, "%", nil, 1),
			expression.NewLiteral(false),
		)
		_, err := in.Evaluate(expr)
		AssertError(t, err)
	})

	t.Run("Error when operands are nil", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(nil),
			*token.NewToken(token.PERCENT, "%", nil, 1),
			expression.NewLiteral(nil),
		)
		_, err := in.Evaluate(expr)
		AssertError(t, err)
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
		AssertNoError(t, err)
		AssertValue(t, val, true)
	})

	t.Run("Greater than with two positive numbers when false", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(5.0),
			*token.NewToken(token.GREATER, ">", nil, 1),
			expression.NewLiteral(10.0),
		)
		val, err := in.Evaluate(expr)
		AssertNoError(t, err)
		AssertValue(t, val, false)
	})

	t.Run("Greater than with negative numbers when true", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(-5.0),
			*token.NewToken(token.GREATER, ">", nil, 1),
			expression.NewLiteral(-10.0),
		)
		val, err := in.Evaluate(expr)
		AssertNoError(t, err)
		AssertValue(t, val, true)
	})

	t.Run("Greater than with negative numbers when false", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(-10.0),
			*token.NewToken(token.GREATER, ">", nil, 1),
			expression.NewLiteral(-5.0),
		)
		val, err := in.Evaluate(expr)
		AssertNoError(t, err)
		AssertValue(t, val, false)
	})

	t.Run("Greater than with two equal numbers", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(10.0),
			*token.NewToken(token.GREATER, ">", nil, 1),
			expression.NewLiteral(10.0),
		)
		val, err := in.Evaluate(expr)
		AssertNoError(t, err)
		AssertValue(t, val, false)
	})

	t.Run("Error when left operand is not a number", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral("a"),
			*token.NewToken(token.GREATER, ">", nil, 1),
			expression.NewLiteral(10.0),
		)
		_, err := in.Evaluate(expr)
		AssertError(t, err)
	})

	t.Run("Error when right operand is not a number", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(10.0),
			*token.NewToken(token.GREATER, ">", nil, 1),
			expression.NewLiteral("a"),
		)
		_, err := in.Evaluate(expr)
		AssertError(t, err)
	})

	t.Run("Error when comparing strings", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral("a"),
			*token.NewToken(token.GREATER, ">", nil, 1),
			expression.NewLiteral("b"),
		)
		_, err := in.Evaluate(expr)
		AssertError(t, err)
	})

	t.Run("Error when comparing booleans", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(true),
			*token.NewToken(token.GREATER, ">", nil, 1),
			expression.NewLiteral(false),
		)
		_, err := in.Evaluate(expr)
		AssertError(t, err)
	})

	t.Run("Error when comparing nil values", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(nil),
			*token.NewToken(token.GREATER, ">", nil, 1),
			expression.NewLiteral(nil),
		)
		_, err := in.Evaluate(expr)
		AssertError(t, err)
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
		AssertNoError(t, err)
		AssertValue(t, val, true)
	})

	t.Run("Greater equal than with two positive numbers when false", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(5.0),
			*token.NewToken(token.GREATER_EQUAL, ">=", nil, 1),
			expression.NewLiteral(10.0),
		)
		val, err := in.Evaluate(expr)
		AssertNoError(t, err)
		AssertValue(t, val, false)
	})

	t.Run("Greater equal than with two equal positive numbers", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(10.0),
			*token.NewToken(token.GREATER_EQUAL, ">=", nil, 1),
			expression.NewLiteral(10.0),
		)
		val, err := in.Evaluate(expr)
		AssertNoError(t, err)
		AssertValue(t, val, true)
	})

	t.Run("Greater equal than with negative numbers when true", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(-5.0),
			*token.NewToken(token.GREATER_EQUAL, ">=", nil, 1),
			expression.NewLiteral(-10.0),
		)
		val, err := in.Evaluate(expr)
		AssertNoError(t, err)
		AssertValue(t, val, true)
	})

	t.Run("Greater equal than with two equal negative numbers", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(-10.0),
			*token.NewToken(token.GREATER_EQUAL, ">=", nil, 1),
			expression.NewLiteral(-10.0),
		)
		val, err := in.Evaluate(expr)
		AssertNoError(t, err)
		AssertValue(t, val, true)
	})

	t.Run("Greater equal than with negative numbers when false", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(-10.0),
			*token.NewToken(token.GREATER_EQUAL, ">=", nil, 1),
			expression.NewLiteral(-5.0),
		)
		val, err := in.Evaluate(expr)
		AssertNoError(t, err)
		AssertValue(t, val, false)
	})

	t.Run("Error when left operand is not a number", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral("a"),
			*token.NewToken(token.GREATER_EQUAL, ">=", nil, 1),
			expression.NewLiteral(10.0),
		)
		_, err := in.Evaluate(expr)
		AssertError(t, err)
	})

	t.Run("Error when right operand is not a number", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(10.0),
			*token.NewToken(token.GREATER_EQUAL, ">=", nil, 1),
			expression.NewLiteral("a"),
		)
		_, err := in.Evaluate(expr)
		AssertError(t, err)
	})

	t.Run("Error when comparing strings", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral("a"),
			*token.NewToken(token.GREATER_EQUAL, ">=", nil, 1),
			expression.NewLiteral("b"),
		)
		_, err := in.Evaluate(expr)
		AssertError(t, err)
	})

	t.Run("Error when comparing booleans", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(true),
			*token.NewToken(token.GREATER_EQUAL, ">=", nil, 1),
			expression.NewLiteral(false),
		)
		_, err := in.Evaluate(expr)
		AssertError(t, err)
	})

	t.Run("Error when comparing nil values", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(nil),
			*token.NewToken(token.GREATER_EQUAL, ">=", nil, 1),
			expression.NewLiteral(nil),
		)
		_, err := in.Evaluate(expr)
		AssertError(t, err)
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
		AssertNoError(t, err)
		AssertValue(t, val, true)
	})

	t.Run("Less than with two positive numbers when false", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(50.0),
			*token.NewToken(token.LESS, "<", nil, 1),
			expression.NewLiteral(10.0),
		)
		val, err := in.Evaluate(expr)
		AssertNoError(t, err)
		AssertValue(t, val, false)
	})

	t.Run("Less than with negative numbers when true", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(-50.0),
			*token.NewToken(token.LESS, "<", nil, 1),
			expression.NewLiteral(-10.0),
		)
		val, err := in.Evaluate(expr)
		AssertNoError(t, err)
		AssertValue(t, val, true)
	})

	t.Run("Less than with negative numbers when false", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(-10.0),
			*token.NewToken(token.LESS, "<", nil, 1),
			expression.NewLiteral(-50.0),
		)
		val, err := in.Evaluate(expr)
		AssertNoError(t, err)
		AssertValue(t, val, false)
	})

	t.Run("Less than with two equal numbers", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(10.0),
			*token.NewToken(token.LESS, "<", nil, 1),
			expression.NewLiteral(10.0),
		)
		val, err := in.Evaluate(expr)
		AssertNoError(t, err)
		AssertValue(t, val, false)
	})

	t.Run("Error when left operand is not a number", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral("a"),
			*token.NewToken(token.LESS, "<", nil, 1),
			expression.NewLiteral(10.0),
		)
		_, err := in.Evaluate(expr)
		AssertError(t, err)
	})

	t.Run("Error when right operand is not a number", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(10.0),
			*token.NewToken(token.LESS, "<", nil, 1),
			expression.NewLiteral("a"),
		)
		_, err := in.Evaluate(expr)
		AssertError(t, err)
	})

	t.Run("Error when comparing strings", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral("a"),
			*token.NewToken(token.LESS, "<", nil, 1),
			expression.NewLiteral("b"),
		)
		_, err := in.Evaluate(expr)
		AssertError(t, err)
	})

	t.Run("Error when comparing booleans", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(true),
			*token.NewToken(token.LESS, "<", nil, 1),
			expression.NewLiteral(false),
		)
		_, err := in.Evaluate(expr)
		AssertError(t, err)
	})

	t.Run("Error when comparing nil values", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(nil),
			*token.NewToken(token.LESS, "<", nil, 1),
			expression.NewLiteral(nil),
		)
		_, err := in.Evaluate(expr)
		AssertError(t, err)
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
		AssertNoError(t, err)
		AssertValue(t, val, true)
	})

	t.Run("Less equal than with two positive numbers when false", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(50.0),
			*token.NewToken(token.LESS_EQUAL, "<=", nil, 1),
			expression.NewLiteral(10.0),
		)
		val, err := in.Evaluate(expr)
		AssertNoError(t, err)
		AssertValue(t, val, false)
	})

	t.Run("Less equal than with two equal positive numbers", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(10.0),
			*token.NewToken(token.LESS_EQUAL, "<=", nil, 1),
			expression.NewLiteral(10.0),
		)
		val, err := in.Evaluate(expr)
		AssertNoError(t, err)
		AssertValue(t, val, true)
	})

	t.Run("Less equal than with negative numbers when true", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(-50.0),
			*token.NewToken(token.LESS_EQUAL, "<=", nil, 1),
			expression.NewLiteral(-10.0),
		)
		val, err := in.Evaluate(expr)
		AssertNoError(t, err)
		AssertValue(t, val, true)
	})

	t.Run("Less equal than with two equal negative numbers", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(-10.0),
			*token.NewToken(token.LESS_EQUAL, "<=", nil, 1),
			expression.NewLiteral(-10.0),
		)
		val, err := in.Evaluate(expr)
		AssertNoError(t, err)
		AssertValue(t, val, true)
	})

	t.Run("Less equal than with negative numbers when false", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(-10.0),
			*token.NewToken(token.LESS_EQUAL, "<=", nil, 1),
			expression.NewLiteral(-50.0),
		)
		val, err := in.Evaluate(expr)
		AssertNoError(t, err)
		AssertValue(t, val, false)
	})

	t.Run("Error when left operand is not a number", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral("a"),
			*token.NewToken(token.LESS_EQUAL, "<=", nil, 1),
			expression.NewLiteral(10.0),
		)
		_, err := in.Evaluate(expr)
		AssertError(t, err)
	})

	t.Run("Error when right operand is not a number", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(10.0),
			*token.NewToken(token.LESS_EQUAL, "<=", nil, 1),
			expression.NewLiteral("a"),
		)
		_, err := in.Evaluate(expr)
		AssertError(t, err)
	})

	t.Run("Error when comparing strings", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral("a"),
			*token.NewToken(token.LESS_EQUAL, "<=", nil, 1),
			expression.NewLiteral("b"),
		)
		_, err := in.Evaluate(expr)
		AssertError(t, err)
	})

	t.Run("Error when comparing booleans", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(true),
			*token.NewToken(token.LESS_EQUAL, "<=", nil, 1),
			expression.NewLiteral(false),
		)
		_, err := in.Evaluate(expr)
		AssertError(t, err)
	})

	t.Run("Error when comparing nil values", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(nil),
			*token.NewToken(token.LESS_EQUAL, "<=", nil, 1),
			expression.NewLiteral(nil),
		)
		_, err := in.Evaluate(expr)
		AssertError(t, err)
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
		AssertNoError(t, err)
		AssertValue(t, val, true)
	})

	t.Run("Equal numbers when false", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(42.0),
			*token.NewToken(token.EQUAL_EQUAL, "==", nil, 1),
			expression.NewLiteral(10.0),
		)
		val, err := in.Evaluate(expr)
		AssertNoError(t, err)
		AssertValue(t, val, false)
	})

	t.Run("Equal negative numbers", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(-10.0),
			*token.NewToken(token.EQUAL_EQUAL, "==", nil, 1),
			expression.NewLiteral(-10.0),
		)
		val, err := in.Evaluate(expr)
		AssertNoError(t, err)
		AssertValue(t, val, true)
	})

	t.Run("Equal strings when true", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral("hello"),
			*token.NewToken(token.EQUAL_EQUAL, "==", nil, 1),
			expression.NewLiteral("hello"),
		)
		val, err := in.Evaluate(expr)
		AssertNoError(t, err)
		AssertValue(t, val, true)
	})

	t.Run("Equal strings when false", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral("hello"),
			*token.NewToken(token.EQUAL_EQUAL, "==", nil, 1),
			expression.NewLiteral("world"),
		)
		val, err := in.Evaluate(expr)
		AssertNoError(t, err)
		AssertValue(t, val, false)
	})

	t.Run("Equal booleans when true", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(true),
			*token.NewToken(token.EQUAL_EQUAL, "==", nil, 1),
			expression.NewLiteral(true),
		)
		val, err := in.Evaluate(expr)
		AssertNoError(t, err)
		AssertValue(t, val, true)
	})

	t.Run("Equal booleans when false", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(true),
			*token.NewToken(token.EQUAL_EQUAL, "==", nil, 1),
			expression.NewLiteral(false),
		)
		val, err := in.Evaluate(expr)
		AssertNoError(t, err)
		AssertValue(t, val, false)
	})

	t.Run("Equal nil with nil", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(nil),
			*token.NewToken(token.EQUAL_EQUAL, "==", nil, 1),
			expression.NewLiteral(nil),
		)
		val, err := in.Evaluate(expr)
		AssertNoError(t, err)
		AssertValue(t, val, true)
	})

	t.Run("Equal nil with non-nil produces false", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(nil),
			*token.NewToken(token.EQUAL_EQUAL, "==", nil, 1),
			expression.NewLiteral(0.0),
		)
		val, err := in.Evaluate(expr)
		AssertNoError(t, err)
		AssertValue(t, val, false)
	})

	t.Run("Equal with different types produces false without error", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(42.0),
			*token.NewToken(token.EQUAL_EQUAL, "==", nil, 1),
			expression.NewLiteral("42.0"),
		)
		val, err := in.Evaluate(expr)
		AssertNoError(t, err)
		AssertValue(t, val, false)
	})

	t.Run("Equal boolean and number produces false without error", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(true),
			*token.NewToken(token.EQUAL_EQUAL, "==", nil, 1),
			expression.NewLiteral(1.0),
		)
		val, err := in.Evaluate(expr)
		AssertNoError(t, err)
		AssertValue(t, val, false)
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
		AssertNoError(t, err)
		AssertValue(t, val, false)
	})

	t.Run("Bang equal numbers when true", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(42.0),
			*token.NewToken(token.BANG_EQUAL, "!=", nil, 1),
			expression.NewLiteral(10.0),
		)
		val, err := in.Evaluate(expr)
		AssertNoError(t, err)
		AssertValue(t, val, true)
	})

	t.Run("Bang equal negative numbers", func(t *testing.T) {
		exprEqual := expression.NewBinary(
			expression.NewLiteral(-10.0),
			*token.NewToken(token.BANG_EQUAL, "!=", nil, 1),
			expression.NewLiteral(-10.0),
		)
		val, err := in.Evaluate(exprEqual)
		AssertNoError(t, err)
		AssertValue(t, val, false)

		exprDiff := expression.NewBinary(
			expression.NewLiteral(-10.0),
			*token.NewToken(token.BANG_EQUAL, "!=", nil, 1),
			expression.NewLiteral(-5.0),
		)
		val, err = in.Evaluate(exprDiff)
		AssertNoError(t, err)
		AssertValue(t, val, true)
	})

	t.Run("Bang equal strings when false", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral("hello"),
			*token.NewToken(token.BANG_EQUAL, "!=", nil, 1),
			expression.NewLiteral("hello"),
		)
		val, err := in.Evaluate(expr)
		AssertNoError(t, err)
		AssertValue(t, val, false)
	})

	t.Run("Bang equal strings when true", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral("hello"),
			*token.NewToken(token.BANG_EQUAL, "!=", nil, 1),
			expression.NewLiteral("world"),
		)
		val, err := in.Evaluate(expr)
		AssertNoError(t, err)
		AssertValue(t, val, true)
	})

	t.Run("Bang equal booleans when false", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(true),
			*token.NewToken(token.BANG_EQUAL, "!=", nil, 1),
			expression.NewLiteral(true),
		)
		val, err := in.Evaluate(expr)
		AssertNoError(t, err)
		AssertValue(t, val, false)
	})

	t.Run("Bang equal booleans when true", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(true),
			*token.NewToken(token.BANG_EQUAL, "!=", nil, 1),
			expression.NewLiteral(false),
		)
		val, err := in.Evaluate(expr)
		AssertNoError(t, err)
		AssertValue(t, val, true)
	})

	t.Run("Bang equal nil with nil produces false", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(nil),
			*token.NewToken(token.BANG_EQUAL, "!=", nil, 1),
			expression.NewLiteral(nil),
		)
		val, err := in.Evaluate(expr)
		AssertNoError(t, err)
		AssertValue(t, val, false)
	})

	t.Run("Bang equal nil with non-nil produces true", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(nil),
			*token.NewToken(token.BANG_EQUAL, "!=", nil, 1),
			expression.NewLiteral(0.0),
		)
		val, err := in.Evaluate(expr)
		AssertNoError(t, err)
		AssertValue(t, val, true)
	})

	t.Run("Bang equal with different types produces true without error", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(42.0),
			*token.NewToken(token.BANG_EQUAL, "!=", nil, 1),
			expression.NewLiteral("42.0"),
		)
		val, err := in.Evaluate(expr)
		AssertNoError(t, err)
		AssertValue(t, val, true)
	})

	t.Run("Bang equal boolean and number produces true without error", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(true),
			*token.NewToken(token.BANG_EQUAL, "!=", nil, 1),
			expression.NewLiteral(1.0),
		)
		val, err := in.Evaluate(expr)
		AssertNoError(t, err)
		AssertValue(t, val, true)
	})
}
