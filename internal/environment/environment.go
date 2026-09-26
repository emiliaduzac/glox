package environment

import (
	"fmt"
	"tp-lox/internal/errors"
	"tp-lox/internal/token"
)

type Environment struct {
	values    map[string]any
	enclosing *Environment
}

func NewEnvironment() *Environment {
	return &Environment{
		values:    make(map[string]any),
		enclosing: nil,
	}
}

func NewEnclosedEnvironment(enclosing *Environment) *Environment {
	return &Environment{
		values:    make(map[string]any),
		enclosing: enclosing,
	}
}

func (e *Environment) Define(name string, value any) {
	e.values[name] = value
}

func (e *Environment) Get(name token.Token) (any, error) {
	if value, ok := e.values[name.Lexeme]; ok {
		return value, nil
	}
	if e.enclosing != nil {
		return e.enclosing.Get(name)
	}
	return nil, errors.NewLoxError(name.Line, fmt.Sprintf("undefined variable '%s'", name.Lexeme))
}

func (e *Environment) Assign(name token.Token, value any) error {
	if _, ok := e.values[name.Lexeme]; ok {
		e.values[name.Lexeme] = value
		return nil
	}
	if e.enclosing != nil {
		return e.enclosing.Assign(name, value)
	}
	return errors.NewLoxError(name.Line, fmt.Sprintf("undefined variable '%s'", name.Lexeme))
}
