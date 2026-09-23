package interpreter

import (
	"fmt"
	"tp-lox/internal/expression"
)

type Interpreter struct{}

func NewInterpreter() *Interpreter {
	return &Interpreter{}
}

func (in *Interpreter) Evaluate(expr expression.Expr) (any, error) {
	switch e := expr.(type) {
	case *expression.Literal:
		return e.Value, nil
	default:
		return nil, nil
	}
}

func Stringify(value any) string {
	if value == nil {
		return "nil"
	}
	return fmt.Sprintf("%v", value)
}

func (in *Interpreter) Interpret(expr expression.Expr) error {
	val, err := in.Evaluate(expr)
	if err != nil {
		return err
	}
	fmt.Println(Stringify(val))
	return nil
}
