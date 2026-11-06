package environment

import (
	"testing"
	"tp-lox/internal/token"
	"tp-lox/internal/utils"
)

func SetUpEnvironment(t *testing.T) *Environment {
	t.Helper()
	return NewEnvironment()
}

func TestEnvironment(t *testing.T) {
	t.Run("Define and get diferent types", func(t *testing.T) {
		env := SetUpEnvironment(t)
		env.Define("number", 19.0)
		env.Define("string", "Hello")
		env.Define("bool", true)
		env.Define("nilVar", nil)

		tokenNumber := *token.NewToken(token.IDENTIFIER, "number", nil, 1)
		tokenString := *token.NewToken(token.IDENTIFIER, "string", nil, 1)
		tokenBool := *token.NewToken(token.IDENTIFIER, "bool", nil, 1)
		tokenNil := *token.NewToken(token.IDENTIFIER, "nilVar", nil, 1)

		valueNumber, err := env.Get(tokenNumber)
		utils.AssertNoError(t, err)
		utils.AssertValue(t, valueNumber, 19.0)

		valueString, err := env.Get(tokenString)
		utils.AssertNoError(t, err)
		utils.AssertValue(t, valueString, "Hello")

		valueBool, err := env.Get(tokenBool)
		utils.AssertNoError(t, err)
		utils.AssertValue(t, valueBool, true)

		valueNil, err := env.Get(tokenNil)
		utils.AssertNoError(t, err)
		utils.AssertValue(t, valueNil, nil)
	})

	t.Run("Redefine variable", func(t *testing.T) {
		env := SetUpEnvironment(t)
		env.Define("number", 19.0)

		tokenNumber := *token.NewToken(token.IDENTIFIER, "number", nil, 1)

		valueNumber, err := env.Get(tokenNumber)
		utils.AssertNoError(t, err)
		utils.AssertValue(t, valueNumber, 19.0)

		env.Define("number", 99.0)

		valueNumber, err = env.Get(tokenNumber)
		utils.AssertNoError(t, err)
		utils.AssertValue(t, valueNumber, 99.0)
	})

	t.Run("Redifine variable with different type", func(t *testing.T) {
		env := SetUpEnvironment(t)
		tokenA := *token.NewToken(token.IDENTIFIER, "a", nil, 1)

		env.Define("a", 19.0)
		valueA, err := env.Get(tokenA)

		utils.AssertNoError(t, err)
		utils.AssertValue(t, valueA, 19.0)

		env.Define("a", "Hello")
		valueA, err = env.Get(tokenA)

		utils.AssertNoError(t, err)
		utils.AssertValue(t, valueA, "Hello")
	})

	t.Run("Get undefined variable", func(t *testing.T) {
		env := SetUpEnvironment(t)

		tokenUndefined := *token.NewToken(token.IDENTIFIER, "undefined", nil, 1)

		_, err := env.Get(tokenUndefined)
		utils.AssertErrorMessage(t, err, "undefined variable 'undefined'")
	})
}

func TestEnvironmentAssign(t *testing.T) {
	t.Run("Assign to existing variable uptdates value", func(t *testing.T) {
		env := SetUpEnvironment(t)
		env.Define("number", 19.0)

		tokenNumber := *token.NewToken(token.IDENTIFIER, "number", nil, 1)
		err := env.Assign(tokenNumber, 99.0)
		utils.AssertNoError(t, err)

		valueNumber, err := env.Get(tokenNumber)
		utils.AssertNoError(t, err)
		utils.AssertValue(t, valueNumber, 99.0)
	})

	t.Run("Assign different type to existing variable updates value", func(t *testing.T) {
		env := SetUpEnvironment(t)
		env.Define("var", 19.0)

		tokenVar := *token.NewToken(token.IDENTIFIER, "var", nil, 1)
		err := env.Assign(tokenVar, "Hello")
		utils.AssertNoError(t, err)

		valueVar, err := env.Get(tokenVar)
		utils.AssertNoError(t, err)
		utils.AssertValue(t, valueVar, "Hello")
	})

	t.Run("Assign to undefined variable returns error", func(t *testing.T) {
		env := SetUpEnvironment(t)
		tokenUndefined := *token.NewToken(token.IDENTIFIER, "undefined", nil, 1)

		err := env.Assign(tokenUndefined, 99.0)
		utils.AssertErrorMessage(t, err, "undefined variable 'undefined'")
	})
}

func TestEnvironmentScopes(t *testing.T) {

	t.Run("Define new variable in enclosed scope", func(t *testing.T) {
		global := SetUpEnvironment(t)
		child := NewEnclosedEnvironment(global)
		child.Define("localVar", 42.0)

		tok := *token.NewToken(token.IDENTIFIER, "localVar", nil, 1)
		val, err := child.Get(tok)
		utils.AssertNoError(t, err)
		utils.AssertValue(t, val, 42.0)
	})

	t.Run("Define redefines variable in same enclosed scope", func(t *testing.T) {
		global := SetUpEnvironment(t)
		child := NewEnclosedEnvironment(global)
		child.Define("x", 1.0)
		child.Define("x", 2.0)

		tok := *token.NewToken(token.IDENTIFIER, "x", nil, 1)
		val, err := child.Get(tok)
		utils.AssertNoError(t, err)
		utils.AssertValue(t, val, 2.0)
	})

	t.Run("Define shadows variable from enclosing scope without modifying it", func(t *testing.T) {
		global := SetUpEnvironment(t)
		global.Define("a", "global_a")

		child := NewEnclosedEnvironment(global)
		child.Define("a", "shadowed_a")

		tokA := *token.NewToken(token.IDENTIFIER, "a", nil, 1)
		childVal, err := child.Get(tokA)
		utils.AssertNoError(t, err)
		utils.AssertValue(t, childVal, "shadowed_a")

		globalVal, err := global.Get(tokA)
		utils.AssertNoError(t, err)
		utils.AssertValue(t, globalVal, "global_a")
	})

	t.Run("Get returns local variable when present in current scope", func(t *testing.T) {
		global := SetUpEnvironment(t)
		global.Define("x", "global")
		child := NewEnclosedEnvironment(global)
		child.Define("y", "local")

		tokY := *token.NewToken(token.IDENTIFIER, "y", nil, 1)
		val, err := child.Get(tokY)
		utils.AssertNoError(t, err)
		utils.AssertValue(t, val, "local")
	})

	t.Run("Get delegates to direct parent environment", func(t *testing.T) {
		global := SetUpEnvironment(t)
		global.Define("parentVar", 10.0)
		child := NewEnclosedEnvironment(global)

		tok := *token.NewToken(token.IDENTIFIER, "parentVar", nil, 1)
		val, err := child.Get(tok)
		utils.AssertNoError(t, err)
		utils.AssertValue(t, val, 10.0)
	})

	t.Run("Get delegates multiple levels up to grandparent environment", func(t *testing.T) {
		global := SetUpEnvironment(t)
		global.Define("rootVar", "from_root")

		middle := NewEnclosedEnvironment(global)
		inner := NewEnclosedEnvironment(middle)

		tok := *token.NewToken(token.IDENTIFIER, "rootVar", nil, 1)
		val, err := inner.Get(tok)
		utils.AssertNoError(t, err)
		utils.AssertValue(t, val, "from_root")
	})

	t.Run("Get returns error when variable is undefined across all enclosing scopes", func(t *testing.T) {
		global := SetUpEnvironment(t)
		child := NewEnclosedEnvironment(global)

		tok := *token.NewToken(token.IDENTIFIER, "unknown", nil, 1)
		_, err := child.Get(tok)
		utils.AssertErrorMessage(t, err, "undefined variable 'unknown'")
	})

	t.Run("Assign updates variable in local scope", func(t *testing.T) {
		global := SetUpEnvironment(t)
		child := NewEnclosedEnvironment(global)
		child.Define("local", 1.0)

		tok := *token.NewToken(token.IDENTIFIER, "local", nil, 1)
		err := child.Assign(tok, 2.0)
		utils.AssertNoError(t, err)

		val, err := child.Get(tok)
		utils.AssertNoError(t, err)
		utils.AssertValue(t, val, 2.0)
	})

	t.Run("Assign updates shadowed variable in child and preserves parent", func(t *testing.T) {
		global := SetUpEnvironment(t)
		global.Define("x", "outer")

		child := NewEnclosedEnvironment(global)
		child.Define("x", "inner")

		tokX := *token.NewToken(token.IDENTIFIER, "x", nil, 1)
		err := child.Assign(tokX, "mutated_inner")
		utils.AssertNoError(t, err)

		childVal, err := child.Get(tokX)
		utils.AssertNoError(t, err)
		utils.AssertValue(t, childVal, "mutated_inner")

		globalVal, err := global.Get(tokX)
		utils.AssertNoError(t, err)
		utils.AssertValue(t, globalVal, "outer")
	})

	t.Run("Assign delegates to direct parent environment to mutate outer variable", func(t *testing.T) {
		global := SetUpEnvironment(t)
		global.Define("counter", 10.0)
		child := NewEnclosedEnvironment(global)

		tok := *token.NewToken(token.IDENTIFIER, "counter", nil, 1)
		err := child.Assign(tok, 20.0)
		utils.AssertNoError(t, err)

		val, err := global.Get(tok)
		utils.AssertNoError(t, err)
		utils.AssertValue(t, val, 20.0)
	})

	t.Run("Assign delegates multiple levels up to mutate grandparent variable", func(t *testing.T) {
		global := SetUpEnvironment(t)
		global.Define("deepCounter", 100.0)

		middle := NewEnclosedEnvironment(global)
		inner := NewEnclosedEnvironment(middle)

		tok := *token.NewToken(token.IDENTIFIER, "deepCounter", nil, 1)
		err := inner.Assign(tok, 200.0)
		utils.AssertNoError(t, err)

		val, err := global.Get(tok)
		utils.AssertNoError(t, err)
		utils.AssertValue(t, val, 200.0)
	})

	t.Run("Assign returns error when variable is undefined across all enclosing scopes", func(t *testing.T) {
		global := SetUpEnvironment(t)
		child := NewEnclosedEnvironment(global)

		tok := *token.NewToken(token.IDENTIFIER, "not_declared", nil, 1)
		err := child.Assign(tok, 99.0)
		utils.AssertErrorMessage(t, err, "undefined variable 'not_declared'")
	})
}
