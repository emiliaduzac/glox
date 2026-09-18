package expression

import "tp-lox/internal/token"

type Expression interface {
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
	right    Expression
}

func NewUnary(operator token.Token, right Expression) *Unary {
	return &Unary{
		operator: operator,
		right:    right,
	}
}

func (u *Unary) Expression() {}

// Binary
type Binary struct {
	left     Expression
	operator token.Token
	right    Expression
}

func NewBinary(left Expression, operator token.Token, right Expression) *Binary {
	return &Binary{
		left:     left,
		operator: operator,
		right:    right,
	}
}

func (b *Binary) Expression() {}

// Grouping
type Grouping struct {
	expression Expression
}

func NewGrouping(expression Expression) *Grouping {
	return &Grouping{
		expression: expression,
	}
}

func (g *Grouping) Expression() {}
