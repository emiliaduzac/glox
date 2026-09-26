package environment

import (
	"testing"
	"tp-lox/internal/token"
)

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
		AssertNoError(t, err)
		AssertValue(t, valueNumber, 19.0)

		valueString, err := env.Get(tokenString)
		AssertNoError(t, err)
		AssertValue(t, valueString, "Hello")

		valueBool, err := env.Get(tokenBool)
		AssertNoError(t, err)
		AssertValue(t, valueBool, true)

		valueNil, err := env.Get(tokenNil)
		AssertNoError(t, err)
		AssertValue(t, valueNil, nil)
	})

	t.Run("Redefine variable", func(t *testing.T) {
		env := SetUpEnvironment(t)
		env.Define("number", 19.0)

		tokenNumber := *token.NewToken(token.IDENTIFIER, "number", nil, 1)

		valueNumber, err := env.Get(tokenNumber)
		AssertNoError(t, err)
		AssertValue(t, valueNumber, 19.0)

		env.Define("number", 99.0)

		valueNumber, err = env.Get(tokenNumber)
		AssertNoError(t, err)
		AssertValue(t, valueNumber, 99.0)
	})

	t.Run("Redifine variable with different type", func(t *testing.T) {
		env := SetUpEnvironment(t)
		tokenA := *token.NewToken(token.IDENTIFIER, "a", nil, 1)

		env.Define("a", 19.0)
		valueA, err := env.Get(tokenA)

		AssertNoError(t, err)
		AssertValue(t, valueA, 19.0)

		env.Define("a", "Hello")
		valueA, err = env.Get(tokenA)

		AssertNoError(t, err)
		AssertValue(t, valueA, "Hello")
	})

	t.Run("Get undefined variable", func(t *testing.T) {
		env := SetUpEnvironment(t)

		tokenUndefined := *token.NewToken(token.IDENTIFIER, "undefined", nil, 1)

		_, err := env.Get(tokenUndefined)
		AssertErrorMessage(t, err, "undefined variable 'undefined'")
	})
}
