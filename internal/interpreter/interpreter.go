package interpreter

import (
	"fmt"
	"math"
	"tp-lox/internal/errors"
	"tp-lox/internal/expression"
	"tp-lox/internal/statements"
	"tp-lox/internal/token"
)

type Interpreter struct{}

func NewInterpreter() *Interpreter {
	return &Interpreter{}
}

func (in *Interpreter) Interpret([]statements.Stmt) error {
	for _, stmt := range []statements.Stmt{} {
		val, err := in.Execute(stmt)
		if err != nil {
			return err
		}
		fmt.Println(Stringify(val))
	}

	return nil
}

// execute ejecuta una sentencia y devuelve su valor.
func (in *Interpreter) Execute(stmt statements.Stmt) (any, error) {
	switch s := stmt.(type) {
	case *statements.PrintStmt:
		val, err := in.Evaluate(s.Expr)
		if err != nil {
			return nil, err
		}
		fmt.Println(Stringify(val))
		return val, nil

	case *statements.ExpressionStmt:
		val, err := in.Evaluate(s.Expr)
		if err != nil {
			return nil, err
		}
		return val, nil

	default:
		return nil, errors.NewLoxError(0, fmt.Sprintf("unknown statement: %T", stmt))
	}
}

// Evaluate evalúa una expresión y devuelve su valor.
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
		return nil, errors.NewLoxError(0, fmt.Sprintf("unknown expression: %T", expr))
	}
}

func (in *Interpreter) evaluateUnary(expr *expression.Unary) (any, error) {
	right, err := in.Evaluate(expr.Right)
	if err != nil {
		return nil, err
	}

	switch expr.Operator.TokenType {
	case token.MINUS:
		num, err := checkNumberOperand(expr.Operator, right)
		if err != nil {
			return nil, err
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
		leftNum, rightNum, err := checkNumberOperands(expr.Operator, left, right)
		if err != nil {
			return nil, err
		}
		return leftNum - rightNum, nil
	case token.STAR:
		leftNum, rightNum, err := checkNumberOperands(expr.Operator, left, right)
		if err != nil {
			return nil, err
		}
		return leftNum * rightNum, nil
	case token.SLASH:
		leftNum, rightNum, err := checkNumberOperands(expr.Operator, left, right)
		if err != nil {
			return nil, err
		}
		return leftNum / rightNum, nil
	case token.PERCENT:
		leftNum, rightNum, err := checkNumberOperands(expr.Operator, left, right)
		if err != nil {
			return nil, err
		}
		return math.Mod(leftNum, rightNum), nil
	case token.GREATER:
		leftNum, rightNum, err := checkNumberOperands(expr.Operator, left, right)
		if err != nil {
			return nil, err
		}
		return leftNum > rightNum, nil
	case token.GREATER_EQUAL:
		leftNum, rightNum, err := checkNumberOperands(expr.Operator, left, right)
		if err != nil {
			return nil, err
		}
		return leftNum >= rightNum, nil
	case token.LESS:
		leftNum, rightNum, err := checkNumberOperands(expr.Operator, left, right)
		if err != nil {
			return nil, err
		}
		return leftNum < rightNum, nil
	case token.LESS_EQUAL:
		leftNum, rightNum, err := checkNumberOperands(expr.Operator, left, right)
		if err != nil {
			return nil, err
		}
		return leftNum <= rightNum, nil
	case token.EQUAL_EQUAL:
		return isEqual(left, right), nil
	case token.BANG_EQUAL:
		return !isEqual(left, right), nil
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

func isEqual(a, b any) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil {
		return false
	}
	return a == b
}

func checkNumberOperand(operator token.Token, operand any) (float64, error) {
	num, ok := operand.(float64)
	if !ok {
		return 0, errors.NewLoxError(operator.Line, "operand must be a number")
	}
	return num, nil
}

func checkNumberOperands(operator token.Token, left, right any) (float64, float64, error) {
	leftNum, okLeft := left.(float64)
	rightNum, okRight := right.(float64)
	if !okLeft || !okRight {
		return 0, 0, errors.NewLoxError(operator.Line, "operands must be numbers")
	}
	return leftNum, rightNum, nil
}
