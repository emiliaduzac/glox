package expression

import "tp-lox/internal/token"

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
	operator token.Token
	right    Expr
}

func NewUnary(operator token.Token, right Expr) *Unary {
	return &Unary{
		operator: operator,
		right:    right,
	}
}

func (u *Unary) Expression() {}

// Binary
type Binary struct {
	left     Expr
	operator token.Token
	right    Expr
}

func NewBinary(left Expr, operator token.Token, right Expr) *Binary {
	return &Binary{
		left:     left,
		operator: operator,
		right:    right,
	}
}

func (b *Binary) Expression() {}

// Grouping
type Grouping struct {
	expression Expr
}

func NewGrouping(expression Expr) *Grouping {
	return &Grouping{
		expression: expression,
	}
}

func (g *Grouping) Expression() {}
