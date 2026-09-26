package environment

import (
	"fmt"
	"tp-lox/internal/errors"
	"tp-lox/internal/token"
)

type Environment struct {
	values map[string]any
}

func NewEnvironment() *Environment {
	return &Environment{
		values: make(map[string]any),
	}
}

func (e *Environment) Define(name string, value any) {
	e.values[name] = value
}

func (e *Environment) Get(name token.Token) (any, error) {
	if value, ok := e.values[name.Lexeme]; ok {
		return value, nil
	}
	return nil, errors.NewLoxError(name.Line, fmt.Sprintf("undefined variable '%s'", name.Lexeme))
}
