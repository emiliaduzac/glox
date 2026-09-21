package expression

import (
	"fmt"
	"tp-lox/internal/token"
)

type Expr interface {
	Expression()
}

// Literal
type Literal struct {
	Value any
}

func NewLiteral(value any) *Literal {
	return &Literal{
		Value: value,
	}
}

func (l *Literal) Expression() {}

// Unary
type Unary struct {
	Operator token.Token
	Right    Expr
}

func NewUnary(operator token.Token, right Expr) *Unary {
	return &Unary{
		Operator: operator,
		Right:    right,
	}
}

func (u *Unary) Expression() {}

// Binary
type Binary struct {
	Left     Expr
	Operator token.Token
	Right    Expr
}

func NewBinary(left Expr, operator token.Token, right Expr) *Binary {
	return &Binary{
		Left:     left,
		Operator: operator,
		Right:    right,
	}
}

func (b *Binary) Expression() {}

// Grouping
type Grouping struct {
	Expr Expr
}

func NewGrouping(expression Expr) *Grouping {
	return &Grouping{
		Expr: expression,
	}
}

func (g *Grouping) Expression() {}

func stringLiteral(value any) string {
	switch v := value.(type) {
	case string:
		return fmt.Sprintf("\"%s\"", v)
	case bool:
		if v {
			return "TRUE"
		}
		return "FALSE"
	case nil:
		return "NIL"
	default:
		return fmt.Sprintf("%v", v)
	}
}

func String(expr Expr) string {
	switch e := expr.(type) {
	case *Literal:
		return fmt.Sprintf("<%v>", stringLiteral(e.Value))
	case *Unary:
		return fmt.Sprintf("(%s %s)", e.Operator.TokenType, String(e.Right))
	case *Binary:
		return fmt.Sprintf("(%s %s %s)", String(e.Left), e.Operator.TokenType, String(e.Right))
	case *Grouping:
		return fmt.Sprintf("(group %s)", String(e.Expr))
	default:
		return ""
	}
}
