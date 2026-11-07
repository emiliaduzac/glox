package utils

import (
	"testing"
)

func TestIsEqual(t *testing.T) {
	t.Run("Nil equals nil", func(t *testing.T) {
		AssertValue(t, IsEqual(nil, nil), true)
	})

	t.Run("Nil does not equal non-nil", func(t *testing.T) {
		AssertValue(t, IsEqual(nil, false), false)
		AssertValue(t, IsEqual(nil, 0.0), false)
		AssertValue(t, IsEqual(nil, ""), false)
		AssertValue(t, IsEqual(false, nil), false)
		AssertValue(t, IsEqual(0.0, nil), false)
		AssertValue(t, IsEqual("", nil), false)
	})

	t.Run("Same numbers are equal", func(t *testing.T) {
		AssertValue(t, IsEqual(11.0, 11.0), true)
		AssertValue(t, IsEqual(-10.0, -10.0), true)
		AssertValue(t, IsEqual(0.0, 0.0), true)
	})

	t.Run("Different numbers are not equal", func(t *testing.T) {
		AssertValue(t, IsEqual(11.0, 10.0), false)
		AssertValue(t, IsEqual(-5.0, 5.0), false)
	})

	t.Run("Same strings are equal", func(t *testing.T) {
		AssertValue(t, IsEqual("hello", "hello"), true)
		AssertValue(t, IsEqual("", ""), true)
	})

	t.Run("Different strings are not equal", func(t *testing.T) {
		AssertValue(t, IsEqual("hello", "world"), false)
	})

	t.Run("Same booleans are equal", func(t *testing.T) {
		AssertValue(t, IsEqual(true, true), true)
		AssertValue(t, IsEqual(false, false), true)
	})

	t.Run("Different booleans are not equal", func(t *testing.T) {
		AssertValue(t, IsEqual(true, false), false)
	})

	t.Run("Different types are not equal without coercion", func(t *testing.T) {
		AssertValue(t, IsEqual(42.0, "42.0"), false)
		AssertValue(t, IsEqual(0.0, false), false)
		AssertValue(t, IsEqual("", false), false)
		AssertValue(t, IsEqual(true, 1.0), false)
	})
}

func TestIsTruthy(t *testing.T) {
	t.Run("Nil is falsy", func(t *testing.T) {
		AssertValue(t, IsTruthy(nil), false)
	})

	t.Run("False is falsy", func(t *testing.T) {
		AssertValue(t, IsTruthy(false), false)
	})

	t.Run("True is truthy", func(t *testing.T) {
		AssertValue(t, IsTruthy(true), true)
	})

	t.Run("Zero number is truthy", func(t *testing.T) {
		AssertValue(t, IsTruthy(0.0), true)
	})

	t.Run("Positive number is truthy", func(t *testing.T) {
		AssertValue(t, IsTruthy(88.0), true)
	})

	t.Run("Empty string is truthy", func(t *testing.T) {
		AssertValue(t, IsTruthy(""), true)
	})

	t.Run("Non-empty string is truthy", func(t *testing.T) {
		AssertValue(t, IsTruthy("hello"), true)
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

func TestCheckNumberOperand(t *testing.T) {
	t.Run("Operand is a number", func(t *testing.T) {
		num, err := CheckNumberOperand(1, 10.0)
		AssertNoError(t, err)
		AssertValue(t, num, 10.0)
	})

	t.Run("Operand is not a number", func(t *testing.T) {
		_, err := CheckNumberOperand(1, "not a number")
		AssertErrorMessage(t, err, "operand must be a number")
	})
}

func TestCheckNumberOperands(t *testing.T) {
	t.Run("Both operands are numbers", func(t *testing.T) {
		left, right, err := CheckNumberOperands(1, 10.0, 20.0)
		AssertNoError(t, err)
		AssertValue(t, left, 10.0)
		AssertValue(t, right, 20.0)
	})

	t.Run("Left operand is not a number", func(t *testing.T) {
		_, _, err := CheckNumberOperands(1, "not a number", 20.0)
		AssertErrorMessage(t, err, "operands must be numbers")
	})

	t.Run("Right operand is not a number", func(t *testing.T) {
		_, _, err := CheckNumberOperands(1, 10.0, "not a number")
		AssertErrorMessage(t, err, "operands must be numbers")
	})

	t.Run("Neither operand is a number", func(t *testing.T) {
		_, _, err := CheckNumberOperands(1, "a", "b")
		AssertErrorMessage(t, err, "operands must be numbers")
	})
}
