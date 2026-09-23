package interpreter

import (
	"testing"
	"tp-lox/internal/expression"
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
