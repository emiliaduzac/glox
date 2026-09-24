package interpreter

import (
	"math"
	"testing"
	"tp-lox/internal/expression"
	"tp-lox/internal/token"
)

func setUpInterpreter(t *testing.T) *Interpreter {
	t.Helper()
	return NewInterpreter()
}

func assertValue(t *testing.T, got any, expected any) {
	t.Helper()
	if got != expected {
		t.Errorf("got %v, expected %v", got, expected)
	}
}

func assertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func assertError(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestEvaluateLiterals(t *testing.T) {
	in := setUpInterpreter(t)

	t.Run("Evaluate positive integer literal", func(t *testing.T) {
		expression := expression.NewLiteral(99.0)
		value, err := in.Evaluate(expression)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		assertValue(t, value, 99.0)
	})

	t.Run("Evaluate empty string literal", func(t *testing.T) {
		expression := expression.NewLiteral("")
		value, err := in.Evaluate(expression)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		assertValue(t, value, "")
	})

	t.Run("Evaluate string literal", func(t *testing.T) {
		expression := expression.NewLiteral("Hello World")
		value, err := in.Evaluate(expression)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		assertValue(t, value, "Hello World")
	})

	t.Run("Evaluate boolean literal", func(t *testing.T) {
		expression := expression.NewLiteral(true)
		value, err := in.Evaluate(expression)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		assertValue(t, value, true)
	})

	t.Run("Evaluate false boolean literal", func(t *testing.T) {
		expression := expression.NewLiteral(false)
		value, err := in.Evaluate(expression)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		assertValue(t, value, false)
	})

	t.Run("Evaluate nil literal", func(t *testing.T) {
		expression := expression.NewLiteral(nil)
		value, err := in.Evaluate(expression)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		assertValue(t, value, nil)
	})
}

func TestStringify(t *testing.T) {
	t.Run("Stringify nil value", func(t *testing.T) {
		assertValue(t, Stringify(nil), "nil")
	})

	t.Run("Stringify integer value", func(t *testing.T) {
		assertValue(t, Stringify(99), "99")
	})

	t.Run("Stringify string value", func(t *testing.T) {
		assertValue(t, Stringify("Hello"), "Hello")
	})

	t.Run("Stringify boolean true value", func(t *testing.T) {
		assertValue(t, Stringify(true), "true")
	})

	t.Run("Stringify boolean false value", func(t *testing.T) {
		assertValue(t, Stringify(false), "false")
	})
}

func TestEvaluateGrouping(t *testing.T) {
	in := setUpInterpreter(t)

	t.Run("Evaluate grouping with literal", func(t *testing.T) {
		expression := expression.NewGrouping(expression.NewLiteral(99))
		value, err := in.Evaluate(expression)
		assertNoError(t, err)
		assertValue(t, value, 99)
	})

	t.Run("Evaluate gruped string literal", func(t *testing.T) {
		expression := expression.NewGrouping(expression.NewLiteral("Hello"))
		value, err := in.Evaluate(expression)
		assertNoError(t, err)
		assertValue(t, value, "Hello")
	})

	t.Run("Evaluate nested grouping with literal", func(t *testing.T) {
		expression := expression.NewGrouping(expression.NewGrouping(expression.NewGrouping(expression.NewLiteral(88))))
		value, err := in.Evaluate(expression)
		assertNoError(t, err)
		assertValue(t, value, 88)
	})
}

func TestIsTruthy(t *testing.T) {
	t.Run("Nil is falsy", func(t *testing.T) {
		assertValue(t, isTruthy(nil), false)
	})

	t.Run("False is falsy", func(t *testing.T) {
		assertValue(t, isTruthy(false), false)
	})

	t.Run("True is truthy", func(t *testing.T) {
		assertValue(t, isTruthy(true), true)
	})

	t.Run("Zero number is truthy", func(t *testing.T) {
		assertValue(t, isTruthy(0.0), true)
	})

	t.Run("Positive number is truthy", func(t *testing.T) {
		assertValue(t, isTruthy(88.0), true)
	})

	t.Run("Empty string is truthy", func(t *testing.T) {
		assertValue(t, isTruthy(""), true)
	})

	t.Run("Non-empty string is truthy", func(t *testing.T) {
		assertValue(t, isTruthy("hello"), true)
	})
}

func TestEvaluateUnaryMinus(t *testing.T) {
	in := setUpInterpreter(t)

	t.Run("Negate positive number", func(t *testing.T) {
		expr := expression.NewUnary(*token.NewToken(token.MINUS, "-", nil, 1), expression.NewLiteral(77.0))
		val, err := in.Evaluate(expr)
		assertNoError(t, err)
		assertValue(t, val, -77.0)
	})

	t.Run("Negate negative number", func(t *testing.T) {
		expr := expression.NewUnary(*token.NewToken(token.MINUS, "-", nil, 1), expression.NewLiteral(-77.0))
		val, err := in.Evaluate(expr)
		assertNoError(t, err)
		assertValue(t, val, 77.0)
	})

	t.Run("Negate zero", func(t *testing.T) {
		expr := expression.NewUnary(*token.NewToken(token.MINUS, "-", nil, 1), expression.NewLiteral(0.0))
		val, err := in.Evaluate(expr)
		assertNoError(t, err)
		assertValue(t, val, 0.0)
	})

	t.Run("Error when negating string", func(t *testing.T) {
		expr := expression.NewUnary(*token.NewToken(token.MINUS, "-", nil, 1), expression.NewLiteral("hello"))
		_, err := in.Evaluate(expr)
		assertError(t, err)
	})

	t.Run("Error when negating boolean", func(t *testing.T) {
		expr := expression.NewUnary(*token.NewToken(token.MINUS, "-", nil, 1), expression.NewLiteral(true))
		_, err := in.Evaluate(expr)
		assertError(t, err)
	})

	t.Run("Error when negating nil", func(t *testing.T) {
		expr := expression.NewUnary(*token.NewToken(token.MINUS, "-", nil, 1), expression.NewLiteral(nil))
		_, err := in.Evaluate(expr)
		assertError(t, err)
	})
}

func TestEvaluateUnaryBang(t *testing.T) {
	in := setUpInterpreter(t)

	t.Run("Bang boolean values", func(t *testing.T) {
		exprTrue := expression.NewUnary(*token.NewToken(token.BANG, "!", nil, 1), expression.NewLiteral(true))
		val, err := in.Evaluate(exprTrue)
		assertNoError(t, err)
		assertValue(t, val, false)

		exprFalse := expression.NewUnary(*token.NewToken(token.BANG, "!", nil, 1), expression.NewLiteral(false))
		val, err = in.Evaluate(exprFalse)
		assertNoError(t, err)
		assertValue(t, val, true)
	})

	t.Run("Bang nil", func(t *testing.T) {
		expr := expression.NewUnary(*token.NewToken(token.BANG, "!", nil, 1), expression.NewLiteral(nil))
		val, err := in.Evaluate(expr)
		assertNoError(t, err)
		assertValue(t, val, true)
	})

	t.Run("Bang truthy non-booleans", func(t *testing.T) {
		exprZero := expression.NewUnary(*token.NewToken(token.BANG, "!", nil, 1), expression.NewLiteral(0.0))
		val, err := in.Evaluate(exprZero)
		assertNoError(t, err)
		assertValue(t, val, false)

		exprStr := expression.NewUnary(*token.NewToken(token.BANG, "!", nil, 1), expression.NewLiteral(""))
		val, err = in.Evaluate(exprStr)
		assertNoError(t, err)
		assertValue(t, val, false)
	})

	t.Run("Propagates error from inner expression", func(t *testing.T) {
		inner := expression.NewUnary(*token.NewToken(token.MINUS, "-", nil, 1), expression.NewLiteral("hello"))
		expr := expression.NewUnary(*token.NewToken(token.BANG, "!", nil, 1), inner)
		_, err := in.Evaluate(expr)
		assertError(t, err)
	})
}

func TestEvaluateBinary(t *testing.T) {
	in := setUpInterpreter(t)

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
		assertError(t, err)
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
		assertError(t, err)
	})
}

func TestEvaluateBinaryPlus(t *testing.T) {
	in := setUpInterpreter(t)

	t.Run("Sum of two positive numbers", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(11.0),
			*token.NewToken(token.PLUS, "+", nil, 1),
			expression.NewLiteral(22.0),
		)
		val, err := in.Evaluate(expr)
		assertNoError(t, err)
		assertValue(t, val, 33.0)
	})

	t.Run("Sum of negative numbers", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(-10.0),
			*token.NewToken(token.PLUS, "+", nil, 1),
			expression.NewLiteral(-5.0),
		)
		val, err := in.Evaluate(expr)
		assertNoError(t, err)
		assertValue(t, val, -15.0)
	})

	t.Run("Sum of positive and negative number", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(10.0),
			*token.NewToken(token.PLUS, "+", nil, 1),
			expression.NewLiteral(-3.0),
		)
		val, err := in.Evaluate(expr)
		assertNoError(t, err)
		assertValue(t, val, 7.0)
	})

	t.Run("Sum of a negative number and a positive number", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(-7.0),
			*token.NewToken(token.PLUS, "+", nil, 1),
			expression.NewLiteral(2.0),
		)
		val, err := in.Evaluate(expr)
		assertNoError(t, err)
		assertValue(t, val, -5.0)
	})

	t.Run("Concatenation of two strings", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral("Hello "),
			*token.NewToken(token.PLUS, "+", nil, 1),
			expression.NewLiteral("world"),
		)
		val, err := in.Evaluate(expr)
		assertNoError(t, err)
		assertValue(t, val, "Hello world")
	})

	t.Run("Concatenation with empty string", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(""),
			*token.NewToken(token.PLUS, "+", nil, 1),
			expression.NewLiteral("abc"),
		)
		val, err := in.Evaluate(expr)
		assertNoError(t, err)
		assertValue(t, val, "abc")
	})

	t.Run("Error when adding string and number", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral("a"),
			*token.NewToken(token.PLUS, "+", nil, 1),
			expression.NewLiteral(1.0),
		)
		_, err := in.Evaluate(expr)
		assertError(t, err)
	})

	t.Run("Error when adding number and string", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(1.0),
			*token.NewToken(token.PLUS, "+", nil, 1),
			expression.NewLiteral("a"),
		)
		_, err := in.Evaluate(expr)
		assertError(t, err)
	})

	t.Run("Error when adding incompatible types", func(t *testing.T) {
		exprBool := expression.NewBinary(
			expression.NewLiteral(true),
			*token.NewToken(token.PLUS, "+", nil, 1),
			expression.NewLiteral(false),
		)
		_, err := in.Evaluate(exprBool)
		assertError(t, err)

		exprNil := expression.NewBinary(
			expression.NewLiteral(nil),
			*token.NewToken(token.PLUS, "+", nil, 1),
			expression.NewLiteral(nil),
		)
		_, err = in.Evaluate(exprNil)
		assertError(t, err)
	})
}

func TestEvaluateBinaryMinus(t *testing.T) {
	in := setUpInterpreter(t)

	t.Run("Subtraction of two positive numbers", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(50.0),
			*token.NewToken(token.MINUS, "-", nil, 1),
			expression.NewLiteral(40.0),
		)
		val, err := in.Evaluate(expr)
		assertNoError(t, err)
		assertValue(t, val, 10.0)
	})

	t.Run("Subtraction of negative numbers", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(-1.0),
			*token.NewToken(token.MINUS, "-", nil, 1),
			expression.NewLiteral(-5.0),
		)
		val, err := in.Evaluate(expr)
		assertNoError(t, err)
		assertValue(t, val, 4.0)
	})

	t.Run("Subtraction with negative result", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(5.0),
			*token.NewToken(token.MINUS, "-", nil, 1),
			expression.NewLiteral(10.0),
		)
		val, err := in.Evaluate(expr)
		assertNoError(t, err)
		assertValue(t, val, -5.0)
	})

	t.Run("Error when left operand is not a number", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral("a"),
			*token.NewToken(token.MINUS, "-", nil, 1),
			expression.NewLiteral(10.0),
		)
		_, err := in.Evaluate(expr)
		assertError(t, err)
	})

	t.Run("Error when right operand is not a number", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(10.0),
			*token.NewToken(token.MINUS, "-", nil, 1),
			expression.NewLiteral("a"),
		)
		_, err := in.Evaluate(expr)
		assertError(t, err)
	})

	t.Run("Error when operands are booleans", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(true),
			*token.NewToken(token.MINUS, "-", nil, 1),
			expression.NewLiteral(false),
		)
		_, err := in.Evaluate(expr)
		assertError(t, err)
	})

	t.Run("Error when operands are nil", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(nil),
			*token.NewToken(token.MINUS, "-", nil, 1),
			expression.NewLiteral(nil),
		)
		_, err := in.Evaluate(expr)
		assertError(t, err)
	})
}

func TestEvaluateBinaryStar(t *testing.T) {
	in := setUpInterpreter(t)

	t.Run("Multiplication of two positive numbers", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(5.0),
			*token.NewToken(token.STAR, "*", nil, 1),
			expression.NewLiteral(4.0),
		)
		val, err := in.Evaluate(expr)
		assertNoError(t, err)
		assertValue(t, val, 20.0)
	})

	t.Run("Multiplication of negative numbers", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(-2.0),
			*token.NewToken(token.STAR, "*", nil, 1),
			expression.NewLiteral(-5.0),
		)
		val, err := in.Evaluate(expr)
		assertNoError(t, err)
		assertValue(t, val, 10.0)
	})

	t.Run("Multiplication with negative result", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(5.0),
			*token.NewToken(token.STAR, "*", nil, 1),
			expression.NewLiteral(-10.0),
		)
		val, err := in.Evaluate(expr)
		assertNoError(t, err)
		assertValue(t, val, -50.0)
	})

	t.Run("Error when left operand is not a number", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral("a"),
			*token.NewToken(token.STAR, "*", nil, 1),
			expression.NewLiteral(10.0),
		)
		_, err := in.Evaluate(expr)
		assertError(t, err)
	})

	t.Run("Error when right operand is not a number", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(10.0),
			*token.NewToken(token.STAR, "*", nil, 1),
			expression.NewLiteral("a"),
		)
		_, err := in.Evaluate(expr)
		assertError(t, err)
	})

	t.Run("Error when operands are booleans", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(true),
			*token.NewToken(token.STAR, "*", nil, 1),
			expression.NewLiteral(false),
		)
		_, err := in.Evaluate(expr)
		assertError(t, err)
	})

	t.Run("Error when operands are nil", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(nil),
			*token.NewToken(token.STAR, "*", nil, 1),
			expression.NewLiteral(nil),
		)
		_, err := in.Evaluate(expr)
		assertError(t, err)
	})
}

func TestEvaluateBinarySlash(t *testing.T) {
	in := setUpInterpreter(t)

	t.Run("Division of two positive numbers", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(20.0),
			*token.NewToken(token.SLASH, "/", nil, 1),
			expression.NewLiteral(4.0),
		)
		val, err := in.Evaluate(expr)
		assertNoError(t, err)
		assertValue(t, val, 5.0)
	})

	t.Run("Division of negative numbers", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(-10.0),
			*token.NewToken(token.SLASH, "/", nil, 1),
			expression.NewLiteral(-2.0),
		)
		val, err := in.Evaluate(expr)
		assertNoError(t, err)
		assertValue(t, val, 5.0)
	})

	t.Run("Division with negative result", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(10.0),
			*token.NewToken(token.SLASH, "/", nil, 1),
			expression.NewLiteral(-2.0),
		)
		val, err := in.Evaluate(expr)
		assertNoError(t, err)
		assertValue(t, val, -5.0)
	})

	t.Run("Division with fractional result", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(7.0),
			*token.NewToken(token.SLASH, "/", nil, 1),
			expression.NewLiteral(2.0),
		)
		val, err := in.Evaluate(expr)
		assertNoError(t, err)
		assertValue(t, val, 3.5)
	})

	t.Run("Error when left operand is not a number", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral("a"),
			*token.NewToken(token.SLASH, "/", nil, 1),
			expression.NewLiteral(10.0),
		)
		_, err := in.Evaluate(expr)
		assertError(t, err)
	})

	t.Run("Error when right operand is not a number", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(10.0),
			*token.NewToken(token.SLASH, "/", nil, 1),
			expression.NewLiteral("a"),
		)
		_, err := in.Evaluate(expr)
		assertError(t, err)
	})

	t.Run("A positive number divided by zero produces positive infinity", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(10.0),
			*token.NewToken(token.SLASH, "/", nil, 1),
			expression.NewLiteral(0.0),
		)
		val, err := in.Evaluate(expr)
		assertNoError(t, err)
		assertValue(t, val, math.Inf(1))
	})

	t.Run("A negative number divided by zero produces negative infinity", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(-10.0),
			*token.NewToken(token.SLASH, "/", nil, 1),
			expression.NewLiteral(0.0),
		)
		val, err := in.Evaluate(expr)
		assertNoError(t, err)
		assertValue(t, val, math.Inf(-1))
	})

	t.Run("Zero divided by zero produces NaN", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(0.0),
			*token.NewToken(token.SLASH, "/", nil, 1),
			expression.NewLiteral(0.0),
		)
		val, err := in.Evaluate(expr)
		assertNoError(t, err)

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
		assertError(t, err)
	})

	t.Run("Error when operands are nil", func(t *testing.T) {
		expr := expression.NewBinary(
			expression.NewLiteral(nil),
			*token.NewToken(token.SLASH, "/", nil, 1),
			expression.NewLiteral(nil),
		)
		_, err := in.Evaluate(expr)
		assertError(t, err)
	})
}
