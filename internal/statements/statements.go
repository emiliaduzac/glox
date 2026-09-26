package statements

import (
	"fmt"
	"tp-lox/internal/expression"
	"tp-lox/internal/token"
)

type Stmt interface {
	Statement()
}

type ExpressionStmt struct {
	Expr expression.Expr
}

func NewExpressionStmt(expr expression.Expr) *ExpressionStmt {
	return &ExpressionStmt{
		Expr: expr,
	}
}

func (e *ExpressionStmt) Statement() {}

type PrintStmt struct {
	Expr expression.Expr
}

func NewPrintStmt(expr expression.Expr) *PrintStmt {
	return &PrintStmt{
		Expr: expr,
	}
}

func (p *PrintStmt) Statement() {}

type VarStmt struct {
	Name        token.Token
	Initializer expression.Expr
}

func NewVarStmt(name token.Token, expr expression.Expr) *VarStmt {
	return &VarStmt{
		Name:        name,
		Initializer: expr,
	}
}

func (p *VarStmt) Statement() {}

func String(s Stmt) string {
	switch s := s.(type) {
	case *PrintStmt:
		return fmt.Sprintf("(print %s)", expression.String(s.Expr))
	case *ExpressionStmt:
		return fmt.Sprintf("(expr %s)", expression.String(s.Expr))
	case *VarStmt:
		if s.Initializer == nil {
			return fmt.Sprintf("(var %s)", s.Name.String())
		}
		return fmt.Sprintf("(var %s %s)", s.Name.String(), expression.String(s.Initializer))
	}
	return fmt.Sprintf("unknown statement: %T", s)
}
