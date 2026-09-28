package expression

import (
	"fmt"
	"strings"
	"tp-lox/internal/token"
)

type Expr interface {
	Expression()
}

type Literal struct {
	Value any
}

func NewLiteral(value any) *Literal {
	return &Literal{
		Value: value,
	}
}

func (l *Literal) Expression() {}

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

type Grouping struct {
	Expr Expr
}

func NewGrouping(expression Expr) *Grouping {
	return &Grouping{
		Expr: expression,
	}
}

func (g *Grouping) Expression() {}

type Variable struct {
	Name token.Token
}

func NewVariable(name token.Token) *Variable {
	return &Variable{
		Name: name,
	}
}

func (v *Variable) Expression() {}

type Assign struct {
	Name  token.Token
	Value Expr
}

func NewAssign(name token.Token, value Expr) *Assign {
	return &Assign{
		Name:  name,
		Value: value,
	}
}

func (a *Assign) Expression() {}

type Logical struct {
	Left     Expr
	Operator token.Token
	Right    Expr
}

func NewLogical(left Expr, operator token.Token, right Expr) *Logical {
	return &Logical{
		Left:     left,
		Operator: operator,
		Right:    right,
	}
}

func (l *Logical) Expression() {}

type Call struct {
	Callee    Expr
	Paren     token.Token
	Arguments []Expr
}

func NewCall(callee Expr, paren token.Token, arguments []Expr) *Call {
	return &Call{
		Callee:    callee,
		Paren:     paren,
		Arguments: arguments,
	}
}

func (c *Call) Expression() {}

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
	case *Variable:
		return fmt.Sprintf("<%s>", e.Name.Lexeme)
	case *Assign:
		return fmt.Sprintf("(= %s %s)", e.Name.Lexeme, String(e.Value))
	case *Logical:
		return fmt.Sprintf("(%s %s %s)", String(e.Left), e.Operator.TokenType, String(e.Right))
	case *Call:
		var args []string
		for _, arg := range e.Arguments {
			args = append(args, String(arg))
		}
		if len(args) == 0 {
			return fmt.Sprintf("(call %s)", String(e.Callee))
		}
		return fmt.Sprintf("(call %s %s)", String(e.Callee), strings.Join(args, " "))
	default:
		return ""
	}
}
