package interpreter

import (
	"fmt"
	"tp-lox/internal/errors"
	"tp-lox/internal/expression"
	"tp-lox/internal/token"
)

type Interpreter struct{}

func NewInterpreter() *Interpreter {
	return &Interpreter{}
}

func (in *Interpreter) Evaluate(expr expression.Expr) (any, error) {
	switch e := expr.(type) {
	case *expression.Literal:
		return e.Value, nil
	case *expression.Grouping:
		return in.Evaluate(e.Expr)
	case *expression.Unary:
		return in.evaluateUnary(e)
	case *expression.Binary:
		return in.evaluateBinary(e)
	default:
		return nil, nil
	}
}

func (in *Interpreter) evaluateUnary(expr *expression.Unary) (any, error) {
	right, err := in.Evaluate(expr.Right)
	if err != nil {
		return nil, err
	}

	switch expr.Operator.TokenType {
	case token.MINUS:
		num, ok := right.(float64)
		if !ok {
			return nil, errors.NewLoxError(expr.Operator.Line, "operand must be a number")
		}
		return -num, nil
	case token.BANG:
		return !isTruthy(right), nil
	default:
		return nil, errors.NewLoxError(expr.Operator.Line, fmt.Sprintf("unknown unary operator: %s", expr.Operator.Lexeme))
	}
}

func (in *Interpreter) evaluateBinary(expr *expression.Binary) (any, error) {
	left, err := in.Evaluate(expr.Left)
	if err != nil {
		return nil, err
	}

	right, err := in.Evaluate(expr.Right)
	if err != nil {
		return nil, err
	}

	switch expr.Operator.TokenType {
	case token.PLUS:
		if leftNum, okLeft := left.(float64); okLeft {
			if rightNum, okRight := right.(float64); okRight {
				return leftNum + rightNum, nil
			}
		}
		if leftStr, okLeft := left.(string); okLeft {
			if rightStr, okRight := right.(string); okRight {
				return leftStr + rightStr, nil
			}
		}
		return nil, errors.NewLoxError(expr.Operator.Line, "operands must be two numbers or two strings")
	case token.MINUS:
		leftNum, okLeft := left.(float64)
		rightNum, okRight := right.(float64)
		if !okLeft || !okRight {
			return nil, errors.NewLoxError(expr.Operator.Line, "operands must be numbers")
		}
		return leftNum - rightNum, nil
	case token.STAR:
		leftNum, okLeft := left.(float64)
		rightNum, okRight := right.(float64)
		if !okLeft || !okRight {
			return nil, errors.NewLoxError(expr.Operator.Line, "operands must be numbers")
		}
		return leftNum * rightNum, nil
	case token.SLASH:
		leftNum, okLeft := left.(float64)
		rightNum, okRight := right.(float64)
		if !okLeft || !okRight {
			return nil, errors.NewLoxError(expr.Operator.Line, "operands must be numbers")
		}
		return leftNum / rightNum, nil
	default:
		return nil, errors.NewLoxError(expr.Operator.Line, fmt.Sprintf("unknown binary operator: %s", expr.Operator.Lexeme))
	}
}

func isTruthy(value any) bool {
	if value == nil {
		return false
	}
	if b, ok := value.(bool); ok {
		return b
	}
	return true
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
